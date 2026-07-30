package query

import (
	"context"
	"math"
	"testing"
	"time"

	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeReader returns queued result sets in call order: first call = current
// window, second call = prior window.
type fakeReader struct {
	calls   int
	results [][]*quiz.ProgressPoint
}

func (f *fakeReader) ListProgressPoints(_ context.Context, _ quiz.ProgressPointsParams) ([]*quiz.ProgressPoint, error) {
	i := f.calls
	f.calls++
	if i < len(f.results) {
		return f.results[i], nil
	}
	return nil, nil
}

func pt(id, pct int64, day int) *quiz.ProgressPoint {
	return &quiz.ProgressPoint{
		QuizId:          id,
		Purpose:         "PRACTICE",
		ScorePercentage: pct,
		CompletedDt:     mtime.NewTime(time.Date(2026, 5, day, 9, 0, 0, 0, time.UTC)),
	}
}

func TestHandle_EmptyIsNoData(t *testing.T) {
	reader := &fakeReader{results: [][]*quiz.ProgressPoint{{}}}
	h := NewGetQuizProgressQueryHandler(reader)
	res, err := h.Handle(context.Background(), GetQuizProgressQuery{ProfileID: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Series) != 0 {
		t.Errorf("series len = %d want 0", len(res.Series))
	}
	if res.Summary.Trend != string(enum.ProgressCommentNoData) {
		t.Errorf("trend = %s want NO_DATA", res.Summary.Trend)
	}
	if res.Summary.AverageScore != nil || res.Summary.AverageDelta != nil {
		t.Error("nullable summary fields should be nil on empty")
	}
	// An empty current window must NOT trigger the prior-window query —
	// "no quiz history yet" is the common path and should cost one round-trip.
	if reader.calls != 1 {
		t.Errorf("repo calls = %d want 1 (no prior-window query on empty)", reader.calls)
	}
}

func TestHandle_SinglePointInsufficient(t *testing.T) {
	// One completed quiz, and a prior window that exists (so we also prove
	// delta still computes) — but the trend must read INSUFFICIENT because
	// a single point carries no direction.
	current := []*quiz.ProgressPoint{pt(2, 80, 5)}
	prior := []*quiz.ProgressPoint{pt(1, 60, 3)} // avg10 = 6.0
	h := NewGetQuizProgressQueryHandler(&fakeReader{results: [][]*quiz.ProgressPoint{current, prior}})

	res, err := h.Handle(context.Background(), GetQuizProgressQuery{ProfileID: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Series) != 1 || res.Series[0].Sequence != 1 {
		t.Fatalf("series wrong: %+v", res.Series)
	}
	if res.Summary.Trend != string(enum.ProgressCommentInsufficient) {
		t.Errorf("trend = %s want INSUFFICIENT", res.Summary.Trend)
	}
	// delta = 8.0 - 6.0 = 2.0 (still surfaced even for a single point).
	if res.Summary.AverageDelta == nil || math.Abs(*res.Summary.AverageDelta-2.0) > 0.01 {
		t.Errorf("average_delta = %v want ~2.0", res.Summary.AverageDelta)
	}
}

func TestHandle_ParseErrorPropagates(t *testing.T) {
	h := NewGetQuizProgressQueryHandler(&fakeReader{results: [][]*quiz.ProgressPoint{{}}})
	_, err := h.Handle(context.Background(), GetQuizProgressQuery{
		ProfileID: 1, Limit: 10, From: "not-a-date",
	})
	if err == nil {
		t.Fatal("expected a parse error to propagate, got nil")
	}
}

func TestHandle_SeriesAscAndSummary(t *testing.T) {
	// Repo returns newest-first: day6=90, day5=70, day4=60.
	current := []*quiz.ProgressPoint{pt(3, 90, 6), pt(2, 70, 5), pt(1, 60, 4)}
	prior := []*quiz.ProgressPoint{pt(0, 50, 3)} // avg10 = 5.0
	h := NewGetQuizProgressQueryHandler(&fakeReader{results: [][]*quiz.ProgressPoint{current, prior}})

	res, err := h.Handle(context.Background(), GetQuizProgressQuery{ProfileID: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	// Ascending: id1 (seq1), id2 (seq2), id3 (seq3).
	if len(res.Series) != 3 || res.Series[0].QuizID != 1 || res.Series[0].Sequence != 1 ||
		res.Series[2].QuizID != 3 || res.Series[2].Sequence != 3 {
		t.Fatalf("series order/sequence wrong: %+v", res.Series)
	}
	if res.Series[0].Score != 6.0 || res.Series[2].Score != 9.0 {
		t.Errorf("score mapping wrong: %v %v", res.Series[0].Score, res.Series[2].Score)
	}
	// avg pct = (90+70+60)/3 = 73.33 → 7.33.
	if res.Summary.AverageScore == nil || math.Abs(*res.Summary.AverageScore-7.33) > 0.01 {
		t.Errorf("average_score = %v want ~7.33", res.Summary.AverageScore)
	}
	if res.Summary.HighestQuizID == nil || *res.Summary.HighestQuizID != 3 {
		t.Errorf("highest quiz id wrong: %v", res.Summary.HighestQuizID)
	}
	if res.Summary.LowestScore == nil || *res.Summary.LowestScore != 6.0 {
		t.Errorf("lowest score = %v want 6.0", res.Summary.LowestScore)
	}
	// delta = 7.33 - 5.0 = 2.33.
	if res.Summary.AverageDelta == nil || math.Abs(*res.Summary.AverageDelta-2.33) > 0.01 {
		t.Errorf("average_delta = %v want ~2.33", res.Summary.AverageDelta)
	}
}
