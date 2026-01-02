package app

import (
	"os"
	"path/filepath"
)

const (
	defaultCompanyDir = "GoAppManager"
	configDirName     = "config"
	logDirName        = "logs"
	secretsDirName    = "secrets"
	certsDirName      = "certs"
)

// ProgramDataDir returns the base directory under ProgramData for the app.
func ProgramDataDir() string {
	if dir := os.Getenv("PROGRAMDATA"); dir != "" {
		return filepath.Join(dir, defaultCompanyDir)
	}
	return filepath.Join(`C:\ProgramData`, defaultCompanyDir)
}

// ConfigPath returns the default config file location.
func ConfigPath() string {
	return filepath.Join(ProgramDataDir(), configDirName, "config.yaml")
}

// LogDir returns the directory for log files.
func LogDir() string {
	return filepath.Join(ProgramDataDir(), logDirName)
}

// SecretsDir returns the directory for secrets.
func SecretsDir() string {
	return filepath.Join(ProgramDataDir(), secretsDirName)
}

// CertsDir returns the directory for TLS certs.
func CertsDir() string {
	return filepath.Join(ProgramDataDir(), certsDirName)
}
