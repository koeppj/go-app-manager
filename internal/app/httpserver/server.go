package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/koeppj/go-app-manager/internal/app/config"
	"github.com/koeppj/go-app-manager/internal/app/process"
	"github.com/koeppj/go-app-manager/internal/app/security"
)

// Server holds HTTP server state.
type Server struct {
	cfg    *config.Config
	logger *log.Logger
	mgr    *process.Manager
	token  string
	cidr   *security.CIDRAllowlist

	router *mux.Router
	srv    *http.Server
}

// New constructs a Server.
func New(cfg *config.Config, logger *log.Logger, mgr *process.Manager, token string) (*Server, error) {
	cidr, err := security.NewCIDRAllowlist(cfg.Server.Network.AllowedCIDRs)
	if err != nil {
		return nil, err
	}
	s := &Server{
		cfg:    cfg,
		logger: logger,
		mgr:    mgr,
		token:  token,
		cidr:   cidr,
	}
	s.router = mux.NewRouter()
	s.registerRoutes()

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Bind, cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	return s, nil
}

// Start begins serving HTTPS requests.
func (s *Server) Start() error {
	if !s.cfg.Server.TLS.Enabled {
		return fmt.Errorf("tls must be enabled")
	}
	tlsCfg, err := security.BuildTLSConfig(s.cfg.Server.TLS)
	if err != nil {
		return err
	}
	s.srv.TLSConfig = tlsCfg

	s.logger.Printf("serving on https://%s", s.srv.Addr)
	go func() {
		if err := s.srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("server error: %v", err)
		}
	}()
	return nil
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
