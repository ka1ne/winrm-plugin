# WinRM Drone Plugin

A Drone plugin for executing PowerShell commands on Windows machines via WinRM. Built for Harness CI/CD pipelines with support for NTLM/Kerberos authentication and proxy configurations.

## Features

- **Authentication**: NTLM and Kerberos support
- **Proxy Support**: HTTP/HTTPS proxy with no_proxy bypass
- **Flexible Execution**: PowerShell commands with configurable options
- **Structured Logging**: JSON and text formats with multiple log levels
- **Retry Logic**: Customizable connection retry with backoff
- **Session Management**: Optional persistent PowerShell sessions

## Usage

### Basic Command Execution

```yaml
# .drone.yml
kind: pipeline
type: docker
name: windows-deployment

steps:
- name: deploy-to-windows
  image: your-registry/winrm-plugin:latest
  settings:
    host: build-server.company.com
    username: domain\\serviceaccount
    password:
      from_secret: windows_password
    command: |
      Get-Service -Name "MyService" | Restart-Service
      Write-Host "Service restarted successfully"
```

### Advanced Configuration

```yaml
- name: complex-deployment
  image: your-registry/winrm-plugin:latest
  settings:
    host: build-server.company.com
    port: 5986
    protocol: https
    auth_type: ntlm
    username: domain\\serviceaccount
    password:
      from_secret: windows_password
    working_dir: "C:\\deployments"
    timeout: 120
    max_retries: 5
    retry_interval: 10
    log_level: debug
    log_format: json
    persist_session: true
    command: |
      # Multi-step deployment
      Stop-Service -Name "MyApp" -Force
      Copy-Item "\\\\share\\releases\\latest\\*" -Destination "C:\\deployments\\" -Recurse -Force
      Start-Service -Name "MyApp"
      Test-NetConnection localhost -Port 8080

### Script File Execution

```yaml
- name: deploy-with-script
  image: your-registry/winrm-plugin:latest
  settings:
    host: build-server.company.com
    username: domain\\serviceaccount
    password:
      from_secret: windows_password
    script_path: "/drone/src/deployment-scripts/deploy.ps1"
    working_dir: "C:\\deployments"

### Inline Script Content

```yaml
- name: deploy-with-inline-script
  image: your-registry/winrm-plugin:latest
  settings:
    host: build-server.company.com
    username: domain\\serviceaccount
    password:
      from_secret: windows_password
    script_content: |
      param(
          [string]$ServiceName = "MyApp",
          [string]$DeployPath = "C:\deployments"
      )
      
      Write-Host "Deploying $ServiceName to $DeployPath"
      Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
      
      # Copy deployment files
      $source = "\\share\releases\latest\*"
      Copy-Item $source -Destination $DeployPath -Recurse -Force
      
      Start-Service -Name $ServiceName
      Get-Service -Name $ServiceName
```

## Configuration

### **Harness Pipeline Settings**
In your Harness pipeline, specify settings directly (Harness automatically adds `PLUGIN_` prefix):

```yaml
settings:
  # Required
  host: build-server.company.com
  username: domain\\serviceaccount
  password:
    from_secret: windows_password
  
  # Execution (choose one)
  command: "Get-Service MyApp | Restart-Service"
  # script_path: "/path/to/script.ps1"
  # script_content: |
  #   Write-Host "Multi-line script"
  #   Get-Date
  
  # Optional
  port: 5986
  protocol: https
  auth_type: ntlm
  working_dir: "C:\\deployments"
  timeout: 30
  max_retries: 3
  retry_interval: 5
  log_level: info
  log_format: text
  persist_session: false
  stream_output: false
```

### **Environment Variables**
When testing locally or using advanced configurations, use these environment variables:

#### Required
- `PLUGIN_HOST` - Target Windows machine hostname/IP
- `PLUGIN_USERNAME` - Windows username (domain\\user format for domain accounts)
- `PLUGIN_PASSWORD` - Windows password

#### Execution (choose one)
- `PLUGIN_COMMAND` - PowerShell command(s) to execute
- `PLUGIN_SCRIPT_PATH` - Local path to PowerShell script file (.ps1)
- `PLUGIN_SCRIPT_CONTENT` - Inline PowerShell script content

