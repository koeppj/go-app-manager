//go:build windows
// +build windows

package installer

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	ServiceName        = "GoAppManager"
	ServiceDisplayName = "Go App Manager Service"
)

// InstallService installs the Windows service.
func InstallService(exePath string, args []string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect service manager: %w", err)
	}
	defer m.Disconnect()

	if s, err := m.OpenService(ServiceName); err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", ServiceName)
	}

	config := mgr.Config{
		DisplayName: ServiceDisplayName,
		StartType:   mgr.StartAutomatic,
	}
	s, err := m.CreateService(ServiceName, exePath, config, args...)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	defer s.Close()
	return nil
}

// UninstallService removes the Windows service.
func UninstallService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("open service: %w", err)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// StartService requests start.
func StartService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(ServiceName)
	if err != nil {
		return err
	}
	defer s.Close()
	return s.Start()
}

// StopService requests stop and waits.
func StopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(ServiceName)
	if err != nil {
		return err
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	if status.State == svc.Stopped {
		return nil
	}
	// Give SCM a moment; no strict wait to avoid blocking.
	return nil
}

// QuoteArgs returns quoted arguments for use in scripts.
func QuoteArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, a := range args {
		if strings.ContainsAny(a, " ") {
			quoted = append(quoted, fmt.Sprintf("\"%s\"", a))
		} else {
			quoted = append(quoted, a)
		}
	}
	return strings.Join(quoted, " ")
}
