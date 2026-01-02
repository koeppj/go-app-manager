$ErrorActionPreference = "Stop"
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$sudoCmd = Get-Command sudo -ErrorAction SilentlyContinue
if (-not $isAdmin) {
    if ($sudoCmd) {
        Write-Host "Not elevated. Re-running with sudo..."
        & $sudoCmd powershell -NoLogo -NoProfile -ExecutionPolicy Bypass -File $PSCommandPath
        exit $LASTEXITCODE
    }
    Write-Host "Not elevated. Re-running with RunAs..."
    $psiArgs = "-NoLogo -NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`""
    $proc = Start-Process -FilePath "powershell.exe" -ArgumentList $psiArgs -Verb RunAs -Wait -PassThru
    if ($proc.ExitCode -ne 0) { exit $proc.ExitCode }
    exit 0
}

Write-Host "Removing tray scheduled task"
if (Get-ScheduledTask -TaskName "GoAppManagerTray" -ErrorAction SilentlyContinue) {
    Unregister-ScheduledTask -TaskName "GoAppManagerTray" -Confirm:$false
}

Write-Host "Removing installed tray binary (if present)"
$installDir = Join-Path ${env:ProgramFiles} "GoAppManager"
$installedExe = Join-Path $installDir "tray.exe"
if (Test-Path $installedExe) {
    Remove-Item $installedExe -Force -ErrorAction SilentlyContinue
}
if (Test-Path $installDir -and -not (Get-ChildItem $installDir)) {
    Remove-Item $installDir -Force -ErrorAction SilentlyContinue
}

Write-Host "Done."
