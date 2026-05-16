package logger

import (
	"io"
	"log/slog"

	"math-ai.com/math-ai/internal/shared/colors"
)

// Options control how an AppLogger writes log lines. The zero value is
// usable: it logs at Info level to os.Stdout with no color. Build it once
// at boot and pass the same Options to the middleware on every request.
type Options struct {
	// Level is the minimum level that gets written. nil → Info.
	Level slog.Leveler
	// Output is the destination writer. nil → os.Stdout.
	Output io.Writer
	// BackgroundColor is the ANSI background applied to every line. Color
	// is enabled only when WithColor is true; the foreground is chosen per
	// log level (Error=red, Warn=yellow, Info=green, Debug=cyan).
	BackgroundColor colors.Color
	// WithColor toggles ANSI colorisation. Off by default — production log
	// shippers typically prefer plain text. Turn on for local dev.
	WithColor bool
}
