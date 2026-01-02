# Go App Manager (Windows)

Windows-only utility that runs as a service (Session 0) to manage a list of executables via a secured HTTPS API and serves a small web UI. A separate tray agent runs in the user session to control the service.

## Features
- HTTPS-only API + web UI with bearer auth and CIDR allowlist.
- Job Object–backed process management (start/stop/restart/status).
- Windows service install helpers and tray agent for interactive control.

## Repository Layout
- `cmd/service`: Windows service entrypoint (SCM-aware).
- `cmd/tray`: Tray agent entrypoint (systray UI + API client).
- `internal/app`: Config, logging, security, HTTP server, process manager, installer helpers.
- `internal/trayui`: Tray UI and API client.
- `web/ui/static`: Embedded SPA served by the service.
- `scripts`: Install/uninstall/dev and self-signed cert helper.
- `assets/sample-config.yaml`: Example configuration.

## Prerequisites
- Go 1.25+ on Windows.
- Administrative rights to install the service and scheduled task for the tray.

## Build
```powershell
go mod tidy
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=1.0.0" -o dist\service.exe .\cmd\service
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=1.0.0" -o dist\tray.exe .\cmd\tray
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
.\scripts\dev-run.ps1 -ConfigPath "C:\ProgramData\GoAppManager\config\config.yaml"
```
Runs in console with logging to `C:\ProgramData\GoAppManager\logs\`.

## Install / Uninstall
```powershell
# Install service + tray (generates token/certs if missing)
.\scripts\install.ps1 -Version 1.0.0 -ConfigPath "C:\ProgramData\GoAppManager\config\config.yaml"

# Uninstall service + tray task
.\scripts\uninstall.ps1
```
- Service name: `GoAppManager` (display: `Go App Manager Service`).
- Service flags available via `service.exe`: `--install`, `--uninstall`, `--start`, `--stop`, `--console`, `--config`.

## Tray Agent
- Installed as a scheduled task (`GoAppManagerTray`) at user logon.
- Flags: `--service-url https://host:8443`, `--token <bearer>`, `--ca <custom CA>`.
- Provides menu actions Start/Stop/Restart per program and “Open Web UI”.

## Web UI
- Served from the service at `/` on the configured HTTPS port.
- Paste the bearer token in the UI to authenticate; lists programs and allows control.

## Security Notes
- HTTPS is mandatory; HTTP is not supported.
- Bearer token required for `/api/*`.
- CIDR allowlist enforced; empty list denies all.
- Optional client CA in config for mTLS-ready deployments.
