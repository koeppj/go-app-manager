param(
    [string]$Version = "1.0.0",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$repo = Split-Path -Parent $root
Set-Location $repo

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Derive commit/date metadata if available.
$commit = ""
try {
    $commit = (git rev-parse --short HEAD)
} catch {
    $commit = "unknown"
}
$buildDate = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssK")

$ldflags = "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=$Version"
$ldflags += " -X github.com/koeppj/go-app-manager/internal/app.BuildCommit=$commit"
$ldflags += " -X github.com/koeppj/go-app-manager/internal/app.BuildDate=$buildDate"

Write-Host "Building tray..."
go build -ldflags $ldflags -o (Join-Path $OutputDir "tray.exe") .\cmd\tray

Write-Host "Done. Artifacts in $OutputDir"
