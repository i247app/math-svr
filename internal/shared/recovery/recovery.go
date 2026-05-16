package recovery

import (
	"fmt"
	"runtime"
	"strings"
)

// PanicInfo contains information about a panic
type PanicInfo struct {
	Value      interface{}
	Goroutine  int
	StackTrace []StackFrame
}

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	File     string
	Line     int
	Function string
	PC       uintptr
}

// RecoverPanic recovers from a panic and returns panic information
// Returns nil if there was no panic
// NOTE: This must be called directly in a deferred function, not through another helper
func RecoverPanic() *PanicInfo {
	if r := recover(); r != nil {
		return &PanicInfo{
			Value:      r,
			Goroutine:  getGoroutineID(),
			StackTrace: captureStackTrace(3), // Skip 3 frames: runtime.Callers, captureStackTrace, RecoverPanic
		}
	}
	return nil
}

// BuildPanicInfo builds PanicInfo from a recovered panic value
// Use this when you've already called recover() directly
// skip indicates how many stack frames to skip (use 3 for middleware)
func BuildPanicInfo(panicValue interface{}, skip int) *PanicInfo {
	return &PanicInfo{
		Value:      panicValue,
		Goroutine:  getGoroutineID(),
		StackTrace: captureStackTrace(skip),
	}
}

// captureStackTrace captures the current stack trace
// skip indicates how many stack frames to skip
func captureStackTrace(skip int) []StackFrame {
	const maxStackDepth = 32
	var pcs [maxStackDepth]uintptr
	n := runtime.Callers(skip, pcs[:])

	frames := make([]StackFrame, 0, n)
	callersFrames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := callersFrames.Next()

		// Filter out runtime frames for cleaner output
		if !strings.Contains(frame.Function, "runtime.") {
			frames = append(frames, StackFrame{
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Function,
				PC:       frame.PC,
			})
		}

		if !more {
			break
		}
	}

	return frames
}

// getGoroutineID returns the current goroutine ID
func getGoroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Parse "goroutine 123 [running]:"
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	var id int
	fmt.Sscanf(idField, "%d", &id)
	return id
}

// Format returns a nicely formatted panic message
func (p *PanicInfo) Format() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔═══════════════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                            🚨 PANIC RECOVERED 🚨                              ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	// Panic type and message
	panicType := fmt.Sprintf("%T", p.Value)
	sb.WriteString(fmt.Sprintf("💥 Panic Type: %s\n", panicType))
	sb.WriteString(fmt.Sprintf("📝 Message: %v\n", p.Value))
	sb.WriteString(fmt.Sprintf("🔢 Goroutine: %d\n", p.Goroutine))
	sb.WriteString("\n")

	// Stack trace
	sb.WriteString("📍 Stack Trace:\n")
	sb.WriteString("───────────────────────────────────────────────────────────────────────────────\n")

	for i, frame := range p.StackTrace {
		// Extract just the filename from the full path
		fileParts := strings.Split(frame.File, "/")
		fileName := frame.File
		if len(fileParts) > 0 {
			// Show last 2 parts of path for context
			if len(fileParts) >= 2 {
				fileName = fileParts[len(fileParts)-2] + "/" + fileParts[len(fileParts)-1]
			} else {
				fileName = fileParts[len(fileParts)-1]
			}
		}

		// Extract function name (remove package path)
		funcParts := strings.Split(frame.Function, "/")
		funcName := frame.Function
		if len(funcParts) > 0 {
			funcName = funcParts[len(funcParts)-1]
		}

		// Format with box drawing characters
		if i == 0 {
			sb.WriteString(fmt.Sprintf("  ┌─ [%d] %s\n", i+1, funcName))
		} else {
			sb.WriteString(fmt.Sprintf("  ├─ [%d] %s\n", i+1, funcName))
		}
		sb.WriteString(fmt.Sprintf("  │   📄 %s:%d\n", fileName, frame.Line))

		if i < len(p.StackTrace)-1 {
			sb.WriteString("  │\n")
		}
	}

	if len(p.StackTrace) > 0 {
		sb.WriteString("  └─────────────────────────────────────────────────────────────────────────\n")
	}

	return sb.String()
}

// FormatShort returns a short, one-line panic summary
func (p *PanicInfo) FormatShort() string {
	var location string
	if len(p.StackTrace) > 0 {
		frame := p.StackTrace[0]
		fileParts := strings.Split(frame.File, "/")
		fileName := fileParts[len(fileParts)-1]
		location = fmt.Sprintf("%s:%d", fileName, frame.Line)
	} else {
		location = "unknown"
	}

	return fmt.Sprintf("PANIC: %v at %s", p.Value, location)
}

// IsNilPointerError checks if the panic is a nil pointer error
func (p *PanicInfo) IsNilPointerError() bool {
	panicStr := fmt.Sprintf("%v", p.Value)
	return strings.Contains(panicStr, "nil pointer") ||
		strings.Contains(panicStr, "invalid memory") ||
		strings.Contains(panicStr, "nil pointer dereference")
}

// GetOriginatingLocation returns the file:line where the panic originated
func (p *PanicInfo) GetOriginatingLocation() string {
	if len(p.StackTrace) > 0 {
		frame := p.StackTrace[0]
		return fmt.Sprintf("%s:%d", frame.File, frame.Line)
	}
	return "unknown location"
}
