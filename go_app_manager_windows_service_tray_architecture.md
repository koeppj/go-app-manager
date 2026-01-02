# Go App Manager – Windows Service + Tray Architecture

**Module name:** `github.com/koeppj/go-app-manager`

This document defines the complete project structure, responsibilities, and implementation context for a **Windows-only Go utility** that:

- Runs as a **Windows Service**
- Exposes a **LAN-accessible web UI + API**
- Manages a defined list of executables (start / stop / restart / status)
- Provides a **system tray UI** (user session) that controls the service

This file is intended to be given directly to **OpenAI Codex** to implement the solution end-to-end.

---

## 1. Architecture Summary

### Core constraints (Windows)
- Windows services run in **Session 0** and **cannot display UI**.
- Tray icons must run in an **interactive user session**.

### Chosen architecture (recommended)

| Component | Responsibility | Session |
|---------|---------------|---------|
| Service | Process control, web UI, API | Session 0 |
| Tray Agent | Tray icon + menu, API client | User session |

### Communication
- **HTTPS over LAN** (`0.0.0.0:<port>`)
- **Bearer token authentication**
- **CIDR allowlist**
- Optional TLS client certs (mTLS-ready)

---

## 2. Repository Layout

```text
go-app-manager/
  README.md

  go.mod
  go.sum

  cmd/
    service/
      main.go                 # Windows service entrypoint
    tray/
      main.go                 # Tray agent entrypoint

  internal/
    app/
      buildinfo.go            # version/commit/date via ldflags
      paths.go                # Windows paths (ProgramData, logs, secrets)

      config/
        config.go             # config structs + load/validate
        defaults.go

      logging/
        logging.go             # file logging + optional Windows Event Log

      security/
        auth.go                # bearer token auth middleware
        tls.go                 # TLS config + cert loading
        cidr.go                # LAN allowlist enforcement

      httpserver/
        server.go              # HTTP server bootstrap
        routes.go              # route registration
        handlers_api.go        # /api/* handlers
        handlers_ui.go         # embedded UI serving

      process/
        manager.go             # program orchestration
        program.go             # program model
        jobobject.go           # Windows Job Object wrapper
        winproc.go             # Windows process flags

      state/
        store.go               # optional runtime persistence

      installer/
        service_install.go    # install/uninstall helpers
        acl.go                # Windows ACL helpers

    trayui/
      tray.go                 # systray UI
      client.go               # service API client
      icons/
        app.ico

  web/
    ui/
      static/
        index.html
        app.js
        styles.css

  assets/
    sample-config.yaml

  scripts/
    install.ps1
    uninstall.ps1
    dev-run.ps1
    gen-selfsigned.ps1

  dist/
```

---

## 3. Module Path Usage

All imports **must** use:

```go
module github.com/koeppj/go-app-manager
```

Example imports:

```go
import "github.com/koeppj/go-app-manager/internal/app/config"
import "github.com/koeppj/go-app-manager/internal/app/process"
```

---

## 4. Service (`cmd/service/main.go`)

### Responsibilities
- Run as a Windows service (SCM-aware)
- Load configuration
- Initialize logging
- Start HTTPS server
- Manage child executables

### Required flags

| Flag | Purpose |
|----|--------|
| `--console` | Run without SCM (dev/debug) |
| `--config` | Path to config file |
| `--install` | Install Windows service |
| `--uninstall` | Remove Windows service |
| `--start` | Start service |
| `--stop` | Stop service |

### Windows service implementation
- Use `golang.org/x/sys/windows/svc`
- Service name: `GoAppManager`
- Display name: `Go App Manager Service`

---

## 5. Tray Agent (`cmd/tray/main.go`)

### Responsibilities
- Show tray icon
- Display managed programs + status
- Call service API
- Open web UI

### Required flags

| Flag | Purpose |
|----|--------|
| `--service-url` | Base URL of service (e.g. https://host:8443) |
| `--token` | API bearer token (optional override) |
| `--ca` | Custom CA cert (self-signed TLS) |

### Tray behavior
- One menu section per program
- Actions: Start / Stop / Restart
- Refresh status every 5 seconds

---

## 6. Configuration File (YAML)

**Default location:**
```
C:\ProgramData\GoAppManager\config\config.yaml
```

### Sample

```yaml
server:
  bind: "0.0.0.0"
  port: 8443
  publicBaseUrl: "https://yourhost:8443"
  tls:
    enabled: true
    certFile: "C:\\ProgramData\\GoAppManager\\certs\\server.crt"
    keyFile:  "C:\\ProgramData\\GoAppManager\\certs\\server.key"
  auth:
    bearerTokenFile: "C:\\ProgramData\\GoAppManager\\secrets\\api.token"
  network:
    allowedCIDRs:
      - "192.168.1.0/24"

programs:
  - name: "nginx"
    command: "C:\\Tools\\nginx\\nginx.exe"
    args: ["-c", "C:\\Tools\\nginx\\conf\\nginx.conf"]
    workDir: "C:\\Tools\\nginx"
    stop:
      method: "jobobject"
      timeoutSeconds: 10
```

---

## 7. Process Management Rules

### Start
- Use `exec.Command`
- Apply `CREATE_NO_WINDOW`
- Assign to a **Windows Job Object**

### Stop
- Preferred: terminate Job Object
- Fallback: optional stop command

### Restart
- Stop → Start

### Status tracking
- running / stopped
- PID
- start time
- last exit code
- last error

---

## 8. HTTP API

### Endpoints

| Method | Path |
|------|------|
| GET | `/api/health` |
| GET | `/api/programs` |
| GET | `/api/programs/{name}` |
| POST | `/api/programs/{name}/start` |
| POST | `/api/programs/{name}/stop` |
| POST | `/api/programs/{name}/restart` |

### Middleware order
1. Logging
2. TLS enforcement
3. CIDR allowlist
4. Bearer token auth

---

## 9. Security Defaults

- HTTPS required
- Bearer token required for all `/api/*`
- CIDR allowlist enforced (default deny)
- Secrets stored under:
  ```
  C:\ProgramData\GoAppManager\secrets
  ```

---

## 10. Installation (PowerShell)

### `scripts/install.ps1`
- Create directories
- Generate or install TLS certs
- Create API token
- Install Windows service (`sc.exe`)
- Configure service recovery
- Register tray agent via Task Scheduler (At logon)

### Output binaries

```text
dist/service.exe
dist/tray.exe
```

---

## 11. Build Commands

```powershell
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=1.0.0" -o dist\service.exe .\cmd\service
go build -ldflags "-X github.com/koeppj/go-app-manager/internal/app.BuildVersion=1.0.0" -o dist\tray.exe .\cmd\tray
```

---

## 12. Non‑Negotiable Notes (Do Not Deviate)

- Tray UI **must not** run inside the service
- Job Objects **must** be used to avoid orphaned child processes
- LAN access **must** be authenticated
- Plain HTTP is **not allowed**

---

## 13. Optional Future Enhancements

- mTLS enforcement
- Role-based access (read-only vs control)
- Windows Event Log integration
- Auto-restart policies
- Program health checks

---

**End of implementation instructions.**

