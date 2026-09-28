//go:build windows

package server

import (
	"errors"
	"os"
)

// reloadSupported is false: Windows has no exec(2), and selfShutdown here
// exits without running the shutdown hooks. HandleReload refuses up front.
const reloadSupported = false

// selfShutdown on Windows: there are no POSIX signals, and os.Process.Signal
// cannot deliver SIGINT/SIGTERM to our own process, so we cannot drive gex's
// graceful-shutdown path the way Unix does. Windows is build/dev-only for this
// server, so we just exit. This skips the OnShutdown hooks (JobRuntime drain +
// session serialization) — acceptable for local development only.
func selfShutdown() error {
	os.Exit(0)
	return nil
}

// Reexec is unreachable on Windows (HandleReload refuses first).
func Reexec() error {
	return errors.New("reload is not supported on windows")
}
