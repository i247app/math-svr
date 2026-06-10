// Package classroomprogress holds the read-side computation for the
// /classrooms/progress/* endpoints. The package is intentionally
// separate from application/query/classroom so the chart + analytics
// math doesn't bloat the existing CRUD-of-classroom code.
package classroomprogress

import (
	"errors"
	"strconv"
	"time"

	"math-ai.com/math-ai/internal/shared/enum"
)

// ErrInvalidTzOffset is returned by ParseTzOffset for strings that
// don't match enum.IsValidTzOffset. The module-layer validator should
// catch this before any query handler runs; query handlers still call
// ParseTzOffset defensively so a malformed value never reaches MySQL.
var ErrInvalidTzOffset = errors.New("invalid tz offset")

// ParseTzOffset turns a "+07:00" / "-05:30"–style string into a
// time.FixedZone usable for bucket-boundary arithmetic. The same
// string is also passed verbatim into MySQL CONVERT_TZ in the repo
// layer; keeping a single source of truth here matters for
// week/month boundaries that depend on local-midnight alignment.
func ParseTzOffset(tz string) (*time.Location, error) {
	if !enum.IsValidTzOffset(tz) {
		return nil, ErrInvalidTzOffset
	}
	// Format guaranteed by the regex: [+-]\d{2}:\d{2}
	sign := 1
	if tz[0] == '-' {
		sign = -1
	}
	h, err := strconv.Atoi(tz[1:3])
	if err != nil {
		return nil, ErrInvalidTzOffset
	}
	m, err := strconv.Atoi(tz[4:6])
	if err != nil {
		return nil, ErrInvalidTzOffset
	}
	offsetSec := sign * (h*3600 + m*60)
	return time.FixedZone(tz, offsetSec), nil
}

// BucketKey is one entry in the zero-fill sequence: the label the
// caller emits (matching the SQL DATE_FORMAT output) plus the bucket's
// local-midnight start instant.
type BucketKey struct {
	Label string
	Start time.Time
}

// BucketLabel formats t into the bucket label used by both the SQL
// aggregation (see bucketLabelExpr in the mysql repo) and the in-Go
// trend-series grouping. t is expected in UTC; loc applies the request
// tz before formatting so the boundaries land on local midnight.
//
// DAY   → "YYYY-MM-DD"   (calendar day in loc)
// WEEK  → "YYYY-MM-DD"   (Monday of the ISO week in loc)
// MONTH → "YYYY-MM"      (calendar month in loc)
//
// Unrecognised bucket values fall through to DAY so the function stays
// total — the module-layer validator already gates on enum.IsValid.
func BucketLabel(t time.Time, bucket string, loc *time.Location) string {
	local := t.In(loc)
	switch bucket {
	case string(enum.ProgressBucketWeek):
		// Go's time.Weekday: Sunday=0, Monday=1, …, Saturday=6.
		// MySQL WEEKDAY: Monday=0, …, Sunday=6. We want Monday-start;
		// shift back to Monday by `(wd+6)%7` days so Sunday rolls back
		// six (matching WEEKDAY semantics).
		wd := int(local.Weekday())
		shift := (wd + 6) % 7
		monday := time.Date(local.Year(), local.Month(), local.Day()-shift, 0, 0, 0, 0, loc)
		return monday.Format("2006-01-02")
	case string(enum.ProgressBucketMonth):
		return local.Format("2006-01")
	default: // DAY
		return local.Format("2006-01-02")
	}
}

// BucketStart returns the local-midnight start instant for the bucket
// that contains t. Used by the chart's zero-fill walk to populate
// each emitted point's `bucket_start`. The returned time is in loc —
// callers typically pass it as-is into mtime.MathTime which serialises
// without a tz suffix.
func BucketStart(t time.Time, bucket string, loc *time.Location) time.Time {
	local := t.In(loc)
	switch bucket {
	case string(enum.ProgressBucketWeek):
		wd := int(local.Weekday())
		shift := (wd + 6) % 7
		return time.Date(local.Year(), local.Month(), local.Day()-shift, 0, 0, 0, 0, loc)
	case string(enum.ProgressBucketMonth):
		return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
	default: // DAY
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	}
}

// bucketAdvance moves from one bucket's start instant to the next.
// Kept separate from BucketStart so the zero-fill walk doesn't
// re-derive the boundary calculation on each step.
func bucketAdvance(start time.Time, bucket string, loc *time.Location) time.Time {
	switch bucket {
	case string(enum.ProgressBucketWeek):
		return start.AddDate(0, 0, 7)
	case string(enum.ProgressBucketMonth):
		return time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, loc)
	default: // DAY
		return start.AddDate(0, 0, 1)
	}
}

// BucketSequence walks the bucket boundaries between from and to in
// loc and returns one BucketKey per bucket that overlaps the range.
// Used by /scores-over-time to zero-fill empty buckets (decision #8).
//
// from and to are both inclusive — a bucket whose start is ≤ to is
// emitted. An inverted range (to < from) yields the empty slice;
// callers are expected to have validated the range upstream.
func BucketSequence(from, to time.Time, bucket string, loc *time.Location) []BucketKey {
	if to.Before(from) {
		return nil
	}
	out := make([]BucketKey, 0, 8)
	cur := BucketStart(from, bucket, loc)
	for !cur.After(to) {
		out = append(out, BucketKey{
			Label: BucketLabel(cur, bucket, loc),
			Start: cur,
		})
		cur = bucketAdvance(cur, bucket, loc)
	}
	return out
}
