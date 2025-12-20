package time

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// DefaultFormat is the default time format used for JSON marshaling/unmarshaling
// RFC3339 is used to maintain compatibility with existing API responses
const DefaultFormat = time.RFC3339Nano

// MathTime is a custom time type that wraps time.Time with custom JSON and database marshaling
type MathTime struct {
	time.Time
}

// Now returns a new MathTime with the current time
func Now() MathTime {
	return MathTime{Time: time.Now()}
}

// NewMathTime creates a new MathTime from a standard time.Time
func NewMathTime(t time.Time) MathTime {
	return MathTime{Time: t}
}

// MarshalJSON implements the json.Marshaler interface
func (mt MathTime) MarshalJSON() ([]byte, error) {
	if mt.IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf("\"%s\"", mt.Format(DefaultFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (mt *MathTime) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" || string(data) == `""` {
		*mt = MathTime{}
		return nil
	}

	// Remove quotes
	str := string(data)
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// Parse the time string
	t, err := time.Parse(DefaultFormat, str)
	if err != nil {
		return fmt.Errorf("failed to parse time: %w", err)
	}

	mt.Time = t
	return nil
}

// Value implements the driver.Valuer interface for database writes
func (mt MathTime) Value() (driver.Value, error) {
	if mt.IsZero() {
		return nil, nil
	}
	return mt.Time, nil
}

// Scan implements the sql.Scanner interface for database reads
func (mt *MathTime) Scan(value interface{}) error {
	if value == nil {
		*mt = MathTime{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		mt.Time = v
		return nil
	case []byte:
		t, err := time.Parse(DefaultFormat, string(v))
		if err != nil {
			return fmt.Errorf("failed to scan time: %w", err)
		}
		mt.Time = t
		return nil
	case string:
		t, err := time.Parse(DefaultFormat, v)
		if err != nil {
			return fmt.Errorf("failed to scan time: %w", err)
		}
		mt.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into MathTime", value)
	}
}

// String returns the time formatted with the default format
func (mt MathTime) String() string {
	if mt.IsZero() {
		return ""
	}
	return mt.Format(DefaultFormat)
}

// ToTime returns the underlying time.Time value
func (mt MathTime) ToTime() time.Time {
	return mt.Time
}

// IsValid checks if the MathTime contains a valid non-zero time
func (mt MathTime) IsValid() bool {
	return !mt.IsZero()
}

// Ptr returns a pointer to the MathTime
// Useful for converting to *MathTime for optional fields
func (mt MathTime) Ptr() *MathTime {
	return &mt
}
