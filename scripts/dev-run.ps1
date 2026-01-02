param(
    [string]$ConfigPath = "$env:ProgramData\\GoAppManager\\config\\config.yaml"
)

$ErrorActionPreference = "Stop"

go run .\\cmd\\service --console --config $ConfigPath
