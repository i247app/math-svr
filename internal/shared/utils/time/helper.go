package time

import "time"

// Parse parses a time string using the default format (RFC3339)
func Parse(value string) (MathTime, error) {
	t, err := time.Parse(DefaultFormat, value)
	if err != nil {
		return MathTime{}, err
	}
	return MathTime{Time: t}, nil
}

// ParseInLocation parses a time string in the given location using the default format
func ParseInLocation(value string, loc *time.Location) (MathTime, error) {
	t, err := time.ParseInLocation(DefaultFormat, value, loc)
	if err != nil {
		return MathTime{}, err
	}
	return MathTime{Time: t}, nil
}

// ParseWithFormat parses a time string using a custom format
func ParseWithFormat(layout, value string) (MathTime, error) {
	t, err := time.Parse(layout, value)
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

// TimeToMathTimePtr converts *time.Time to *MathTime
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