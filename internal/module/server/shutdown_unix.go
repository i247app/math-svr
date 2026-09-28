//go:build !windows

package server

import (
	"fmt"
	"os"
	"syscall"
)

const reloadSupported = true

// selfShutdown raises SIGTERM against our own PID so gex's signal handler
// (signal.NotifyContext on SIGINT/SIGTERM) runs every OnShutdown hook and
// then shuts the HTTP server down gracefully.
func selfShutdown() error {
	return syscall.Kill(os.Getpid(), syscall.SIGTERM)
}

// Reexec replaces the process image with a fresh copy of the binary, keeping
// the PID (systemd's Type=simple unit sees no restart) and the original
// argv, under the start-up environment so the new process reads .env afresh.
// Go opens every fd close-on-exec, so no listener or log file leaks across.
// It only returns on failure.
func Reexec() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	if err := syscall.Exec(exe, os.Args, baseEnviron); err != nil {
		return fmt.Errorf("exec %s: %w", exe, err)
	}
	return nil
}
