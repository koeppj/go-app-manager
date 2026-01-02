package config

import (
	"time"

	"github.com/koeppj/go-app-manager/internal/app"
)

// DefaultConfig returns a Config populated with defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Bind:          "0.0.0.0",
			Port:          8443,
			PublicBaseURL: "https://localhost:8443",
			TLS: TLSConfig{
				Enabled:  true,
				CertFile: app.CertsDir() + `\server.crt`,
				KeyFile:  app.CertsDir() + `\server.key`,
			},
			Auth: AuthConfig{
				BearerTokenFile: app.SecretsDir() + `\api.token`,
			},
			Network: NetworkConfig{
				AllowedCIDRs: []string{},
			},
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Programs: []ProgramEntry{},
	}
}

// DefaultStopConfig returns sensible defaults for stop behavior.
func DefaultStopConfig() StopConfig {
	return StopConfig{
		Method:         "jobobject",
		TimeoutSeconds: 10,
	}
}
