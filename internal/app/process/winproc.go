//go:build windows
// +build windows

package process

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// defaultSysProcAttr hides console windows for child processes.
func defaultSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
}
