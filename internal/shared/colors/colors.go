package colors

import (
	"fmt"
)

type Color int

// ANSI color constants for foreground
const (
	FGBlack   Color = 30
	FGRed     Color = 31
	FGGreen   Color = 32
	FGYellow  Color = 33
	FGBlue    Color = 34
	FGMagenta Color = 35
	FGCyan    Color = 36
	FGWhite   Color = 37
	FGOrange  Color = 91
	FGPurple  Color = 35
	FGReset   Color = 0
)

// ANSI color constants for background
const (
	BGBlack   Color = 40
	BGRed     Color = 41
	BGGreen   Color = 42
	BGYellow  Color = 43
	BGBlue    Color = 44
	BGMagenta Color = 45
	BGCyan    Color = 46
	BGWhite   Color = 47
	BGOrange  Color = 101
	BGPurple  Color = 45
	BGReset   Color = 49
)

// colorizeString creates a colored string using ANSI escape codes
func Colorize(foregroundColor, backgroundColor Color, text string) string {
	// Validate color codes (basic validation)
	if foregroundColor < 0 || foregroundColor > 107 {
		foregroundColor = FGReset
	}
	if backgroundColor < 0 || backgroundColor > 107 {
		backgroundColor = BGReset
	}

	return fmt.Sprintf("\x1b[%d;%dm%s\x1b[0m", foregroundColor, backgroundColor, text)
}

func Bold(text string) string {
	return fmt.Sprintf("\x1b[1m%s\x1b[0m", text)
}
