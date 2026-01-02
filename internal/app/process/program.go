//go:build windows
// +build windows

package process

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/koeppj/go-app-manager/internal/app/config"
	"golang.org/x/sys/windows"
)

// Program represents a managed executable.
type Program struct {
	cfg     config.ProgramEntry
	logger  *log.Logger

	cmd     *exec.Cmd
	job     *JobObject
	status  ProgramStatus
	waitCh  chan struct{}
	mu      sync.Mutex
}

// ProgramStatus captures runtime status.
type ProgramStatus struct {
	Name         string    `json:"name"`
	Running      bool      `json:"running"`
	PID          int       `json:"pid"`
	StartTime    time.Time `json:"startTime"`
	LastExitCode int       `json:"lastExitCode"`
	LastError    string    `json:"lastError"`
}

func NewProgram(cfg config.ProgramEntry, logger *log.Logger) *Program {
	stopCfg := cfg.Stop
	if stopCfg.Method == "" {
		stopCfg = config.DefaultStopConfig()
		cfg.Stop = stopCfg
	}
	return &Program{
		cfg:    cfg,
		logger: logger,
		status: ProgramStatus{Name: cfg.Name},
	}
}

// Start launches the program and assigns it to a Job Object.
func (p *Program) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status.Running {
		return nil
	}

	cmd := exec.Command(p.cfg.Command, p.cfg.Args...)
	cmd.Dir = p.cfg.WorkDir
	cmd.SysProcAttr = defaultSysProcAttr()
	cmd.Env = mergeEnv(p.cfg.Env)
	if err := cmd.Start(); err != nil {
		p.status.LastError = err.Error()
		return fmt.Errorf("start %s: %w", p.cfg.Name, err)
	}

	access := uint32(windows.PROCESS_SET_QUOTA | windows.PROCESS_TERMINATE | windows.PROCESS_QUERY_INFORMATION)
	procHandle, err := windows.OpenProcess(access, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("open process handle: %w", err)
	}

	job, err := NewJobObject(p.cfg.Name)
	if err != nil {
		windows.CloseHandle(procHandle)
		_ = cmd.Process.Kill()
		return err
	}
	if err := job.Assign(procHandle); err != nil {
		windows.CloseHandle(procHandle)
		_ = job.Close()
		_ = cmd.Process.Kill()
		return err
	}
	windows.CloseHandle(procHandle)

	waitCh := make(chan struct{})
	p.cmd = cmd
	p.job = job
	p.waitCh = waitCh
	p.status.Running = true
	p.status.PID = cmd.Process.Pid
	p.status.StartTime = time.Now()
	p.status.LastError = ""

	go p.waitForExit(cmd, job, waitCh)
	return nil
}

func (p *Program) waitForExit(cmd *exec.Cmd, job *JobObject, waitCh chan struct{}) {
	err := cmd.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()
	defer close(waitCh)
	if job != nil {
		_ = job.Close()
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			p.status.LastExitCode = status.ExitStatus()
		}
		p.status.LastError = exitErr.Error()
	} else if err != nil {
		p.status.LastError = err.Error()
	} else {
		p.status.LastExitCode = 0
		p.status.LastError = ""
	}
	p.status.Running = false
	p.status.PID = 0
	p.cmd = nil
	p.job = nil
}

// Stop terminates the program using the configured strategy.
func (p *Program) Stop(timeout time.Duration) error {
	p.mu.Lock()
	if !p.status.Running || p.cmd == nil {
		p.mu.Unlock()
		return nil
	}
	waitCh := p.waitCh
	stopCfg := p.cfg.Stop
	job := p.job
	p.mu.Unlock()

	if stopCfg.TimeoutSeconds > 0 {
		timeout = time.Duration(stopCfg.TimeoutSeconds) * time.Second
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	if stopCfg.Method == "command" && stopCfg.Command != "" {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		stopCmd := exec.CommandContext(ctx, stopCfg.Command, stopCfg.Args...)
		stopCmd.Dir = p.cfg.WorkDir
		stopCmd.SysProcAttr = defaultSysProcAttr()
		stopCmd.Env = mergeEnv(p.cfg.Env)
		_ = stopCmd.Run()
	}

	if job != nil {
		if err := job.Terminate(1); err != nil {
			p.logger.Printf("terminate job %s failed: %v", p.cfg.Name, err)
		}
	}

	select {
	case <-waitCh:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout waiting for process to exit")
	}
}

// Restart performs stop then start.
func (p *Program) Restart() error {
	if err := p.Stop(0); err != nil {
		return err
	}
	return p.Start()
}

func (p *Program) Status() ProgramStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}

func mergeEnv(custom map[string]string) []string {
	env := os.Environ()
	for k, v := range custom {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}
