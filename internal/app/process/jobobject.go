//go:build windows
// +build windows

package process

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// JobObject wraps a Windows Job Object handle.
type JobObject struct {
	handle windows.Handle
}

// NewJobObject creates a new Job Object.
func NewJobObject(name string) (*JobObject, error) {
	var objName *uint16
	if name != "" {
		objName = windows.StringToUTF16Ptr(name)
	}
	h, err := windows.CreateJobObject(nil, objName)
	if err != nil {
		return nil, fmt.Errorf("create job object: %w", err)
	}
	return &JobObject{handle: h}, nil
}

// Assign attaches a process to the Job Object.
func (j *JobObject) Assign(process windows.Handle) error {
	if err := windows.AssignProcessToJobObject(j.handle, process); err != nil {
		return fmt.Errorf("assign process to job: %w", err)
	}
	return nil
}

// Terminate terminates all processes within the job.
func (j *JobObject) Terminate(exitCode uint32) error {
	if err := windows.TerminateJobObject(j.handle, exitCode); err != nil {
		return fmt.Errorf("terminate job: %w", err)
	}
	return nil
}

// Close releases the handle.
func (j *JobObject) Close() error {
	return windows.CloseHandle(j.handle)
}
