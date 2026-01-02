package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Setup configures logging to a rotating daily file plus stderr.
func Setup(dir string) (*log.Logger, *os.File, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("create log dir: %w", err)
	}
	filename := filepath.Join(dir, fmt.Sprintf("service-%s.log", time.Now().Format("20060102")))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	logger := log.New(io.MultiWriter(os.Stderr, file), "", log.LstdFlags|log.Lmicroseconds)
	return logger, file, nil
}
