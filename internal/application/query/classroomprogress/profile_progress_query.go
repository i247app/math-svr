package classroomprogress

import (
	"context"
	"fmt"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
	exercise "math-ai.com/math-ai/internal/domain/exercise"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// PassScorePctThreshold is the minimum score_percentage (0–100) a graded
// submission must reach to count as "đạt". 50 == 5.0/10. One const so a
// future product tweak is a single line. See FEATURE-SPEC §4.
const PassScorePctThreshold int64 = 50

// ProfileProgressQuery carries the resolved inputs for the single-student
// progress detail. The module layer resolves TargetProfileID (self vs
// other) and the tz default before handing this in.
type ProfileProgressQuery struct {
	ClassroomID     int64
	TargetProfileID int64
	From            string
	To              string
	Purpose         *string // nil => all exercise purposes
}

// ProfileProgressResult is the query's pure-data output. The module
// wraps it into dto.ProfileProgressRes with the request echoes.
type ProfileProgressResult struct {
	Series  []dto.ExercisePoint
	Summary dto.ProgressSummary
}

type ProfileProgressQueryHandler struct {
	submissionRepo exercise.ISubmissionRepository
	exerciseRepo   exercise.IRepository
}

func NewProfileProgressQueryHandler(
	submissionRepo exercise.ISubmissionRepository,
	exerciseRepo exercise.IRepository,
) *ProfileProgressQueryHandler {
	return &ProfileProgressQueryHandler{
		submissionRepo: submissionRepo,
		exerciseRepo:   exerciseRepo,
	}
}

func (h *ProfileProgressQueryHandler) Handle(ctx context.Context, q ProfileProgressQuery) (*ProfileProgressResult, error) {
	queryParams := exercise.ProfileSubmissionsRangeParams{
		ClassroomID: q.ClassroomID,
		ProfileID:   q.TargetProfileID,
		// From:        q.From,
		// To:          q.To,
	}

	if q.From != "" && q.To != "" {
		parseFromDate, err := mtime.ParseFromString(q.From)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		queryParams.From = parseFromDate

		parseToDate, err := mtime.ParseFromString(q.To)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		queryParams.To = parseToDate
	}

	// Current period.
	curSubs, err := h.submissionRepo.ListProfileSubmissionsInRange(ctx, queryParams)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	curExMap, err := h.hydrateExercises(ctx, curSubs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	points := buildPoints(curSubs, curExMap, q.Purpose)

	// Prior period: the immediately-preceding window of the same length.
	priorAvg10, err := h.priorPeriodAvg10(ctx, q)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	summary := buildSummary(points, priorAvg10)
	return &ProfileProgressResult{Series: points, Summary: summary}, nil
}

// priorPeriodAvg10 fetches the same-length window ending just before
// q.From and returns its 10-point average, or nil when that window has
// no graded submissions (after the purpose filter).
func (h *ProfileProgressQueryHandler) priorPeriodAvg10(ctx context.Context, q ProfileProgressQuery) (*float64, error) {
	parseFromDate, err := mtime.ParseFromString(q.From)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	parseToDate, err := mtime.ParseFromString(q.To)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	dur := parseToDate.Sub(parseFromDate.Time)
	if dur <= 0 {
		return nil, nil
	}
	priorTo := parseFromDate.Time.Add(-time.Microsecond)
	priorFrom := parseFromDate.Time.Add(-dur)

	subs, err := h.submissionRepo.ListProfileSubmissionsInRange(ctx, exercise.ProfileSubmissionsRangeParams{
		ClassroomID: q.ClassroomID,
		ProfileID:   q.TargetProfileID,
		From:        mtime.NewTime(priorFrom),
		To:          mtime.NewTime(priorTo),
	})
	if err != nil {
		return nil, err
	}
	exMap, err := h.hydrateExercises(ctx, subs)
	if err != nil {
		return nil, err
	}
	points := buildPoints(subs, exMap, q.Purpose)
	if len(points) == 0 {
		return nil, nil
	}
	avg := avg10OfPoints(points)
	return &avg, nil
}

// hydrateExercises loads the exercise rows behind a submission set,
// keyed by classroom_exercise_id, for title + purpose lookup.
func (h *ProfileProgressQueryHandler) hydrateExercises(ctx context.Context, subs []*exercise.Submission) (map[int64]*exercise.Exercise, error) {
	if len(subs) == 0 {
		return map[int64]*exercise.Exercise{}, nil
	}
	ids := make([]int64, 0, len(subs))
	seen := make(map[int64]struct{}, len(subs))
	for _, s := range subs {
		id := s.ClassroomExerciseId()
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	exRows, err := h.exerciseRepo.ListByClassroomExerciseIds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("profile progress hydrate exercises: %w", err)
	}
	out := make(map[int64]*exercise.Exercise, len(exRows))
	for _, e := range exRows {
		out[e.ClassroomExerciseId()] = e
	}
	return out, nil
}

// buildPoints maps submissions → chart points, applying the optional
// purpose filter. When a purpose filter is active, submissions whose
// exercise can't be resolved (or whose purpose differs) are dropped.
// When no filter is set, every graded submission is kept with a "#<id>"
// title fallback.
func buildPoints(subs []*exercise.Submission, exMap map[int64]*exercise.Exercise, purposeFilter *string) []dto.ExercisePoint {
	out := make([]dto.ExercisePoint, 0, len(subs))
	for _, s := range subs {
		if s.ScorePercentage() == nil {
			continue
		}
		exID := s.ClassroomExerciseId()
		ex := exMap[exID]
		if purposeFilter != nil {
			if ex == nil || ex.Purpose() != *purposeFilter {
				continue
			}
		}
		pct := *s.ScorePercentage()
		title := fmt.Sprintf("#%d", exID)
		if ex != nil && ex.Title() != "" {
			title = ex.Title()
		}
		out = append(out, dto.ExercisePoint{
			ExerciseID:     exID,
			Title:          title,
			SubmittedDt:    s.SubmittedDt().String(),
			Score:          PctTo10Pt(float64(pct)),
			ScorePct:       pct,
			CorrectNumber:  s.CorrectNumber(),
			TotalQuestions: s.TotalQuestions(),
			Passed:         pct >= PassScorePctThreshold,
		})
	}
	return out
}

// avg10OfPoints returns the 10-point average over the points' raw pct.
func avg10OfPoints(points []dto.ExercisePoint) float64 {
	var sumPct int64
	for _, p := range points {
		sumPct += p.ScorePct
	}
	return PctTo10Pt(float64(sumPct) / float64(len(points)))
}

// buildSummary aggregates the four cards + trend from the chart points.
func buildSummary(points []dto.ExercisePoint, priorAvg10 *float64) dto.ProgressSummary {
	graded := int64(len(points))
	summary := dto.ProgressSummary{
		TotalExercises: graded,
		GradedCount:    graded,
	}
	if graded == 0 {
		summary.Trend = string(enum.ProgressCommentNoData)
		return summary
	}

	var sumPct int64
	var passed int64
	scores10 := make([]float64, 0, len(points))
	var hi *dto.ExercisePoint
	for i := range points {
		p := points[i]
		sumPct += p.ScorePct
		if p.Passed {
			passed++
		}
		scores10 = append(scores10, p.Score)
		// Highest by pct; ties resolve to the latest (points are ASC by
		// submitted_dt, so >= keeps the most recent).
		if hi == nil || p.ScorePct >= hi.ScorePct {
			hi = &points[i]
		}
	}

	avgPct := float64(sumPct) / float64(graded)
	avg10 := PctTo10Pt(avgPct)
	slope := LinearSlope(scores10)

	summary.AverageScore = &avg10
	summary.AverageScorePct = &avgPct
	summary.PassedCount = passed
	correctRate := float64(passed) / float64(graded)
	summary.CorrectRate = &correctRate
	summary.Trend = string(Classify(int(graded), avg10, slope))

	if hi != nil {
		hiScore := hi.Score
		hiPct := hi.ScorePct
		hiID := hi.ExerciseID
		hiTitle := hi.Title
		summary.HighestScore = &hiScore
		summary.HighestScorePct = &hiPct
		summary.HighestExerciseID = &hiID
		summary.HighestExerciseTitle = &hiTitle
		summary.HighestSubmittedDt = hi.SubmittedDt
	}

	if priorAvg10 != nil {
		delta := avg10 - *priorAvg10
		summary.AverageDelta = &delta
	}

	return summary
}
