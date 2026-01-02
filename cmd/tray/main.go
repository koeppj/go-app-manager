package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koeppj/go-app-manager/internal/app"
	"github.com/koeppj/go-app-manager/internal/app/config"
	"github.com/koeppj/go-app-manager/internal/app/httpserver"
	"github.com/koeppj/go-app-manager/internal/app/logging"
	"github.com/koeppj/go-app-manager/internal/app/process"
	"github.com/koeppj/go-app-manager/internal/app/security"
	"github.com/koeppj/go-app-manager/internal/trayui"
)

func main() {
	cfgPath := flag.String("config", app.ConfigPath(), "Path to config file")
	serviceURL := flag.String("service-url", "", "Override service URL (default uses config bind/port)")
	tokenOverride := flag.String("token", "", "Bearer token override")
	ca := flag.String("ca", "", "Custom CA certificate (defaults to server cert)")
	hideConsole := flag.Bool("hide-console", false, "Detach console window at startup")
	flag.Parse()

	maybeHideConsole(*hideConsole)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, file, err := logging.Setup(app.LogDir())
	if err != nil {
		log.Fatalf("logging: %v", err)
	}
	defer file.Close()

	token := *tokenOverride
	if token == "" {
		token, err = security.LoadBearerToken(cfg.Server.Auth.BearerTokenFile)
		if err != nil {
			log.Fatalf("load token: %v", err)
		}
	}

	mgr := process.NewManager(cfg, logger)
	mgr.AutoStart()

	server, err := httpserver.New(cfg, logger, mgr, token)
	if err != nil {
		log.Fatalf("http server: %v", err)
	}
	if err := server.Start(); err != nil {
		log.Fatalf("start server: %v", err)
	}

	baseURL := *serviceURL
	if baseURL == "" {
		if cfg.Server.PublicBaseURL != "" {
			baseURL = cfg.Server.PublicBaseURL
		} else {
			baseURL = fmt.Sprintf("https://localhost:%d", cfg.Server.Port)
		}
	}
	caPath := *ca
	if caPath == "" {
		caPath = cfg.Server.TLS.CertFile
	}

	// Handle Ctrl+C to exit cleanly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		trayui.Quit()
	}()

	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Stop(shutdownCtx)
		mgr.StopAll(3 * time.Second)
	}

	if err := trayui.Run(baseURL, token, caPath, cleanup); err != nil {
		log.Fatalf("tray error: %v", err)
	}
}
