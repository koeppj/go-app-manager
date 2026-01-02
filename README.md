# Go App Manager (Windows)

Windows-only utility that runs entirely under the interactive user session (tray app). The tray hosts the HTTPS API + web UI and manages configured executables (start/stop/restart/status) using Windows Job Objects.

## Features
- HTTPS-only API + web UI with bearer auth and CIDR allowlist.
- Job Object–backed process management to avoid orphaned processes.
- Tray-driven deployment (no Windows service required).

## Repository Layout
- `cmd/tray`: Tray entrypoint (hosts API + web UI + systray controls).
- `internal/app`: Config, logging, security, HTTP server, process manager.
- `internal/trayui`: Tray UI and API client.
- `web/ui/static`: Embedded SPA served by the tray-hosted server.
- `scripts`: Install/uninstall/dev/build and self-signed cert helper.
- `assets/sample-config.yaml`: Example configuration.

## Prerequisites
- Go 1.25+ on Windows.
- Rights to register a logon scheduled task for the tray (optional convenience).

## Build
```powershell
go mod tidy
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=1.0.0" -o dist\tray.exe .\cmd\tray
```
or
```powershell
.\scripts\build.ps1 -Version 1.0.0
```

## Configuration
- Default path: `C:\ProgramData\GoAppManager\config\config.yaml`
- Copy and edit `assets\sample-config.yaml`.
- Key settings:
  - `server.bind` / `server.port`
  - `server.tls.certFile` / `server.tls.keyFile` (HTTPS required)
  - `server.auth.bearerTokenFile` path to token (plain text)
  - `server.network.allowedCIDRs` list of allowed client subnets (empty = deny)
  - `programs[]` entries (`name`, `command`, `args`, `workDir`, `stop` method, optional `autoStart`)
- Tokens stored under `C:\ProgramData\GoAppManager\secrets\api.token`. Certificates under `C:\ProgramData\GoAppManager\certs\`.

## Running (dev/console)
```powershell
.\scripts\dev-run.ps1 -ConfigPath "C:\ProgramData\GoAppManager\config\config.yaml" -CaPath "C:\ProgramData\GoAppManager\certs\server.crt"
```
Runs the tray (with systray icon) and hosts the API/UI. Logs go to `C:\ProgramData\GoAppManager\logs\`.

## Install / Uninstall
```powershell
# Install tray (generates token/certs if missing; registers logon task)
.\scripts\install.ps1 -Version 1.0.0 -ConfigPath "C:\ProgramData\GoAppManager\config\config.yaml"

# Uninstall tray task
.\scripts\uninstall.ps1
```
- Scheduled task name: `GoAppManagerTray`.

## Tray Agent
- Hosts HTTPS API + web UI and provides systray menu actions (Start/Stop/Restart per program, Open Web UI).
- Flags: `--config`, `--service-url` (override), `--token` (override), `--ca` (custom CA; defaults to server cert).

## Web UI
- Served from the tray at `/` on the configured HTTPS port.
- Paste the bearer token in the UI to authenticate; lists programs and allows control.

## Security Notes
- HTTPS is mandatory; HTTP is not supported.
- Bearer token required for `/api/*`.
- CIDR allowlist enforced; include loopback if you access via `localhost`.
- Optional client CA in config for mTLS-ready deployments.
