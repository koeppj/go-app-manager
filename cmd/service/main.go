package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"

	"github.com/koeppj/go-app-manager/internal/app"
	"github.com/koeppj/go-app-manager/internal/app/config"
	"github.com/koeppj/go-app-manager/internal/app/httpserver"
	"github.com/koeppj/go-app-manager/internal/app/installer"
	"github.com/koeppj/go-app-manager/internal/app/logging"
	"github.com/koeppj/go-app-manager/internal/app/process"
	"github.com/koeppj/go-app-manager/internal/app/security"
)

func main() {
	cfgPath := flag.String("config", app.ConfigPath(), "path to config file")
	console := flag.Bool("console", false, "run in console mode")
	install := flag.Bool("install", false, "install Windows service")
	uninstall := flag.Bool("uninstall", false, "uninstall Windows service")
	startSvc := flag.Bool("start", false, "start Windows service")
	stopSvc := flag.Bool("stop", false, "stop Windows service")
	flag.Parse()

	exe, _ := os.Executable()
	exe = filepath.Clean(exe)

	switch {
	case *install:
		args := []string{"--config", *cfgPath}
		if err := installer.InstallService(exe, args); err != nil {
			log.Fatalf("install failed: %v", err)
		}
		log.Println("service installed")
		return
	case *uninstall:
		if err := installer.UninstallService(); err != nil {
			log.Fatalf("uninstall failed: %v", err)
		}
		log.Println("service removed")
		return
	case *startSvc:
		if err := installer.StartService(); err != nil {
			log.Fatalf("start failed: %v", err)
		}
		log.Println("service start requested")
		return
	case *stopSvc:
		if err := installer.StopService(); err != nil {
			log.Fatalf("stop failed: %v", err)
		}
		log.Println("service stop requested")
		return
	}

	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to detect service: %v", err)
	}
	if *console || !isService {
		runConsole(*cfgPath)
		return
	}

	if err := svc.Run(installer.ServiceName, &serviceHandler{cfgPath: *cfgPath}); err != nil {
		log.Fatalf("service run failed: %v", err)
	}
}

func runConsole(cfgPath string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := runCore(ctx, cfgPath); err != nil {
		log.Fatalf("service error: %v", err)
	}
}

type serviceHandler struct {
	cfgPath string
}

func (h *serviceHandler) Execute(args []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runCore(ctx, h.cfgPath)
	}()

	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				status <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending}
				cancel()
				err := <-errCh
				if err != nil {
					return true, 1
				}
				status <- svc.Status{State: svc.Stopped}
				return false, 0
			}
		case err := <-errCh:
			if err != nil {
				return true, 1
			}
			return false, 0
		}
	}
}

func runCore(ctx context.Context, cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	logger, file, err := logging.Setup(app.LogDir())
	if err != nil {
		return err
	}
	defer file.Close()

	logger.Printf("Go App Manager %s", app.BuildInfo())
	logger.Printf("config: %s", cfgPath)

	token, err := security.LoadBearerToken(cfg.Server.Auth.BearerTokenFile)
	if err != nil {
		return err
	}

	manager := process.NewManager(cfg, logger)
	manager.AutoStart()

	server, err := httpserver.New(cfg, logger, manager, token)
	if err != nil {
		return err
	}
	if err := server.Start(); err != nil {
		return err
	}

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Stop(shutdownCtx)
	manager.StopAll(3 * time.Second)
	return nil
}
