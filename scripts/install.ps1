param(
    [string]$Version = "1.0.0",
    [string]$ConfigPath = "$env:ProgramData\\GoAppManager\\config\\config.yaml"
)

$ErrorActionPreference = "Stop"
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
$sudoCmd = Get-Command sudo -ErrorAction SilentlyContinue
if (-not $isAdmin) {
    $argList = @()
    foreach ($item in $PSBoundParameters.GetEnumerator()) {
        $argList += "-$($item.Key)"
        $argList += "`"$($item.Value)`""
    }

    if ($sudoCmd) {
        Write-Host "Not elevated. Re-running with sudo..."
        & $sudoCmd powershell -NoLogo -NoProfile -ExecutionPolicy Bypass -File $PSCommandPath @argList
        exit $LASTEXITCODE
    }

    Write-Host "Not elevated. Re-running with RunAs..."
    $psiArgs = "-NoLogo -NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`" $($argList -join ' ')"
    $proc = Start-Process -FilePath "powershell.exe" -ArgumentList $psiArgs -Verb RunAs -Wait -PassThru
    if ($proc.ExitCode -ne 0) { exit $proc.ExitCode }
    exit 0
}

$base = "$env:ProgramData\\GoAppManager"
$configDir = Join-Path $base "config"
$logsDir = Join-Path $base "logs"
$secretsDir = Join-Path $base "secrets"
$certsDir = Join-Path $base "certs"
$installDir = Join-Path ${env:ProgramFiles} "GoAppManager"

Write-Host "Creating directories under $base"
@($base, $configDir, $logsDir, $secretsDir, $certsDir) | ForEach-Object {
    if (-not (Test-Path $_)) { New-Item -ItemType Directory -Path $_ | Out-Null }
}

if (-not (Test-Path $ConfigPath)) {
    Write-Host "Placing sample config at $ConfigPath"
    Copy-Item ".\\assets\\sample-config.yaml" $ConfigPath
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

Write-Host "Building tray"
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=$Version" -o dist\\tray.exe .\\cmd\\tray

Write-Host "Installing tray binary to $installDir"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir | Out-Null }
Copy-Item ".\\dist\\tray.exe" (Join-Path $installDir "tray.exe") -Force
$trayExe = (Resolve-Path (Join-Path $installDir "tray.exe")).Path

Write-Host "Registering tray agent (current user logon)"
$args = "--config `"$ConfigPath`" --ca `"$certsDir\\server.crt`" --hide-console"
Register-ScheduledTask -TaskName "GoAppManagerTray" -Action (New-ScheduledTaskAction -Execute $trayExe -Argument $args) -Trigger (New-ScheduledTaskTrigger -AtLogOn) -RunLevel Highest -Force

Write-Host "Done."
