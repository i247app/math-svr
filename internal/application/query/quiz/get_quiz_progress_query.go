package query

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/application/query/progress"
	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// quizProgressReader is the narrow slice of quiz.IRepository this query
// needs. Accepting an interface (not the concrete repo) keeps the handler
// unit-testable with a hand-rolled fake. quiz.IRepository satisfies it.
type quizProgressReader interface {
	ListProgressPoints(ctx context.Context, params quiz.ProgressPointsParams) ([]*quiz.ProgressPoint, error)
}

// GetQuizProgressQuery carries the resolved inputs. From/To are optional
// (empty string = open bound). Purpose nil = all. Limit is pre-clamped.
type GetQuizProgressQuery struct {
	ProfileID int64
	Purpose   *string
	From      string
	To        string
	Limit     int64
}

// GetQuizProgressResult is the pure-data output; the module wraps it with
// the request echoes into dto.QuizProgressRes.
type GetQuizProgressResult struct {
	Series  []dto.QuizPoint
	Summary dto.QuizProgressSummary
}

type GetQuizProgressQueryHandler struct {
	reader quizProgressReader
}

func NewGetQuizProgressQueryHandler(reader quizProgressReader) *GetQuizProgressQueryHandler {
	return &GetQuizProgressQueryHandler{reader: reader}
}

func (h *GetQuizProgressQueryHandler) Handle(ctx context.Context, q GetQuizProgressQuery) (*GetQuizProgressResult, error) {
	params := quiz.ProgressPointsParams{
		ProfileID: q.ProfileID,
		Purpose:   q.Purpose,
		Limit:     q.Limit,
	}
	if q.From != "" {
		from, err := mtime.ParseFromString(q.From)
		if err != nil {
			return nil, err
		}
		params.From = &from
	}
	if q.To != "" {
		to, err := mtime.ParseFromString(q.To)
		if err != nil {
			return nil, err
		}
		params.To = &to
	}

	// Current window (newest-first from the repo).
	desc, err := h.reader.ListProgressPoints(ctx, params)
	if err != nil {
		return nil, err
	}
	series := toSeriesAsc(desc)

	// Prior window: the same-size set immediately before the current
	// window's oldest point. Only meaningful when the current window has
	// data.
	var priorAvg10 *float64
	if len(series) > 0 {
		anchor := desc[len(desc)-1].CompletedDt // oldest of current
		priorParams := quiz.ProgressPointsParams{
			ProfileID:       q.ProfileID,
			Purpose:         q.Purpose,
			CompletedBefore: &anchor,
			Limit:           q.Limit,
		}
		prior, err := h.reader.ListProgressPoints(ctx, priorParams)
		if err != nil {
			return nil, err
		}
		if avg, ok := avg10OfPoints(prior); ok {
			priorAvg10 = &avg
		}
	}

	return &GetQuizProgressResult{
		Series:  series,
		Summary: buildSummary(series, priorAvg10),
	}, nil
}

// toSeriesAsc reverses the newest-first projection into chronological
// order and assigns the 1..N sequence label.
func toSeriesAsc(desc []*quiz.ProgressPoint) []dto.QuizPoint {
	n := len(desc)
	out := make([]dto.QuizPoint, 0, n)
	for i := n - 1; i >= 0; i-- {
		p := desc[i]
		out = append(out, dto.QuizPoint{
			Sequence:       int64(n - i),
			QuizID:         p.QuizId,
			Purpose:        p.Purpose,
			TypeOfQuiz:     p.TypeOfQuiz,
			Title:          p.Title,
			ShortText:      p.ShortText,
			CompletedDt:    p.CompletedDt.String(),
			Score:          progress.PctTo10Pt(float64(p.ScorePercentage)),
			ScorePct:       p.ScorePercentage,
			CorrectNumber:  p.CorrectNumber,
			TotalQuestions: p.TotalQuestions,
		})
	}
	return out
}

// avg10OfPoints returns the 10-point average of a projection slice and
// whether it had any points.
func avg10OfPoints(points []*quiz.ProgressPoint) (float64, bool) {
	if len(points) == 0 {
		return 0, false
	}
	var sum int64
	for _, p := range points {
		sum += p.ScorePercentage
	}
	return progress.PctTo10Pt(float64(sum) / float64(len(points))), true
}

// buildSummary aggregates count, averages, highest/lowest, delta, trend.
func buildSummary(series []dto.QuizPoint, priorAvg10 *float64) dto.QuizProgressSummary {
	count := int64(len(series))
	summary := dto.QuizProgressSummary{Count: count}
	if count == 0 {
		summary.Trend = string(enum.ProgressCommentNoData)
		return summary
	}

	var sumPct int64
	scores10 := make([]float64, 0, len(series))
	hi := series[0]
	lo := series[0]
	for _, p := range series {
		sumPct += p.ScorePct
		scores10 = append(scores10, p.Score)
		if p.ScorePct >= hi.ScorePct { // ties → latest (series is ASC)
			hi = p
		}
		if p.ScorePct < lo.ScorePct {
			lo = p
		}
	}

	avgPct := float64(sumPct) / float64(count)
	avg10 := progress.PctTo10Pt(avgPct)
	slope := progress.LinearSlope(scores10)

	summary.AverageScore = &avg10
	summary.AverageScorePct = &avgPct
	summary.Trend = string(progress.Classify(int(count), avg10, slope))

	hiScore := hi.Score
	hiPct := hi.ScorePct
	hiID := hi.QuizID
	summary.HighestScore = &hiScore
	summary.HighestScorePct = &hiPct
	summary.HighestQuizID = &hiID

	loScore := lo.Score
	summary.LowestScore = &loScore

	if priorAvg10 != nil {
		delta := avg10 - *priorAvg10
		summary.AverageDelta = &delta
	}
	return summary
}
