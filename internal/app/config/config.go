package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration for the application.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Programs []ProgramEntry `yaml:"programs"`
}

type ServerConfig struct {
	Bind          string         `yaml:"bind"`
	Port          int            `yaml:"port"`
	PublicBaseURL string         `yaml:"publicBaseUrl"`
	TLS           TLSConfig      `yaml:"tls"`
	Auth          AuthConfig     `yaml:"auth"`
	Network       NetworkConfig  `yaml:"network"`
	ReadTimeout   time.Duration  `yaml:"readTimeout"`
	WriteTimeout  time.Duration  `yaml:"writeTimeout"`
	IdleTimeout   time.Duration  `yaml:"idleTimeout"`
}

type TLSConfig struct {
	Enabled     bool   `yaml:"enabled"`
	CertFile    string `yaml:"certFile"`
	KeyFile     string `yaml:"keyFile"`
	ClientCA    string `yaml:"clientCaFile"`
	ForceClient bool   `yaml:"forceClientCert"`
}

type AuthConfig struct {
	BearerTokenFile string `yaml:"bearerTokenFile"`
}

type NetworkConfig struct {
	AllowedCIDRs []string `yaml:"allowedCIDRs"`
}

type ProgramEntry struct {
	Name      string            `yaml:"name"`
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	WorkDir   string            `yaml:"workDir"`
	Env       map[string]string `yaml:"env"`
	Stop      StopConfig        `yaml:"stop"`
	AutoStart bool              `yaml:"autoStart"`
}

type StopConfig struct {
	Method         string   `yaml:"method"` // jobobject|command
	TimeoutSeconds int      `yaml:"timeoutSeconds"`
	Command        string   `yaml:"command"`
	Args           []string `yaml:"args"`
}

// Load reads a config file from disk.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate performs minimal validation of configuration.
func (c *Config) Validate() error {
	if c.Server.Port == 0 {
		return errors.New("server.port is required")
	}
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" || c.Server.TLS.KeyFile == "" {
			return errors.New("tls.enabled true requires certFile and keyFile")
		}
	}
	if len(c.Programs) == 0 {
		return errors.New("no programs configured")
	}
	for _, p := range c.Programs {
		if p.Name == "" {
			return errors.New("program name is required")
		}
		if p.Command == "" {
			return fmt.Errorf("program %s missing command", p.Name)
		}
	}
	return nil
}
