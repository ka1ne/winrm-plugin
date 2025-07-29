#!/bin/bash
# Quick start script for Vagrant testing

set -e

echo "🚀 WinRM Plugin - Vagrant Quick Start"
echo "====================================="
echo ""

# Check prerequisites
echo "🔍 Checking prerequisites..."

if ! command -v vagrant &> /dev/null; then
    echo "❌ Vagrant is not installed"
    echo ""
    echo "📦 Installation instructions:"
    echo "  macOS:     brew install vagrant"
    echo "  Ubuntu:    https://www.vagrantup.com/downloads"
    echo "  Windows:   https://www.vagrantup.com/downloads"
    echo ""
    exit 1
fi

if ! command -v VBoxManage &> /dev/null && ! command -v vmrun &> /dev/null; then
    echo "⚠️  No Vagrant provider found"
    echo ""
    echo "📦 Please install a provider:"
    echo "  VirtualBox: https://www.virtualbox.org/wiki/Downloads"
    echo "  VMware:     https://www.vmware.com/products/workstation-pro.html"
    echo "  Hyper-V:    Enable via Windows Features (Windows only)"
    echo ""
    echo "💡 VirtualBox is free and works on all platforms"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo "✅ Prerequisites look good!"
echo ""

# Build plugin
echo "🔨 Building WinRM plugin..."
go build -o winrm-plugin .
echo "✅ Plugin built successfully"
echo ""

# Show what will happen
echo "📋 What this script will do:"
echo "  1. Start Windows Server 2022 VM (downloads ~4GB on first run)"
echo "  2. Configure WinRM for testing"
echo "  3. Run comprehensive integration tests"
echo "  4. Leave VM running for future tests"
echo ""
echo "⏱️  First run: ~10-15 minutes (downloads Windows box)"
echo "⏱️  Subsequent runs: ~3-5 minutes"
echo ""

read -p "Continue with Vagrant setup? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "Setup cancelled"
    exit 0
fi

echo ""
echo "🎯 Starting Vagrant integration test..."
echo ""

# Run the test
./scripts/vagrant-test.sh test

echo ""
echo "🎉 Vagrant quick start completed!"
echo ""
echo "📚 Next steps:"
echo "  • VM is still running for faster future tests"
echo "  • Run tests again: ./scripts/vagrant-test.sh test"
echo "  • Stop VM: ./scripts/vagrant-test.sh stop"
echo "  • Destroy VM: ./scripts/vagrant-test.sh destroy"
echo "  • VM status: ./scripts/vagrant-test.sh status"
echo ""
echo "🔧 Manual testing:"
echo "  export PLUGIN_HOST=localhost"
echo "  export PLUGIN_PORT=15985"
echo "  export PLUGIN_USERNAME=winrm-plugin-test"
echo "  export PLUGIN_PASSWORD=PluginTest123!"
echo "  export PLUGIN_COMMAND='Write-Host \"Manual test successful\"'"
echo "  ./winrm-plugin"
echo ""
echo "�� Happy testing! 🚀" 