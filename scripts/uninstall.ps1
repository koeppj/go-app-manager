$ErrorActionPreference = "Stop"

Write-Host "Stopping service"
.\dist\service.exe --stop

Write-Host "Uninstalling service"
.\dist\service.exe --uninstall

Write-Host "Removing tray scheduled task"
if (Get-ScheduledTask -TaskName "GoAppManagerTray" -ErrorAction SilentlyContinue) {
    Unregister-ScheduledTask -TaskName "GoAppManagerTray" -Confirm:$false
}

Write-Host "Done."
