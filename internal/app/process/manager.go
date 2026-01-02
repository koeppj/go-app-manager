//go:build windows
// +build windows

package process

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/koeppj/go-app-manager/internal/app/config"
)

// Manager orchestrates configured programs.
type Manager struct {
	programs map[string]*Program
	logger   *log.Logger
	mu       sync.RWMutex
}

func NewManager(cfg *config.Config, logger *log.Logger) *Manager {
	progs := make(map[string]*Program)
	for _, p := range cfg.Programs {
		progs[p.Name] = NewProgram(p, logger)
	}
	return &Manager{programs: progs, logger: logger}
}

func (m *Manager) AutoStart() {
	for _, p := range m.programs {
		if p.cfg.AutoStart {
			if err := p.Start(); err != nil {
				m.logger.Printf("auto start %s failed: %v", p.cfg.Name, err)
			}
		}
	}
}

func (m *Manager) Start(name string) error {
	p, err := m.get(name)
	if err != nil {
		return err
	}
	return p.Start()
}

func (m *Manager) Stop(name string) error {
	p, err := m.get(name)
	if err != nil {
		return err
	}
	return p.Stop(0)
}

func (m *Manager) Restart(name string) error {
	p, err := m.get(name)
	if err != nil {
		return err
	}
	return p.Restart()
}

func (m *Manager) Status(name string) (ProgramStatus, error) {
	p, err := m.get(name)
	if err != nil {
		return ProgramStatus{}, err
	}
	return p.Status(), nil
}

func (m *Manager) StatusAll() []ProgramStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	results := make([]ProgramStatus, 0, len(m.programs))
	for _, p := range m.programs {
		results = append(results, p.Status())
	}
	return results
}

func (m *Manager) get(name string) (*Program, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.programs[name]
	if !ok {
		return nil, fmt.Errorf("program not found: %s", name)
	}
	return p, nil
}

// StopAll stops all programs with timeout to avoid blocking shutdown.
func (m *Manager) StopAll(timeout time.Duration) {
	var wg sync.WaitGroup
	for _, p := range m.programs {
		wg.Add(1)
		go func(pr *Program) {
			defer wg.Done()
			_ = pr.Stop(timeout)
		}(p)
	}
	wg.Wait()
}
