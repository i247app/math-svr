package job

import (
	"errors"
	"testing"
	"time"
)

func mustLoad(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Skipf("timezone %s unavailable on this host: %v", name, err)
	}
	return loc
}

func TestScheduleValidate(t *testing.T) {
	la := mustLoad(t, "America/Los_Angeles")

	tests := []struct {
		name    string
		sched   Schedule
		wantErr bool
	}{
		{"every positive", EveryDuration(15 * time.Minute), false},
		{"every zero", EveryDuration(0), true},
		{"every negative", EveryDuration(-time.Second), true},

		{"daily valid", DailyAt(0, 0, la), false},
		{"daily end of range", DailyAt(23, 59, la), false},
		{"daily nil location", DailyAt(6, 30, nil), false},
		{"daily hour too large", DailyAt(24, 0, la), true},
		{"daily hour negative", DailyAt(-1, 0, la), true},
		{"daily minute too large", DailyAt(12, 60, la), true},
		{"daily minute negative", DailyAt(12, -1, la), true},

		{"weekly valid sunday", WeeklyAt(time.Sunday, 8, 0, la), false},
		{"weekly valid saturday", WeeklyAt(time.Saturday, 8, 0, la), false},
		{"weekly weekday too large", WeeklyAt(time.Weekday(7), 8, 0, la), true},
		{"weekly weekday negative", WeeklyAt(time.Weekday(-1), 8, 0, la), true},
		{"weekly bad hour", WeeklyAt(time.Monday, 99, 0, la), true},

		{"zero value schedule", Schedule{}, true},
		{"unknown kind", Schedule{Kind: ScheduleKind(99)}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sched.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Validate() = nil, want error")
				}
				if !errors.Is(err, ErrInvalidSchedule) {
					t.Fatalf("Validate() error = %v, want it to wrap ErrInvalidSchedule", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

// TestValidateAgreesWithNext locks in the relationship the cron loop
// depends on: anything Validate accepts must produce a usable fire time,
// so a validated schedule can never park a scheduler loop.
func TestValidateAgreesWithNext(t *testing.T) {
	la := mustLoad(t, "America/Los_Angeles")
	now := time.Now()

	valid := []Schedule{
		EveryDuration(time.Minute),
		DailyAt(0, 0, la),
		DailyAt(23, 59, nil),
		WeeklyAt(time.Sunday, 0, 0, la),
		WeeklyAt(time.Saturday, 23, 59, la),
	}
	for _, s := range valid {
		if err := s.Validate(); err != nil {
			t.Fatalf("Validate(%s) = %v, want nil", s, err)
		}
		if next := s.Next(now); next.IsZero() {
			t.Fatalf("Next(%s) returned the zero Time for a schedule Validate accepted", s)
		}
	}
}

// TestScheduleNextResolvesDateInScheduleLocation is the regression test
// for the cross-date-boundary bug: Next used to read Year/Month/Day off
// `now` in NOW's location instead of the schedule's. A server running in
// Asia/Ho_Chi_Minh computing an America/Los_Angeles schedule therefore
// picked up Vietnam's calendar date — which, once VN has rolled past
// midnight while LA has not, is one day ahead — and scheduled the job a
// full 24 hours late.
func TestScheduleNextResolvesDateInScheduleLocation(t *testing.T) {
	hcm := mustLoad(t, "Asia/Ho_Chi_Minh")
	la := mustLoad(t, "America/Los_Angeles")
	tokyo := mustLoad(t, "Asia/Tokyo")

	// 2026-06-07 19:52 UTC is 2026-06-08 02:52 in Vietnam (already
	// tomorrow) but still 2026-06-07 12:52 in Los Angeles.
	instant := time.Date(2026, time.June, 7, 19, 52, 0, 0, time.UTC)

	tests := []struct {
		name  string
		now   time.Time
		sched Schedule
		want  time.Time
	}{
		{
			// The reported failure: 2 minutes out, observed from a VN server.
			name:  "LA schedule seen from a Vietnam-local now",
			now:   instant.In(hcm),
			sched: DailyAt(12, 54, la),
			want:  time.Date(2026, time.June, 7, 12, 54, 0, 0, la),
		},
		{
			// Same instant, same expectation, regardless of now's location.
			name:  "LA schedule seen from a UTC now",
			now:   instant,
			sched: DailyAt(12, 54, la),
			want:  time.Date(2026, time.June, 7, 12, 54, 0, 0, la),
		},
		{
			// Time already passed in LA today → roll to tomorrow, in LA.
			name:  "LA schedule already past today",
			now:   instant.In(hcm),
			sched: DailyAt(12, 50, la),
			want:  time.Date(2026, time.June, 8, 12, 50, 0, 0, la),
		},
		{
			// The cases that worked before the fix must keep working.
			name:  "Vietnam schedule seen from a Vietnam-local now",
			now:   instant.In(hcm),
			sched: DailyAt(2, 54, hcm),
			want:  time.Date(2026, time.June, 8, 2, 54, 0, 0, hcm),
		},
		{
			name:  "Tokyo schedule seen from a Vietnam-local now",
			now:   instant.In(hcm),
			sched: DailyAt(4, 54, tokyo),
			want:  time.Date(2026, time.June, 8, 4, 54, 0, 0, tokyo),
		},
		{
			// Weekly reads the same date components and had the same bug.
			// 2026-06-07 is a Sunday in LA, still Monday-eve in Vietnam.
			name:  "weekly LA schedule seen from a Vietnam-local now",
			now:   instant.In(hcm),
			sched: WeeklyAt(time.Sunday, 12, 54, la),
			want:  time.Date(2026, time.June, 7, 12, 54, 0, 0, la),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.sched.Next(tc.now)
			if !got.Equal(tc.want) {
				t.Fatalf("Next(%s) = %s (%s UTC), want %s (%s UTC)",
					tc.now, got, got.UTC(), tc.want, tc.want.UTC())
			}
			if !got.After(tc.now) {
				t.Fatalf("Next(%s) = %s is not strictly after now", tc.now, got)
			}
		})
	}
}
