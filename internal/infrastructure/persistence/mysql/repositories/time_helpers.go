package repositories

import "time"

// nullableTime maps the zero time to SQL NULL. Domain entities hold a zero
// time.Time for an unset nullable column (e.g. a journey that was never
// submitted to), and it must be written as NULL, not as year one.
func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// timeOrZero maps a NULL column (nil) to the zero time the domain uses for
// "unset".
func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
