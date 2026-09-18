package query_test

import (
	"context"
	"testing"
	"time"

	query "math-ai.com/math-ai/internal/application/query/exam"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeJourneyProgressRepo answers ListProgressPoints the way the MySQL
// repository does: newest submission first, scored rows only, the
// SubmittedBefore call reading strictly before the anchor, both capped
// at Limit.
type fakeJourneyProgressRepo struct {
	rows  []*exam.UserExam // newest first
	calls []exam.JourneyProgressParams
}

func (f *fakeJourneyProgressRepo) ListProgressPoints(_ context.Context, p exam.JourneyProgressParams) ([]*exam.UserExam, error) {
	f.calls = append(f.calls, p)
	var out []*exam.UserExam
	for _, r := range f.rows {
		if r.ResScorePercentage() == nil {
			continue
		}
		if p.ExamType != nil && r.ReqExamType() != *p.ExamType {
			continue
		}
		if p.ExamType == nil && r.ReqExamType() == string(enum.ExamTypePractice) {
			continue
		}
		if p.SubmittedBefore != nil && !r.LastSubmittedDt().Before(p.SubmittedBefore.Time) {
			continue
		}
		out = append(out, r)
		if int64(len(out)) == p.Limit {
			break
		}
	}
	return out, nil
}

func scoredJourney(id int64, typ enum.ExamType, pct int, day int) *exam.UserExam {
	r := row(id, typ)
	r.SetResScorePercentage(&pct)
	r.SetResCorrectNumber(pct / 10)
	r.SetResTotalQuestions(10)
	r.SetLastSubmittedDt(mtime.MathTime{Time: time.Date(2026, 9, day, 8, 0, 0, 0, time.UTC)})
	return r
}

// TestJourneyProgressWindowAndDelta: the chart reads only the newest
// Limit journeys, in chronological order, and compares their average to
// the same-size window just before them. A journey with no score and a
// PRACTICE row are not points.
func TestJourneyProgressWindowAndDelta(t *testing.T) {
	repo := &fakeJourneyProgressRepo{rows: []*exam.UserExam{
		row(60, enum.ExamTypeAssessment),                   // open, never submitted
		scoredJourney(50, enum.ExamTypeAssessment, 90, 15), // window
		scoredJourney(50, enum.ExamTypePractice, 10, 15),   // nested practice: skipped
		scoredJourney(40, enum.ExamTypeGrade, 70, 14),      // window
		scoredJourney(30, enum.ExamTypeAssessment, 60, 13), // prior
		scoredJourney(20, enum.ExamTypeAssessment, 40, 12), // prior
		scoredJourney(10, enum.ExamTypeAssessment, 90, 11), // beyond both windows
	}}
	h := query.NewGetJourneyProgressQueryHandler(repo)

	res, err := h.Handle(context.Background(), query.GetJourneyProgressQuery{UserID: 1, ProfileID: 2, Limit: 2})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(res.Series) != 2 || res.Series[0].UserExamID != 40 || res.Series[1].UserExamID != 50 {
		t.Fatalf("series = %+v, want journeys 40 then 50", res.Series)
	}
	if res.Series[0].Sequence != 1 || res.Series[1].Sequence != 2 {
		t.Errorf("sequence labels = %d, %d, want 1, 2", res.Series[0].Sequence, res.Series[1].Sequence)
	}
	if p := res.Series[1]; p.Score != 9 || p.ScorePct != 90 || p.CorrectNumber != 9 || p.TotalQuestions != 10 || p.LastSubmittedDt == "" {
		t.Errorf("point 50 = %+v", p)
	}

	s := res.Summary
	if s.Count != 2 || s.AverageScorePct == nil || *s.AverageScorePct != 80 {
		t.Errorf("count/average = %d/%v, want 2/80", s.Count, s.AverageScorePct)
	}
	if s.HighestUserExamID == nil || *s.HighestUserExamID != 50 {
		t.Errorf("highest_user_exam_id = %v, want 50", s.HighestUserExamID)
	}
	// Prior window = journeys 30 and 20 → avg 50% = 5.0; current 8.0.
	if s.AverageDelta == nil || *s.AverageDelta != 3 {
		t.Errorf("average_delta = %v, want 3", s.AverageDelta)
	}

	if len(repo.calls) != 2 {
		t.Fatalf("%d repository calls, want the window and its prior", len(repo.calls))
	}
	if prior := repo.calls[1]; prior.SubmittedBefore == nil || !prior.SubmittedBefore.Equal(time.Date(2026, 9, 14, 8, 0, 0, 0, time.UTC)) || prior.Limit != 2 {
		t.Errorf("prior window params = %+v, want anchored at journey 40's submission with the same limit", prior)
	}
}

// TestJourneyProgressEmpty: no scored journey → the empty banner, one
// repository call, no prior-window read.
func TestJourneyProgressEmpty(t *testing.T) {
	repo := &fakeJourneyProgressRepo{rows: []*exam.UserExam{row(60, enum.ExamTypeAssessment)}}
	res, err := query.NewGetJourneyProgressQueryHandler(repo).Handle(context.Background(), query.GetJourneyProgressQuery{UserID: 1, ProfileID: 2, Limit: 10})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(res.Series) != 0 || res.Summary.Count != 0 || res.Summary.Trend != string(enum.ProgressCommentNoData) || res.Summary.AverageDelta != nil {
		t.Errorf("empty result = %+v / %+v", res.Series, res.Summary)
	}
	if len(repo.calls) != 1 {
		t.Errorf("%d repository calls, want 1 (no prior window without a current one)", len(repo.calls))
	}
}

// TestJourneyProgressBadRange: an unparseable bound is an error, not an
// open window.
func TestJourneyProgressBadRange(t *testing.T) {
	repo := &fakeJourneyProgressRepo{}
	if _, err := query.NewGetJourneyProgressQueryHandler(repo).Handle(context.Background(), query.GetJourneyProgressQuery{UserID: 1, ProfileID: 2, From: "not-a-date", Limit: 10}); err == nil {
		t.Fatal("expected an error for an unparseable from_dt")
	}
	if len(repo.calls) != 0 {
		t.Error("the repository must not be read with a broken window")
	}
}
