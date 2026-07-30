// Package progress holds score/trend math shared by the learning-progress
// query handlers (classroom exercises + standalone quizzes). Pure
// functions, no I/O — the single source of truth for the 10-point
// rounding rule, the least-squares slope, and the trend classification.
package progress

import (
	"math"

	"math-ai.com/math-ai/internal/shared/enum"
)

// PctTo10Pt converts a 0–100 score_percentage to the 10-point UI scale
// with TWO decimal places: pct=87.5 → 8.75, pct=100 → 10.00.
func PctTo10Pt(pct float64) float64 {
	return math.Round(pct*10) / 100
}

// LinearSlope returns the least-squares slope of values onto their
// integer indices. NaN entries are skipped (keeping original index as x).
// Returns 0 with fewer than 2 non-NaN points or a vanishing denominator.
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

// Classify maps (count, avg10pt, slope) to a comment enum in
// first-match-wins order. Threshold magic-numbers live here. See
// docs/features/002-profile-learning-progress/FEATURE-SPEC.md §4.
func Classify(count int, avg10pt, slope float64) enum.ProgressComment {
	switch {
	case count == 0:
		return enum.ProgressCommentNoData
	case count == 1:
		return enum.ProgressCommentInsufficient
	case avg10pt < 5.0 || slope <= -0.05:
		return enum.ProgressCommentNeedToTry
	case avg10pt >= 7.5 && slope >= 0.0:
		return enum.ProgressCommentGoodProgress
	default:
		return enum.ProgressCommentProgress
	}
}
