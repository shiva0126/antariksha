# Run as Administrator on Windows after reviewing the router/DNS checklist.
# Uses dedicated host ports; does not modify unrelated port 80/443 services.
param([string]$Distro = 'Ubuntu')
$ErrorActionPreference = 'Stop'
$principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run PowerShell as Administrator.' }
$addresses = (& wsl.exe -d $Distro hostname -I).Trim() -split '\s+'
if ($LASTEXITCODE -ne 0) { throw 'Cannot read the WSL address.' }
$address = $addresses | Where-Object { $_ -match '^\d+\.\d+\.\d+\.\d+$' } | Select-Object -First 1
if (-not $address) { throw 'WSL IPv4 address not found.' }
$parsed = [System.Net.IPAddress]::Parse($address)
$existing = (& netsh interface portproxy show v4tov4) -join "`n"
foreach ($port in @(8088,8443)) {
 if ($existing -match "(?m)^\s*\S+\s+$port\s+") { throw "A forwarding rule already owns port $port. Review it manually before proceeding." }
 if (Get-NetTCPConnection -State Listen -LocalPort $port -ErrorAction SilentlyContinue) { throw "Port $port is already in use." }
}
foreach ($port in @(8088,8443)) {
 & netsh interface portproxy add v4tov4 listenaddress=0.0.0.0 listenport=$port connectaddress=$address connectport=$port
 if ($LASTEXITCODE -ne 0) { throw "Port forwarding failed at $port; inspect rules before retrying." }
 if (-not (Get-NetFirewallRule -Name "Astrisk-WSL-$port" -ErrorAction SilentlyContinue)) {
  New-NetFirewallRule -Name "Astrisk-WSL-$port" -DisplayName "Astrisk WSL $port" -Direction Inbound -Action Allow -Protocol TCP -LocalPort $port -Profile Private
 }
}
Write-Output 'Forward router WAN TCP 80 to this PC port 8088, and WAN TCP 443 to port 8443.'
Write-Output 'WSL addresses may change after restart. Review and update these two rules when needed.'
