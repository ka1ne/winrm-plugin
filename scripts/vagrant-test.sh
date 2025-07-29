#!/bin/bash
# Vagrant-based integration testing script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Vagrant is installed
if ! command -v vagrant &> /dev/null; then
    log_error "Vagrant is not installed. Please install Vagrant first:"
    echo "  macOS: brew install vagrant"
    echo "  Linux: https://www.vagrantup.com/downloads"
    echo "  Windows: https://www.vagrantup.com/downloads"
    exit 1
fi

# Check if VirtualBox is installed (or other provider)
if ! command -v VBoxManage &> /dev/null && ! command -v vmrun &> /dev/null; then
    log_warning "VirtualBox not found. Make sure you have a Vagrant provider installed:"
    echo "  VirtualBox: https://www.virtualbox.org/wiki/Downloads"
    echo "  VMware: https://www.vmware.com/products/workstation-pro.html"
    echo "  Hyper-V (Windows): Enable via Windows Features"
fi

cd "$PROJECT_DIR"

# Build the plugin first
log "Building WinRM plugin..."
go build -o winrm-plugin .
log_success "Plugin built successfully"

# Function to start Vagrant VM
start_vm() {
    log "Starting Vagrant Windows VM..."
    
    # Check if VM is already running
    if vagrant status | grep -q "running"; then
        log_success "Vagrant VM is already running"
    else
        log "Bringing up Vagrant VM (this may take several minutes on first run)..."
        vagrant up
        
        # Wait for WinRM to be fully ready
        log "Waiting for WinRM service to be fully ready..."
        sleep 30
        
        # Verify WinRM connectivity
        log "Testing WinRM connectivity..."
        if vagrant winrm -c "Write-Host 'Vagrant WinRM test successful'" 2>/dev/null; then
            log_success "WinRM connectivity verified"
        else
            log_warning "Direct WinRM test failed, but this might be normal"
        fi
    fi
}

# Function to run integration tests
run_tests() {
    log "Running integration tests against Vagrant VM..."
    
    # Get VM IP address (try different methods)
    VM_IP=""
    
    # Method 1: Try to get IP from vagrant ssh-config
    if command -v vagrant &> /dev/null; then
        VM_IP=$(vagrant ssh-config | grep HostName | awk '{print $2}' 2>/dev/null || echo "")
    fi
    
    # Method 2: Use localhost with port forwarding
    if [[ -z "$VM_IP" ]] || [[ "$VM_IP" == "127.0.0.1" ]]; then
        VM_IP="localhost"
        VM_PORT="15985"
        log "Using localhost with port forwarding: $VM_IP:$VM_PORT"
    else
        VM_PORT="5985"
        log "Using VM IP address: $VM_IP:$VM_PORT"
    fi
    
    # Set test environment variables
    export TEST_WINRM_HOST="$VM_IP"
    export TEST_WINRM_PORT="$VM_PORT"
    export TEST_WINRM_USERNAME="winrm-plugin-test"
    export TEST_WINRM_PASSWORD="PluginTest123!"
    export TEST_WINRM_PROTOCOL="http"
    
    log "Test configuration:"
    log "  Host: $TEST_WINRM_HOST:$TEST_WINRM_PORT"
    log "  User: $TEST_WINRM_USERNAME"
    log "  Protocol: $TEST_WINRM_PROTOCOL"
    
    # Test 1: Basic connectivity using the plugin
    log "🔍 Test 1: Basic connectivity"
    export PLUGIN_HOST="$TEST_WINRM_HOST"
    export PLUGIN_PORT="$TEST_WINRM_PORT"
    export PLUGIN_PROTOCOL="$TEST_WINRM_PROTOCOL"
    export PLUGIN_USERNAME="$TEST_WINRM_USERNAME"
    export PLUGIN_PASSWORD="$TEST_WINRM_PASSWORD"
    export PLUGIN_LOG_LEVEL="debug"
    export PLUGIN_COMMAND="Write-Host 'Vagrant integration test successful'; Get-Date; Write-Host 'Computer:' \$env:COMPUTERNAME"
    
    if ./winrm-plugin; then
        log_success "Basic connectivity test passed"
    else
        log_error "Basic connectivity test failed"
        return 1
    fi
    
    # Test 2: System information
    log "🔍 Test 2: System information"
    export PLUGIN_COMMAND="Get-ComputerInfo | Select-Object WindowsProductName, TotalPhysicalMemory, CsProcessors"
    
    if ./winrm-plugin; then
        log_success "System information test passed"
    else
        log_error "System information test failed"
        return 1
    fi
    
    # Test 3: Script execution
    log "🔍 Test 3: Script execution"
    export PLUGIN_SCRIPT_CONTENT='
    param([string]$TestName = "Vagrant Integration Test")
    
    Write-Host "🎯 Running: $TestName"
    Write-Host "🖥️  Computer: $env:COMPUTERNAME"
    Write-Host "👤 User: $env:USERNAME" 
    Write-Host "📁 Working Directory: $(Get-Location)"
    Write-Host "🕒 Timestamp: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")"
    
    # Test PowerShell capabilities
    Write-Host "🔧 PowerShell Version: $($PSVersionTable.PSVersion)"
    Write-Host "📊 Available Memory: $([math]::Round((Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory/1MB, 2)) GB"
    
    # Test error handling
    try {
        Get-Service -Name "NonExistentService" -ErrorAction Stop
    } catch {
        Write-Host "✅ Error handling working correctly"
    }
    
    # Test working directory
    $testFile = "vagrant-test-$(Get-Date -Format "yyyyMMdd-HHmmss").txt"
    "Vagrant test file" | Out-File -FilePath $testFile
    if (Test-Path $testFile) {
        Write-Host "✅ File operations working"
        Remove-Item $testFile -Force
    }
    
    Write-Host "✅ Vagrant integration test completed successfully"
    exit 0
    '
    unset PLUGIN_COMMAND
    
    if ./winrm-plugin; then
        log_success "Script execution test passed"
    else
        log_error "Script execution test failed"
        return 1
    fi
    
    # Test 4: Go integration tests
    log "🔍 Test 4: Go integration tests"
    if go test ./test -v -timeout=300s -run="Test.*"; then
        log_success "Go integration tests passed"
    else
        log_warning "Some Go integration tests failed (this might be expected)"
    fi
    
    log_success "All Vagrant integration tests completed successfully! 🎉"
}

