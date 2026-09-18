package query

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/query/progress"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// journeyProgressReader is the slice of the journey repository this
// query needs, so a hand-rolled fake can stand in for it.
type journeyProgressReader interface {
	ListProgressPoints(ctx context.Context, params exam.JourneyProgressParams) ([]*exam.UserExam, error)
}

// GetJourneyProgressQuery carries the resolved inputs. From/To are
// optional (empty means an open bound); ExamType nil means every journey
// type; Limit is pre-clamped by the validator.
type GetJourneyProgressQuery struct {
	UserID    int64
	ProfileID int64
	ExamType  *string
	From      string
	To        string
	Limit     int64
}

type GetJourneyProgressResult struct {
	Series  []dto.JourneyPoint
	Summary dto.ExamStatsSummary
}

// GetJourneyProgressQueryHandler is the journey-level twin of
// GetExamProgressQueryHandler: the same windowed series and banner, read
// over ma_user_exams rows instead of sittings. Reading a bounded window
// rather than every journey is the point — the stats list's banner
// aggregates a child's whole history in one call, and that does not
// scale with the table.
type GetJourneyProgressQueryHandler struct {
	reader journeyProgressReader
}

func NewGetJourneyProgressQueryHandler(reader journeyProgressReader) *GetJourneyProgressQueryHandler {
	return &GetJourneyProgressQueryHandler{reader: reader}
}

func (h *GetJourneyProgressQueryHandler) Handle(ctx context.Context, q GetJourneyProgressQuery) (*GetJourneyProgressResult, error) {
	params := exam.JourneyProgressParams{
		UserID:    q.UserID,
		ProfileID: q.ProfileID,
		ExamType:  q.ExamType,
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

	// Current window, newest submission first from the repository.
	desc, err := h.reader.ListProgressPoints(ctx, params)
	if err != nil {
		return nil, err
	}
	series := toJourneySeriesAsc(desc)

	// Prior window: the same-size set of journeys submitted before the
	// current window's oldest point, so average_delta compares two
	// windows rather than floating on its own.
	var priorAvg10 *float64
	if len(series) > 0 {
		anchor := desc[len(desc)-1].LastSubmittedDt()
		prior, err := h.reader.ListProgressPoints(ctx, exam.JourneyProgressParams{
			UserID:          q.UserID,
			ProfileID:       q.ProfileID,
			ExamType:        q.ExamType,
			SubmittedBefore: &anchor,
			Limit:           q.Limit,
		})
		if err != nil {
			return nil, err
		}
		if avg, ok := avg10OfJourneys(prior); ok {
			priorAvg10 = &avg
		}
	}

	points := make([]scorePoint, 0, len(series))
	for _, p := range series {
		points = append(points, scorePoint{ID: p.UserExamID, ScorePct: p.ScorePct})
	}
	core, hiID := summarizeScores(points, priorAvg10)

	return &GetJourneyProgressResult{
		Series:  series,
		Summary: dto.ExamStatsSummary{ExamScoreSummary: core, HighestUserExamID: hiID},
	}, nil
}

// toJourneySeriesAsc reverses the newest-first rows into chronological
// order and assigns the 1..N sequence label the chart's x-axis uses.
// Rows without a score never reach here (the repository filters them),
// so a nil percentage is a defect, not a case — it is skipped rather
// than shown as zero.
func toJourneySeriesAsc(desc []*exam.UserExam) []dto.JourneyPoint {
	n := len(desc)
	out := make([]dto.JourneyPoint, 0, n)
	for i := n - 1; i >= 0; i-- {
		j := desc[i]
		pct := j.ResScorePercentage()
		if pct == nil {
			continue
		}
		p := dto.JourneyPoint{
			Sequence:       int64(len(out) + 1),
			UserExamID:     j.UserExamId(),
			ExamType:       j.ReqExamType(),
			Grade:          j.CurrentGrade(),
			Score:          progress.PctTo10Pt(float64(*pct)),
			ScorePct:       int64(*pct),
			CorrectNumber:  j.ResCorrectNumber(),
			TotalQuestions: j.ResTotalQuestions(),
		}
		if st := j.UserExamStatus(); st != nil {
			p.Status = *st
		}
		if !j.LastSubmittedDt().IsZero() {
			p.LastSubmittedDt = j.LastSubmittedDt().String()
		}
		out = append(out, p)
	}
	return out
}

func avg10OfJourneys(rows []*exam.UserExam) (float64, bool) {
	var sum, n int64
	for _, j := range rows {
		if pct := j.ResScorePercentage(); pct != nil {
			sum += int64(*pct)
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return progress.PctTo10Pt(float64(sum) / float64(n)), true
}
