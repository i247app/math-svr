package logger

import (
	"io"
	"log/slog"

	"math-ai.com/math-ai/internal/shared/colors"
)

// Format selects the on-the-wire log encoding.
type Format string

const (
	// FormatText is the human-readable prefixed line. Used for local console
	// reading (optionally colored).
	FormatText Format = "text"
	// FormatJSON is one JSON object per line. Used when a log shipper
	// (Grafana Alloy → Loki) tails the file and parses fields.
	FormatJSON Format = "json"
)

// Options control how an AppLogger writes. An AppLogger can fan out to two
// destinations with independent encodings: a console (default stdout) and an
// optional file. The common local setup is console=text (easy to read) +
// file=json (shipped to Loki).
//
// The zero value is usable: it logs at Info to stdout in text.
type Options struct {
	// Level is the minimum level that gets written. nil → Info.
	Level slog.Leveler

	// ConsoleWriter is the console destination. nil → os.Stdout (unless a
	// FileWriter is set and you want file-only, in which case set this to
	// io.Discard). ConsoleFormat selects its encoding ("" → text).
	ConsoleWriter io.Writer
	ConsoleFormat Format

	// FileWriter is an optional second destination (the persistent log file).
	// nil → no file output. FileFormat selects its encoding ("" → json).
	FileWriter io.Writer
	FileFormat Format

	// BackgroundColor / WithColor apply to the TEXT console only. When any
	// destination is JSON, inline SQL coloring is force-disabled so the JSON
	// `msg` field stays free of ANSI escape codes (whole-line level color on
	// a text console is unaffected — it is applied per destination).
	BackgroundColor colors.Color
	WithColor       bool
}