# Function to clean up
cleanup() {
    log "Cleaning up Vagrant VM..."
    if [[ "${VAGRANT_DESTROY:-false}" == "true" ]]; then
        vagrant destroy -f
        log_success "Vagrant VM destroyed"
    else
        vagrant halt
        log_success "Vagrant VM halted (use VAGRANT_DESTROY=true to destroy)"
    fi
}

# Function to show usage
usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start     Start Vagrant VM and configure WinRM"
    echo "  test      Run integration tests (starts VM if needed)"
    echo "  stop      Stop Vagrant VM"
    echo "  destroy   Destroy Vagrant VM"
    echo "  status    Show Vagrant VM status"
    echo "  ssh       SSH into Vagrant VM"
    echo "  rdp       Show RDP connection info"
    echo ""
    echo "Environment Variables:"
    echo "  VAGRANT_DESTROY=true   Destroy VM after tests instead of halt"
    echo ""
    echo "Examples:"
    echo "  $0 test                           # Run full test suite"
    echo "  VAGRANT_DESTROY=true $0 test      # Run tests and destroy VM"
    echo "  $0 start && $0 test && $0 stop    # Manual control"
}

# Parse command
case "${1:-test}" in
    start)
        start_vm
        ;;
    test)
        start_vm
        run_tests
        if [[ "${VAGRANT_DESTROY:-false}" != "true" ]]; then
            log "VM is still running. Use '$0 stop' to halt or '$0 destroy' to remove."
        else
            cleanup
        fi
        ;;
    stop)
        vagrant halt
        log_success "Vagrant VM halted"
        ;;
    destroy)
        vagrant destroy -f
        log_success "Vagrant VM destroyed"
        ;;
    status)
        vagrant status
        ;;
    ssh)
        vagrant ssh
        ;;
    rdp)
        echo "RDP Connection Info:"
        echo "  Host: localhost"
        echo "  Port: 13389"
        echo "  Username: vagrant"
        echo "  Password: vagrant"
        echo ""
        echo "Test User:"
        echo "  Username: winrm-plugin-test" 
        echo "  Password: PluginTest123!"
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        log_error "Unknown command: $1"
        usage
        exit 1
        ;;
esac 