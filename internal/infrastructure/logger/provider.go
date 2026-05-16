package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Provider is the process-scoped owner of the logger's underlying I/O —
// principally the optional persistent log file. Construct one at boot via
// NewProvider, hand it to the middleware, and call Close() on shutdown.
//
// AppLoggers spawned via Provider.New / Provider.Background share the same
// io.Writer (stdout tee'd with the log file). They never own the file
// handle themselves — closing AppLogger after a single request would yank
// the file out from under every later request.
type Provider struct {
	opts    Options
	closers []io.Closer
}

// NewProvider builds a Provider whose AppLoggers write to stdout AND, when
// logFile is non-empty, to that file (append mode, file is created if
// missing along with any parent directories). Color, level, and other
// formatting choices come from opts.
//
// Returned Provider must have Close() called on shutdown to flush and
// release the file handle.
func NewProvider(opts Options, logFile string) (*Provider, error) {
	p := &Provider{opts: opts}

	writers := []io.Writer{os.Stdout}
	if logFile != "" {
		f, err := openLogFile(logFile)
		if err != nil {
			return nil, fmt.Errorf("logger: open log file %q: %w", logFile, err)
		}
		writers = append(writers, f)
		p.closers = append(p.closers, f)
	}

	p.opts.Output = io.MultiWriter(writers...)
	return p, nil
}

// New returns a per-request AppLogger bound to ctx and r. The middleware is
// the canonical caller — handlers retrieve the result via logger.From(ctx).
func (p *Provider) New(ctx context.Context, r *http.Request) *AppLogger {
	return New(ctx, r, p.opts)
}

// Background returns an AppLogger detached from any HTTP request, suitable
// for goroutines whose lifetime exceeds the request that spawned them
// (NATS publishers, scheduled jobs, shutdown hooks).
func (p *Provider) Background(ctx context.Context) *AppLogger {
	return NewBackground(ctx, p.opts)
}

// Close releases every file handle the Provider opened. Stdout and stderr
// are intentionally never closed. Errors are joined so one bad closer does
// not mask the rest. If this Provider is the registered default
// (SetDefault), the registration is cleared so callers don't fall back to
// a closed file handle.
func (p *Provider) Close() error {
	if Default() == p {
		SetDefault(nil)
	}

	var errs []error
	for _, c := range p.closers {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	p.closers = nil
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// openLogFile opens path for append-only writes, creating parent
// directories as needed. Permissions follow the standard log file mode
// (0644 for the file, 0755 for any directory we create).
func openLogFile(path string) (*os.File, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir %q: %w", dir, err)
		}
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}
