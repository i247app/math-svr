package query

import (
	"math"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/query/progress"
	"math-ai.com/math-ai/internal/shared/enum"
)

// scorePoint is one scored thing in chronological order — a sitting for
// the progress chart, a journey for the stats list. The summary math
// does not care which; it needs the score and an id to name the best one.
type scorePoint struct {
	ID       int64
	ScorePct int64
}

// summarizeScores aggregates count, averages, highest/lowest, delta and
// trend over a chronological series. It returns the id of the highest
// point separately because the two callers publish it under different
// names. Ties on the high score resolve to the latest point.
func summarizeScores(points []scorePoint, priorAvg10 *float64) (dto.ExamScoreSummary, *int64) {
	count := int64(len(points))
	summary := dto.ExamScoreSummary{Count: count}
	if count == 0 {
		summary.Trend = string(enum.ProgressCommentNoData)
		return summary, nil
	}

	var sumPct int64
	scores10 := make([]float64, 0, len(points))
	hi := points[0]
	lo := points[0]
	for _, p := range points {
		sumPct += p.ScorePct
		scores10 = append(scores10, progress.PctTo10Pt(float64(p.ScorePct)))
		if p.ScorePct >= hi.ScorePct {
			hi = p
		}
		if p.ScorePct < lo.ScorePct {
			lo = p
		}
	}

	// get 2 number after comma
	avgPct := float64(sumPct) / float64(count)
	avgPct = math.Round(avgPct*100) / 100

	avg10 := progress.PctTo10Pt(avgPct)
	slope := progress.LinearSlope(scores10)

	summary.AverageScore = &avg10
	summary.AverageScorePct = &avgPct
	summary.Trend = string(progress.Classify(int(count), avg10, slope))

	hiScore := progress.PctTo10Pt(float64(hi.ScorePct))
	hiPct := hi.ScorePct
	summary.HighestScore = &hiScore
	summary.HighestScorePct = &hiPct

	loScore := progress.PctTo10Pt(float64(lo.ScorePct))
	summary.LowestScore = &loScore

	if priorAvg10 != nil {
		delta := avg10 - *priorAvg10
		summary.AverageDelta = &delta
	}
	hiID := hi.ID
	return summary, &hiID
}
