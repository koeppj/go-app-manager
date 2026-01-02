package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/koeppj/go-app-manager/internal/app/config"
)

// BuildTLSConfig constructs a tls.Config from settings.
func BuildTLSConfig(cfg config.TLSConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load tls cert: %w", err)
	}
	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}
	if cfg.ClientCA != "" {
		caData, err := os.ReadFile(cfg.ClientCA)
		if err != nil {
			return nil, fmt.Errorf("read client ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("client ca not PEM")
		}
		tlsCfg.ClientCAs = pool
		if cfg.ForceClient {
			tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsCfg.ClientAuth = tls.VerifyClientCertIfGiven
		}
	}
	return tlsCfg, nil
}
