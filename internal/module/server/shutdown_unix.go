//go:build !windows

package server

import (
	"os"
	"syscall"
)

// selfShutdown raises SIGTERM against our own PID so gex's signal handler
// (signal.NotifyContext on SIGINT/SIGTERM) runs every OnShutdown hook and
// then shuts the HTTP server down gracefully.
func selfShutdown() error {
	return syscall.Kill(os.Getpid(), syscall.SIGTERM)
}
