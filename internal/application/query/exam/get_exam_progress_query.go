package query

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/query/progress"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// examProgressReader is the narrow slice of the attempt repository this
// query needs. Taking an interface rather than the concrete repo keeps
// the handler testable with a hand-rolled fake.
type examProgressReader interface {
	ListProgressPoints(ctx context.Context, params exam.ProgressPointsParams) ([]*exam.ProgressPoint, error)
}

// GetExamProgressQuery carries the resolved inputs. From/To are optional
// (empty means an open bound); ExamType nil means every type; Limit is
// pre-clamped by the validator.
type GetExamProgressQuery struct {
	ProfileID  int64
	ExamType   *string
	UserExamID *int64
	From       string
	To         string
	Limit      int64
}

type GetExamProgressResult struct {
	Series  []dto.ExamPoint
	Summary dto.ExamProgressSummary
}

type GetExamProgressQueryHandler struct {
	reader examProgressReader
}

func NewGetExamProgressQueryHandler(reader examProgressReader) *GetExamProgressQueryHandler {
	return &GetExamProgressQueryHandler{reader: reader}
}

func (h *GetExamProgressQueryHandler) Handle(ctx context.Context, q GetExamProgressQuery) (*GetExamProgressResult, error) {
	params := exam.ProgressPointsParams{
		ProfileID:  q.ProfileID,
		ExamType:   q.ExamType,
		UserExamID: q.UserExamID,
		Limit:      q.Limit,
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

	// Current window, newest-first from the repository.
	desc, err := h.reader.ListProgressPoints(ctx, params)
	if err != nil {
		return nil, err
	}
	series := toSeriesAsc(desc)

	// Prior window: the same-size set immediately before the current
	// window's oldest point, which is what makes the delta a comparison
	// rather than a number floating on its own. Only meaningful when the
	// current window has anything in it.
	var priorAvg10 *float64
	if len(series) > 0 {
		anchor := desc[len(desc)-1].CompletedDt
		prior, err := h.reader.ListProgressPoints(ctx, exam.ProgressPointsParams{
			ProfileID:       q.ProfileID,
			ExamType:        q.ExamType,
			UserExamID:      q.UserExamID,
			CompletedBefore: &anchor,
			Limit:           q.Limit,
		})
		if err != nil {
			return nil, err
		}
		if avg, ok := avg10OfPoints(prior); ok {
			priorAvg10 = &avg
		}
	}

	return &GetExamProgressResult{
		Series:  series,
		Summary: buildSummary(series, priorAvg10),
	}, nil
}

// toSeriesAsc reverses the newest-first projection into chronological
// order and assigns the 1..N sequence label the chart's x-axis uses.
func toSeriesAsc(desc []*exam.ProgressPoint) []dto.ExamPoint {
	n := len(desc)
	out := make([]dto.ExamPoint, 0, n)
	for i := n - 1; i >= 0; i-- {
		p := desc[i]
		out = append(out, dto.ExamPoint{
			Sequence:       int64(n - i),
			UserAiExamID:   p.UserAiExamId,
			ExamType:       p.ExamType,
			Grade:          p.Grade,
			CompletedDt:    p.CompletedDt.String(),
			Score:          progress.PctTo10Pt(float64(p.ScorePercentage)),
			ScorePct:       p.ScorePercentage,
			CorrectNumber:  p.CorrectNumber,
			TotalQuestions: p.TotalQuestions,
		})
	}
	return out
}

func avg10OfPoints(points []*exam.ProgressPoint) (float64, bool) {
	if len(points) == 0 {
		return 0, false
	}
	var sum int64
	for _, p := range points {
		sum += p.ScorePercentage
	}
	return progress.PctTo10Pt(float64(sum) / float64(len(points))), true
}

// buildSummary renders the banner over the chart's series.
func buildSummary(series []dto.ExamPoint, priorAvg10 *float64) dto.ExamProgressSummary {
	points := make([]scorePoint, 0, len(series))
	for _, p := range series {
		points = append(points, scorePoint{ID: p.UserAiExamID, ScorePct: p.ScorePct})
	}
	core, hiID := summarizeScores(points, priorAvg10)
	return dto.ExamProgressSummary{ExamScoreSummary: core, HighestExamID: hiID}
}
