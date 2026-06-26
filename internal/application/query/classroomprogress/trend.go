package classroomprogress

import (
	"math"

	"math-ai.com/math-ai/internal/shared/enum"
)

// PctTo10Pt converts a 0–100 score_percentage to the 10-point UI scale
// with TWO decimal places:
//
//	score = round(pct / 10, 2)
//
// Worked examples:
//
//	pct = 87.5    → 8.75
//	pct = 62.4    → 6.24
//	pct = 100     → 10.00
//
// Single source of truth for the rounding rule across the progress
// endpoints — anywhere a 10-point score is emitted, it goes through here.
func PctTo10Pt(pct float64) float64 {
	return math.Round(pct*10) / 100
}

// LinearSlope returns the slope of the least-squares regression of
// values onto their integer indices. NaN entries are skipped, with each
// remaining value keeping its original index as x. Returns 0 when fewer
// than 2 non-NaN points are present, or when the x values collapse to a
// single index (denominator vanishes).
//
// For the profile-progress view the x-axis is the exercise index, so the
// slope unit is "10-point pts per exercise".
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

// Classify maps a (submissionCount, avg10pt, slope) tuple to a comment
// enum in first-match-wins order — NO_DATA / INSUFFICIENT first so a
// student with too few points always shows the sentinel, then
// NEED_TO_TRY catches everyone struggling (low avg OR downward slope),
// then GOOD_PROGRESS demands BOTH a high avg AND non-negative slope,
// then PROGRESS is the catch-all. See
// docs/features/002-profile-learning-progress/FEATURE-SPEC.md §4.
//
// All threshold magic-numbers live here so a future product tweak is one
// file to touch.
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
