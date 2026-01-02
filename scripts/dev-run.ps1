param(
    [string]$ConfigPath = "$env:ProgramData\\GoAppManager\\config\\config.yaml",
    [string]$CaPath = "$env:ProgramData\\GoAppManager\\certs\\server.crt"
)

$ErrorActionPreference = "Stop"

go run .\\cmd\\tray --config $ConfigPath --ca $CaPath
