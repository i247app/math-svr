package classroomprogress

import (
	"math"
	"sort"
	"time"

	exercise "math-ai.com/math-ai/internal/domain/exercise"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SlopeWindowDays is the size of the rolling window used to compute a
// student's slope + comment chip, in calendar days. Fixed at 5 per
// decision #1 — independent of the chart's bucket setting.
const SlopeWindowDays = 5

// PctTo10Pt converts a 0–100 score_percentage to the 10-point UI
// scale with TWO decimal places. The boss's formula (2026-06-10
// refinement):
//
//	avg_score = round(avg_score_pct / 10, 2)
//
// Worked examples:
//
//	pct = 87.5    → 8.75
//	pct = 62.4    → 6.24
//	pct = 65.456  → 6.55
//	pct = 100     → 10.00
//
// Single source of truth for the rounding rule across the progress
// endpoints — anywhere `avg_score` / `highest_score` / `lowest_score`
// is emitted, it goes through this helper.
func PctTo10Pt(pct float64) float64 {
	return math.Round(pct*10) / 100
}

// LinearSlope returns the slope of the least-squares regression of
// values onto their integer indices. NaN entries are skipped, with
// each remaining value keeping its original index as x — so a vector
// like [6.2, NaN, 7.4, NaN, 8.0] regresses (0,6.2), (2,7.4), (4,8.0).
// This preserves the per-day rate when some calendar days have no
// submissions (decision: slope unit is "10-point pts per day").
//
// Returns 0 when fewer than 2 non-NaN points are present — keeps the
// caller's downstream Classify call total. Returns 0 when the x values
// all collapse to a single index (denominator vanishes).
func LinearSlope(values []float64) float64 {
	type pt struct{ x, y float64 }
	pts := make([]pt, 0, len(values))
	for i, v := range values {
		if math.IsNaN(v) {
			continue
		}
		pts = append(pts, pt{x: float64(i), y: v})
	}
	if len(pts) < 2 {
		return 0
	}
	var sx, sy float64
	for _, p := range pts {
		sx += p.x
		sy += p.y
	}
	n := float64(len(pts))
	meanX := sx / n
	meanY := sy / n
	var num, den float64
	for _, p := range pts {
		dx := p.x - meanX
		num += dx * (p.y - meanY)
		den += dx * dx
	}
	if den == 0 {
		return 0
	}
	return num / den
}

// Classify maps a student's (submissionCount, avg10pt, slope) tuple to
// a comment enum. The rule table mirrors FEATURE-SPEC §4 in
// first-match-wins order — NO_DATA / INSUFFICIENT first so a student
// with too few points always shows the sentinel, then NEED_TO_TRY
// catches everyone who's actively struggling (low avg OR downward
// slope), then GOOD_PROGRESS demands BOTH a high avg AND non-negative
// slope, then PROGRESS is the catch-all.
//
// All threshold magic-numbers live here so a future product tweak is
// one file to touch. Slope unit: 10-point pts per day.
func Classify(submissionCount int, avg10pt, slope float64) enum.ProgressComment {
	switch {
	case submissionCount == 0:
		return enum.ProgressCommentNoData
	case submissionCount == 1:
		return enum.ProgressCommentInsufficient
	case avg10pt < 5.0 || slope <= -0.05:
		return enum.ProgressCommentNeedToTry
	case avg10pt >= 7.5 && slope >= 0.0:
		return enum.ProgressCommentGoodProgress
	default:
		return enum.ProgressCommentProgress
	}
}

// BucketAvg is one in-Go aggregation step: the bucket label plus the
// raw 0–100 average for that bucket. Used as the intermediate shape
// between rows-by-student and the sparkline DTO.
type BucketAvg struct {
	Label  string
	AvgPct float64
}

// BuildTrendSeries buckets a single student's rows per bucket+loc,
// computes the per-bucket average, sorts by label ascending, and
// returns the last n entries — matching decision #14 ("last 5 buckets
// at the chart's bucket granularity"). Rows are assumed already
// filtered to the requested date range by the repo.
func BuildTrendSeries(rows []*exercise.ProgressRow, bucket string, loc *time.Location, n int) []BucketAvg {
	if len(rows) == 0 || n <= 0 {
		return nil
	}
	type acc struct {
		sum   int64
		count int64
	}
	byLabel := make(map[string]acc, n)
	for _, r := range rows {
		label := BucketLabel(r.SubmittedDt.Time, bucket, loc)
		a := byLabel[label]
		a.sum += r.ScorePercentage
		a.count++
		byLabel[label] = a
	}
	labels := make([]string, 0, len(byLabel))
	for k := range byLabel {
		labels = append(labels, k)
	}
	sort.Strings(labels)
	if len(labels) > n {
		labels = labels[len(labels)-n:]
	}
	out := make([]BucketAvg, 0, len(labels))
	for _, lab := range labels {
		a := byLabel[lab]
		out = append(out, BucketAvg{
			Label:  lab,
			AvgPct: float64(a.sum) / float64(a.count),
		})
	}
	return out
}

// LastNExerciseRows returns the trailing slice of up to n elements from
// rows. Repo guarantees rows arrive already sorted by submitted_dt ASC
// per profile (see ListForProgress), so the trailing slice is exactly
// "the n most recent submissions oldest→newest" — ready to feed
// trend_series_by_exercise.
func LastNExerciseRows(rows []*exercise.ProgressRow, n int) []*exercise.ProgressRow {
	if n <= 0 || len(rows) == 0 {
		return nil
	}
	if len(rows) <= n {
		return rows
	}
	return rows[len(rows)-n:]
}

// BuildSlopeWindow computes the daily-average vector for the last 5
// calendar days ending at `to` (inclusive), in loc. Returns exactly
// SlopeWindowDays entries (oldest first); days with no submissions
// land as NaN so LinearSlope can preserve per-day spacing.
//
// `to` is typically the request's ToDt — for the prior-period delta
// pass, the caller passes the prior window's ToDt.
func BuildSlopeWindow(rows []*exercise.ProgressRow, loc *time.Location, to time.Time) []float64 {
	// Index rows by calendar-day label in loc for O(N) lookup.
	type acc struct {
		sum   int64
		count int64
	}
	byDay := make(map[string]acc, SlopeWindowDays*2)
	for _, r := range rows {
		label := BucketLabel(r.SubmittedDt.Time, string(enum.ProgressBucketDay), loc)
		a := byDay[label]
		a.sum += r.ScorePercentage
		a.count++
		byDay[label] = a
	}

	// Walk the 5 days oldest→newest so x=0 is the oldest day. End-day
	// is the day containing `to` in loc.
	localTo := to.In(loc)
	endDay := time.Date(localTo.Year(), localTo.Month(), localTo.Day(), 0, 0, 0, 0, loc)
	out := make([]float64, SlopeWindowDays)
	for i := range SlopeWindowDays {
		// daysBack so index 0 = oldest (4 days before endDay),
		// index 4 = newest (endDay itself).
		daysBack := SlopeWindowDays - 1 - i
		day := endDay.AddDate(0, 0, -daysBack)
		label := day.Format("2006-01-02")
		if a, ok := byDay[label]; ok && a.count > 0 {
			out[i] = PctTo10Pt(float64(a.sum) / float64(a.count))
		} else {
			out[i] = math.NaN()
		}
	}
	return out
}
