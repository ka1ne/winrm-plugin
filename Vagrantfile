# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  # Windows Server 2022 box
  config.vm.box = "gusztavvargadr/windows-server-2022-standard"
  config.vm.box_version = "~> 2022.0"
  
  # VM configuration
  config.vm.hostname = "winrm-test"
  config.vm.define "winrm-test"
  
  # Network configuration for WinRM access
  config.vm.network "private_network", type: "dhcp"
  config.vm.network "forwarded_port", guest: 5985, host: 15985, id: "winrm-http"
  config.vm.network "forwarded_port", guest: 5986, host: 15986, id: "winrm-https"
  config.vm.network "forwarded_port", guest: 3389, host: 13389, id: "rdp"
  
  # Provider-specific configuration
  config.vm.provider "virtualbox" do |vb|
    vb.name = "winrm-plugin-test"
    vb.memory = "2048"
    vb.cpus = 2
    vb.gui = false
    
    # Enable clipboard and drag & drop
    vb.customize ["modifyvm", :id, "--clipboard-mode", "bidirectional"]
    vb.customize ["modifyvm", :id, "--draganddrop", "bidirectional"]
  end
  
  # VMware provider (alternative)
  config.vm.provider "vmware_desktop" do |vmware|
    vmware.vmx["memsize"] = "2048"
    vmware.vmx["numvcpus"] = "2"
    vmware.vmx["displayName"] = "WinRM Plugin Test"
  end
  
  # Hyper-V provider (Windows hosts)
  config.vm.provider "hyperv" do |h|
    h.memory = 2048
    h.cpus = 2
    h.vmname = "winrm-plugin-test"
  end
  
  # Configure WinRM
  config.vm.communicator = "winrm"
  config.winrm.username = "vagrant"
  config.winrm.password = "vagrant"
  config.winrm.transport = :plaintext
  config.winrm.basic_auth_only = true
  config.winrm.execution_time_limit = 300
  
  # Provision WinRM for testing
  config.vm.provision "shell", privileged: true, powershell_elevated_interactive: false, inline: <<-SHELL
    Write-Host "🔧 Configuring WinRM for plugin testing..." -ForegroundColor Green
    
    # Enable WinRM
    Enable-PSRemoting -Force
    
    # Configure authentication methods
    winrm set winrm/config/service/Auth '@{Basic="true"}'
    winrm set winrm/config/service/Auth '@{Kerberos="true"}'
    winrm set winrm/config/service/Auth '@{Negotiate="true"}'
    
    # Allow unencrypted traffic (for testing only)
    winrm set winrm/config/service '@{AllowUnencrypted="true"}'
    
    # Increase memory limit
    winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="1024"}'
    
    # Configure firewall
    New-NetFirewallRule -DisplayName "WinRM-HTTP" -Direction Inbound -LocalPort 5985 -Protocol TCP -Action Allow -Force
    New-NetFirewallRule -DisplayName "WinRM-HTTPS" -Direction Inbound -LocalPort 5986 -Protocol TCP -Action Allow -Force
    
    # Create test user for plugin testing
    $testPassword = ConvertTo-SecureString "PluginTest123!" -AsPlainText -Force
    try {
        New-LocalUser -Name "winrm-plugin-test" -Password $testPassword -Description "WinRM plugin testing account" -ErrorAction Stop
        Add-LocalGroupMember -Group "Administrators" -Member "winrm-plugin-test" -ErrorAction Stop
        Write-Host "✅ Created test user 'winrm-plugin-test'" -ForegroundColor Green
    } catch {
        Write-Host "⚠️ Test user might already exist: $($_.Exception.Message)" -ForegroundColor Yellow
    }
    
    # Configure HTTPS listener with self-signed certificate
    $cert = New-SelfSignedCertificate -DnsName "localhost", "winrm-test" -CertStoreLocation "cert:\LocalMachine\My"
    $httpsListener = Get-WSManInstance -ResourceURI "winrm/config/listener" -SelectorSet @{Address="*";Transport="HTTPS"} -ErrorAction SilentlyContinue
    if (-not $httpsListener) {
        winrm create winrm/config/Listener?Address=*+Transport=HTTPS "@{Hostname=`"localhost`";CertificateThumbprint=`"$($cert.Thumbprint)`"}"
        Write-Host "✅ Created HTTPS listener" -ForegroundColor Green
    }
    
    # Test WinRM configuration
    Write-Host "🧪 Testing WinRM configuration..." -ForegroundColor Cyan
    $testResult = Test-WSMan -ComputerName localhost -Port 5985 -UseSSL:$false
    if ($testResult) {
        Write-Host "✅ WinRM HTTP is working" -ForegroundColor Green
    }
    
    try {
        $testResultHTTPS = Test-WSMan -ComputerName localhost -Port 5986 -UseSSL:$true
        if ($testResultHTTPS) {
            Write-Host "✅ WinRM HTTPS is working" -ForegroundColor Green
        }
    } catch {
        Write-Host "⚠️ WinRM HTTPS test failed (this is often normal): $($_.Exception.Message)" -ForegroundColor Yellow
    }
    
    # Display configuration
    Write-Host "`n📋 WinRM Configuration Summary:" -ForegroundColor Cyan
    winrm get winrm/config
    
    Write-Host "`n🎯 Plugin test commands:" -ForegroundColor Cyan
    Write-Host "export TEST_WINRM_HOST=localhost" -ForegroundColor White
    Write-Host "export TEST_WINRM_PORT=15985" -ForegroundColor White
    Write-Host "export TEST_WINRM_USERNAME=winrm-plugin-test" -ForegroundColor White
    Write-Host "export TEST_WINRM_PASSWORD=PluginTest123!" -ForegroundColor White
    Write-Host "./scripts/run-integration-tests.sh --target vagrant" -ForegroundColor White
    
    Write-Host "`n✅ WinRM plugin test environment ready!" -ForegroundColor Green
  SHELL
  
  # Additional provisioning for specific test scenarios
  config.vm.provision "test-apps", type: "shell", privileged: true, run: "never", inline: <<-SHELL
    Write-Host "🚀 Installing test applications..." -ForegroundColor Green
    
    # Install IIS for web application testing
    Enable-WindowsOptionalFeature -Online -FeatureName IIS-WebServerRole -All -NoRestart
    
    # Create a simple test service
    $servicePath = "C:\\TestService"
    if (-not (Test-Path $servicePath)) {
        New-Item -Path $servicePath -ItemType Directory -Force
        
        # Simple PowerShell service script
        $serviceScript = @'
while ($true) {
    Write-Host "Test service running at $(Get-Date)"
    Start-Sleep -Seconds 10
}
'@
        $serviceScript | Out-File -FilePath "$servicePath\\TestService.ps1" -Encoding UTF8
        
        Write-Host "✅ Test service created at $servicePath" -ForegroundColor Green
    }
  SHELL
end 