#### Optional
- `PLUGIN_PORT` - WinRM port (default: 5986)
- `PLUGIN_PROTOCOL` - http or https (default: https)
- `PLUGIN_AUTH_TYPE` - ntlm or kerberos (default: ntlm)
- `PLUGIN_WORKING_DIR` - Working directory (default: C:\\)
- `PLUGIN_TIMEOUT` - Connection timeout in seconds (default: 30)
- `PLUGIN_MAX_RETRIES` - Maximum retry attempts (default: 3)
- `PLUGIN_RETRY_INTERVAL` - Retry interval in seconds (default: 5)
- `PLUGIN_LOG_LEVEL` - error, warn, info, debug, verbose (default: info)
- `PLUGIN_LOG_FORMAT` - text or json (default: text)
- `PLUGIN_DEBUG_MODE` - Enable debug mode (default: false)
- `PLUGIN_PERSIST_SESSION` - Keep PowerShell session open (default: false)
- `PLUGIN_STREAM_OUTPUT` - Stream output in real-time (default: false)

#### Proxy Configuration
- `HTTP_PROXY` - HTTP proxy URL
- `HTTPS_PROXY` - HTTPS proxy URL  
- `NO_PROXY` - Comma-separated list of hosts to bypass proxy

## Building

```bash
# Build binary
go build -o winrm-plugin .

# Build Docker image
docker build -t winrm-plugin .
```

## Development

### Prerequisites
- Go 1.21+
- Access to Windows machine with WinRM enabled

### Testing
```bash
# Set environment variables
export PLUGIN_HOST=your-windows-host
export PLUGIN_USERNAME=domain\\user
export PLUGIN_PASSWORD=password
export PLUGIN_COMMAND="Get-ComputerInfo | Select-Object WindowsProductName, TotalPhysicalMemory"
export PLUGIN_LOG_LEVEL=debug

# Run plugin
go run main.go
```

## Testing

Integration testing uses Vagrant for isolated, reproducible Windows VMs:

```bash
# Prerequisites
brew install vagrant virtualbox  # macOS

# Run tests (first time downloads Windows box ~4GB)
./scripts/vagrant-test.sh test

# Quick start with guided setup
./scripts/vagrant-quick-start.sh
```

**Test Environment:**
- Windows Server 2022 VM
- Pre-configured WinRM (HTTP:15985, HTTPS:15986)
- Test user: `winrm-plugin-test` / `PluginTest123!`
- Port forwarding to localhost

**Commands:**
```bash
./scripts/vagrant-test.sh test      # Full test suite
./scripts/vagrant-test.sh start     # Start VM only
./scripts/vagrant-test.sh stop      # Stop VM
./scripts/vagrant-test.sh destroy   # Remove VM
```

## TODO

- [ ] Retry logic implementation
- [ ] Session persistence
- [ ] Real-time output streaming
- [ ] Remote script URL support
- [ ] Certificate authentication
- [ ] Multi-architecture Docker builds

## Architecture

```
winrm-plugin/
├── main.go                  # Entry point and configuration
├── internal/
│   ├── logger/             # Structured logging
│   ├── winrm/              # WinRM client and connection handling
│   └── powershell/         # PowerShell execution (planned)
├── Dockerfile              # Container build
└── README.md              # This file
```

## Security Considerations

- Store sensitive credentials in Harness secrets
- Use HTTPS protocol for WinRM connections
- Validate and sanitize PowerShell commands
- Enable WinRM logging on target machines for audit trails
- Consider certificate-based authentication for production

## Troubleshooting

### Common Issues

1. **Connection Refused**: Verify WinRM is enabled and firewall allows connections
2. **Authentication Failed**: Check domain\\username format and credentials
3. **Proxy Issues**: Verify proxy settings and no_proxy configuration
4. **PowerShell Errors**: Enable debug logging to see detailed command output

### WinRM Configuration

Enable WinRM on target Windows machine:
```powershell
# Enable WinRM
Enable-PSRemoting -Force

# Configure HTTPS listener (recommended)
winrm create winrm/config/Listener?Address=*+Transport=HTTPS @{Hostname="your-hostname";CertificateThumbprint="your-cert-thumbprint"}

# Allow domain authentication
Set-Item WSMan:\localhost\Service\Auth\Kerberos -Value $true
Set-Item WSMan:\localhost\Service\Auth\Negotiate -Value $true
```

## License

MIT License - see LICENSE file for details. 