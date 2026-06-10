package classroomprogress

import (
	"context"
	"fmt"

	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	exercise "math-ai.com/math-ai/internal/domain/exercise"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// ScoresOverTimeQuery is the read input for the chart endpoint. Mirrors
// dto.ScoresOverTimeReq minus ProfileID (consumed by the module layer's
// auth gate before the query runs).
type ScoresOverTimeQuery struct {
	ClassroomID     int64
	Bucket          string
	From            mtime.MathTime
	To              mtime.MathTime
	TzOffset        string
	FilterProfileID *int64
}

// ScoresOverTimeQueryHandler runs the chart aggregation + zero-fill.
// Stateless apart from the submission repo it holds onto — safe to
// share across requests.
type ScoresOverTimeQueryHandler struct {
	submissionRepo exercise.ISubmissionRepository
}

func NewScoresOverTimeQueryHandler(submissionRepo exercise.ISubmissionRepository) *ScoresOverTimeQueryHandler {
	return &ScoresOverTimeQueryHandler{submissionRepo: submissionRepo}
}

// Handle fetches the bucketed aggregation from MySQL, then walks the
// bucket sequence between From and To in the request tz and emits
// every bucket. Empty buckets carry nil score fields (decision #8) so
// the client can render gaps on a continuous x-axis.
func (h *ScoresOverTimeQueryHandler) Handle(ctx context.Context, q ScoresOverTimeQuery) (*dto.ScoresOverTimeRes, error) {
	loc, err := ParseTzOffset(q.TzOffset)
	if err != nil {
		return nil, fmt.Errorf("scores-over-time: %w", err)
	}

	dbRows, err := h.submissionRepo.ListBucketedScores(ctx, exercise.BucketedScoresParams{
		ClassroomID:     q.ClassroomID,
		Bucket:          q.Bucket,
		TzOffset:        q.TzOffset,
		From:            q.From,
		To:              q.To,
		FilterProfileID: q.FilterProfileID,
	})
	if err != nil {
		return nil, err
	}

	byLabel := make(map[string]*exercise.BucketedScoreRow, len(dbRows))
	for _, r := range dbRows {
		byLabel[r.BucketLabel] = r
	}

	sequence := BucketSequence(q.From.Time, q.To.Time, q.Bucket, loc)
	points := make([]dto.ChartPoint, 0, len(sequence))
	for _, b := range sequence {
		p := dto.ChartPoint{
			BucketLabel: b.Label,
			BucketStart: mtime.MathTime{Time: b.Start}.String(),
		}
		if r, ok := byLabel[b.Label]; ok {
			avg10 := PctTo10Pt(r.AvgPct)
			max10 := PctTo10Pt(float64(r.MaxPct))
			min10 := PctTo10Pt(float64(r.MinPct))
			avgPct := r.AvgPct
			maxPct := float64(r.MaxPct)
			minPct := float64(r.MinPct)

			p.SubmissionCount = r.SubmissionCount
			p.AvgScore = &avg10
			p.HighestScore = &max10
			p.LowestScore = &min10
			p.AvgScorePct = &avgPct
			p.HighestScorePct = &maxPct
			p.LowestScorePct = &minPct
		}
		points = append(points, p)
	}

	return &dto.ScoresOverTimeRes{
		Bucket: q.Bucket,
		FromDt: q.From.String(),
		ToDt:   q.To.String(),
		Tz:     q.TzOffset,
		Points: points,
	}, nil
}
