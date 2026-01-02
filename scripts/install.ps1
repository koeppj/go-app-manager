param(
    [string]$Version = "1.0.0",
    [string]$ConfigPath = "$env:ProgramData\\GoAppManager\\config\\config.yaml"
)

$ErrorActionPreference = "Stop"

$base = "$env:ProgramData\\GoAppManager"
$configDir = Join-Path $base "config"
$logsDir = Join-Path $base "logs"
$secretsDir = Join-Path $base "secrets"
$certsDir = Join-Path $base "certs"

Write-Host "Creating directories under $base"
@($base, $configDir, $logsDir, $secretsDir, $certsDir) | ForEach-Object {
    if (-not (Test-Path $_)) { New-Item -ItemType Directory -Path $_ | Out-Null }
}

if (-not (Test-Path $ConfigPath)) {
    Write-Host "Placing sample config at $ConfigPath"
    Copy-Item "..\\assets\\sample-config.yaml" $ConfigPath
}

$tokenFile = Join-Path $secretsDir "api.token"
if (-not (Test-Path $tokenFile)) {
    Write-Host "Generating API token"
    [guid]::NewGuid().ToString() | Out-File -Encoding ascii $tokenFile
}

if (-not (Test-Path (Join-Path $certsDir "server.crt"))) {
    Write-Host "Generating self-signed certificate"
    & ".\\gen-selfsigned.ps1" -CertPath (Join-Path $certsDir "server.crt") -KeyPath (Join-Path $certsDir "server.key")
}

Write-Host "Building service and tray"
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=$Version" -o dist\\service.exe .\\cmd\\service
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=$Version" -o dist\\tray.exe .\\cmd\\tray

Write-Host "Installing Windows service"
.\dist\service.exe --install --config $ConfigPath

Write-Host "Registering tray agent (current user logon)"
$trayExe = (Resolve-Path ".\\dist\\tray.exe").Path
Register-ScheduledTask -TaskName "GoAppManagerTray" -Action (New-ScheduledTaskAction -Execute $trayExe -Argument "--service-url https://localhost:8443") -Trigger (New-ScheduledTaskTrigger -AtLogOn) -RunLevel Highest -Force

Write-Host "Done."
