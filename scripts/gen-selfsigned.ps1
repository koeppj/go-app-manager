param(
    [string]$CertPath = "server.crt",
    [string]$KeyPath = "server.key",
	[string]$Hostname = [System.Net.Dns]::GetHostByName([System.Net.Dns]::GetHostName()).HostName
)

$ErrorActionPreference = "Stop"

$paths = @($CertPath, $KeyPath) | ForEach-Object { Split-Path -Parent $_ } | Where-Object { $_ -ne "" } | Select-Object -Unique
foreach ($p in $paths) {
    if (-not (Test-Path $p)) {
        New-Item -ItemType Directory -Path $p -Force | Out-Null
    }
}

$genDir = Join-Path $PSScriptRoot "gencert"
& go run "$genDir" --cert "$CertPath" --key "$KeyPath" --host "$Hostname"
