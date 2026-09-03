package job

import (
	"context"
	"testing"
	"time"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func weekdayPtr(d int) *int { return &d }

// TestParseSchedule covers the wire contract of /jobs/schedule/update.
// want is the Schedule.String() the request must resolve to — asserting
// on the rendered form keeps the timezone visible in the expectation,
// which is the field most likely to be resolved wrongly.
func TestParseSchedule(t *testing.T) {
	tests := []struct {
		name string
		req  *ScheduleRequest
		want string
	}{
		{"every minute (the floor)", &ScheduleRequest{Kind: "every", EverySeconds: 60}, "every 1m0s"},
		{"every 15 minutes", &ScheduleRequest{Kind: "every", EverySeconds: 900}, "every 15m0s"},
		{"hourly", &ScheduleRequest{Kind: "every", EverySeconds: 3600}, "every 1h0m0s"},
		{"every 2 hours", &ScheduleRequest{Kind: "every", EverySeconds: 7200}, "every 2h0m0s"},

		{
			"daily with timezone",
			&ScheduleRequest{Kind: "daily", Hour: 6, Minute: 30, Timezone: "Asia/Ho_Chi_Minh"},
			"daily 06:30 Asia/Ho_Chi_Minh",
		},
		{
			"daily without timezone defaults to UTC",
			&ScheduleRequest{Kind: "daily", Hour: 6, Minute: 30},
			"daily 06:30 UTC",
		},
		{
			"daily midnight",
			&ScheduleRequest{Kind: "daily", Hour: 0, Minute: 0, Timezone: "America/Los_Angeles"},
			"daily 00:00 America/Los_Angeles",
		},
		{
			"weekly on Sunday (weekday 0 is not 'omitted')",
			&ScheduleRequest{Kind: "weekly", Weekday: weekdayPtr(0), Hour: 8, Timezone: "Asia/Tokyo"},
			"weekly Sunday 08:00 Asia/Tokyo",
		},
		{
			"weekly on Monday",
			&ScheduleRequest{Kind: "weekly", Weekday: weekdayPtr(1), Hour: 8, Minute: 15, Timezone: "Asia/Tokyo"},
			"weekly Monday 08:15 Asia/Tokyo",
		},

		{"kind is case-insensitive", &ScheduleRequest{Kind: "  DAILY ", Hour: 7}, "daily 07:00 UTC"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSchedule(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("parseSchedule() = %v, want nil", err)
			}
			if got.String() != tc.want {
				t.Fatalf("parseSchedule() = %q, want %q", got.String(), tc.want)
			}
			if err := got.Validate(); err != nil {
				t.Fatalf("parseSchedule produced a schedule the runtime rejects: %v", err)
			}
		})
	}
}

func TestParseScheduleRejects(t *testing.T) {
	tests := []struct {
		name string
		req  *ScheduleRequest
	}{
		{"nil schedule", nil},
		{"empty kind", &ScheduleRequest{}},
		{"unknown kind", &ScheduleRequest{Kind: "cron"}},

		{"every below the 1-minute floor", &ScheduleRequest{Kind: "every", EverySeconds: 59}},
		{"every one second", &ScheduleRequest{Kind: "every", EverySeconds: 1}},
		{"every zero", &ScheduleRequest{Kind: "every", EverySeconds: 0}},
		{"every negative", &ScheduleRequest{Kind: "every", EverySeconds: -60}},
		{"every above the 30-day ceiling", &ScheduleRequest{Kind: "every", EverySeconds: 31 * 24 * 3600}},
		{"every large enough to overflow a Duration", &ScheduleRequest{Kind: "every", EverySeconds: 1 << 62}},

		{"daily hour out of range", &ScheduleRequest{Kind: "daily", Hour: 24}},
		{"daily hour negative", &ScheduleRequest{Kind: "daily", Hour: -1}},
		{"daily minute out of range", &ScheduleRequest{Kind: "daily", Hour: 6, Minute: 60}},
		{"daily unknown timezone", &ScheduleRequest{Kind: "daily", Hour: 6, Timezone: "Mars/Olympus_Mons"}},

		{"weekly missing weekday", &ScheduleRequest{Kind: "weekly", Hour: 8}},
		{"weekly weekday out of range", &ScheduleRequest{Kind: "weekly", Weekday: weekdayPtr(7), Hour: 8}},
		{"weekly weekday negative", &ScheduleRequest{Kind: "weekly", Weekday: weekdayPtr(-1), Hour: 8}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseSchedule(context.Background(), tc.req)
			if err == nil {
				t.Fatal("parseSchedule() = nil error, want rejection")
			}
			mErr, ok := errs.IsMathError(err)
			if !ok {
				t.Fatalf("parseSchedule() = %v, want a MathError", err)
			}
			if got := mErr.GetStatusCode(); got != status.JOB_INVALID_SCHEDULE {
				t.Fatalf("status code = %d, want JOB_INVALID_SCHEDULE (%d)", got, status.JOB_INVALID_SCHEDULE)
			}
		})
	}
}

// TestParseTimezoneFailsLoudly is the guard against the failure mode
// that motivated embedding tzdata: jobs.loadProjectTimezone silently
// substitutes UTC for an unresolvable zone, which over an API would ack
// a schedule the operator never asked for. The API path must reject.
func TestParseTimezoneFailsLoudly(t *testing.T) {
	if _, err := parseTimezone(context.Background(), "Definitely/Not_A_Zone"); err == nil {
		t.Fatal("parseTimezone(unknown) = nil error, want rejection rather than a silent UTC fallback")
	}

	loc, err := parseTimezone(context.Background(), "")
	if err != nil {
		t.Fatalf("parseTimezone(\"\") = %v, want UTC", err)
	}
	if loc != time.UTC {
		t.Fatalf("parseTimezone(\"\") = %v, want UTC", loc)
	}

	// tzdata is embedded by cmd/mathsvr, but the test binary links only
	// this package — so this asserts the host can resolve it, matching
	// what the infrastructure tests assume.
	if _, err := parseTimezone(context.Background(), "Asia/Ho_Chi_Minh"); err != nil {
		t.Fatalf("parseTimezone(Asia/Ho_Chi_Minh) = %v, want success", err)
	}
}
