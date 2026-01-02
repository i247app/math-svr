package logger

// ANSI color codes for background colors
// These codes wrap text with background colors in terminal output

type BackgroundColor string

const (
	// No background color (transparent/default)
	BgNone BackgroundColor = ""

	// Standard background colors
	BgBlack   BackgroundColor = "\033[40m"
	BgRed     BackgroundColor = "\033[41m"
	BgGreen   BackgroundColor = "\033[42m"
	BgYellow  BackgroundColor = "\033[43m"
	BgBlue    BackgroundColor = "\033[44m"
	BgMagenta BackgroundColor = "\033[45m"
	BgCyan    BackgroundColor = "\033[46m"
	BgWhite   BackgroundColor = "\033[47m"

	// Bright background colors
	BgBrightBlack   BackgroundColor = "\033[100m"
	BgBrightRed     BackgroundColor = "\033[101m"
	BgBrightGreen   BackgroundColor = "\033[102m"
	BgBrightYellow  BackgroundColor = "\033[103m"
	BgBrightBlue    BackgroundColor = "\033[104m"
	BgBrightMagenta BackgroundColor = "\033[105m"
	BgBrightCyan    BackgroundColor = "\033[106m"
	BgBrightOrange  BackgroundColor = "\033[48;5;214m"
	BgBrightWhite   BackgroundColor = "\033[107m"

	// Reset code to clear all formatting
	colorReset = "\033[0m"
)

// applyBackgroundColor wraps text with the background color ANSI codes
func applyBackgroundColor(text string, bgColor BackgroundColor) string {
	if bgColor == "" || bgColor == BgNone {
		// No background color, return text as-is
		return text
	}
	// Apply background color only to the text content, then reset
	return string(bgColor) + text + colorReset
}
