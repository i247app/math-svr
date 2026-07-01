package time

import (
	"fmt"
	"time"
)

// allowedLayouts defines all supported date/time formats for parsing
var allowedLayouts = []string{
	DateTimeFormat20FSP, // "2006-01-02 15:04:05.000000" - MySQL DATETIME(6)
	CoreDateTimeFormat,  // "2006-01-02T15:04:05Z" - Core API (ISO 8601)
	DefaultFormat,       // "2006-01-02T15:04:05Z07:00" - RFC3339
	Format20FSP,         // "20060102150405.000000" - Compact with microseconds
	Format17FSP,         // "20060102150405.000" - Compact with milliseconds
	DateOnly,            // "2006-01-02" - Date only
}

// ParseWithFormat parses a time string using a custom format.
// If the provided layout fails, it tries all allowed layouts as fallbacks.
func ParseWithFormat(layout, value string) (MathTime, error) {
	if value == "" {
		return MathTime{}, nil
	}

	// Try the provided layout first
	t, err := time.Parse(layout, value)
	if err == nil {
		return MathTime{Time: t}, nil
	}

	// Try all allowed layouts as fallbacks
	for _, l := range allowedLayouts {
		if l == layout {
			continue // Skip the layout we already tried
		}
		t, err := time.Parse(l, value)
		if err == nil {
			return MathTime{Time: t}, nil
		}
	}

	return MathTime{}, err
}

func ParseFromString(value string) (MathTime, error) {
	if value == "" {
		return MathTime{}, nil
	}

	// Try all allowed layouts as fallbacks
	for _, l := range allowedLayouts {
		t, err := time.Parse(l, value)
		if err == nil {
			return MathTime{Time: t}, nil
		}
	}

	return MathTime{}, fmt.Errorf("failed to parse time: wrong format with %s or not in supported format", value)
}

// ParseInLocation parses a time string in the given location using the default format
func ParseInLocation(value string, loc *time.Location) (MathTime, error) {
	t, err := time.ParseInLocation(DefaultFormat, value, loc)
	if err != nil {
		return MathTime{}, err
	}
	return MathTime{Time: t}, nil
}

// ParseDate parses a date-only string (YYYY-MM-DD)
func ParseDate(value string) (MathTime, error) {
	t, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return MathTime{}, err
	}
	return MathTime{Time: t}, nil
}

// NewMathTimePtr creates a new *MathTime from a standard time.Time
// Returns nil if the time is zero
func NewMathTimePtr(t time.Time) *MathTime {
	if t.IsZero() {
		return nil
	}
	mt := MathTime{Time: t}
	return &mt
}

// MathTimeFromPtr converts *time.Time to MathTime
// Returns zero MathTime if pointer is nil
func MathTimeFromPtr(t *time.Time) MathTime {
	if t == nil {
		return MathTime{}
	}
	return MathTime{Time: *t}
}

// TimeTobinbaseTimePtr converts *time.Time to *binbaseTime
// Returns nil if pointer is nil
func TimeToMathTimePtr(t *time.Time) *MathTime {
	if t == nil {
		return nil
	}
	return NewMathTimePtr(*t)
}

// MathTimePtrToTime converts *MathTime to *time.Time
// Returns nil if pointer is nil
func MathTimePtrToTime(mt *MathTime) *time.Time {
	if mt == nil {
		return nil
	}
	return &mt.Time
}
