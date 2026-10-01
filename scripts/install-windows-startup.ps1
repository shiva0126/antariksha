param([string]$Distro = 'Ubuntu')
$ErrorActionPreference = 'Stop'
# Run as the Windows user who owns the WSL distribution.
# Keep WSL alive; enabled Linux user services manage the application and tunnel.
$account = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
$action = New-ScheduledTaskAction -Execute "$env:SystemRoot\System32\wsl.exe" -Argument "-d $Distro --exec /usr/bin/sleep infinity"
$trigger = New-ScheduledTaskTrigger -AtLogOn -User $account
$principal = New-ScheduledTaskPrincipal -UserId $account -LogonType Interactive -RunLevel Limited
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -ExecutionTimeLimit ([TimeSpan]::Zero) -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName 'Astrisk WSL Server' -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Description 'Keep the Astrisk WSL server running after Windows sign-in.' -Force | Out-Null
Start-ScheduledTask -TaskName 'Astrisk WSL Server'
Get-ScheduledTask -TaskName 'Astrisk WSL Server' | Select-Object TaskName, State
