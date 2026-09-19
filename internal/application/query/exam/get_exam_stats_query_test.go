package query_test

import (
	"context"
	"testing"

	query "math-ai.com/math-ai/internal/application/query/exam"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

type fakeJourneyRepo struct {
	exam.IUserExamRepository
	rows       []*exam.UserExam
	lastFilter exam.ListJourneysFilter
}

func (f *fakeJourneyRepo) ListByUserProfile(ctx context.Context, userID, profileID int64, filter exam.ListJourneysFilter) ([]*exam.UserExam, error) {
	f.lastFilter = filter
	if filter.ExamType == nil {
		return f.rows, nil
	}
	var out []*exam.UserExam
	for _, r := range f.rows {
		if r.ReqExamType() == *filter.ExamType {
			out = append(out, r)
		}
	}
	return out, nil
}

type fakeAttemptRepo struct {
	exam.IUserAiExamRepository
	open []*exam.UserAiExam
}

func (f *fakeAttemptRepo) ListInProgressByProfile(ctx context.Context, profileID int64) ([]*exam.UserAiExam, error) {
	return f.open, nil
}

type fakeAiExamRepo struct {
	exam.IAiExamRepository
	sets map[int64]*exam.AiExam
}

func (f *fakeAiExamRepo) ListByAiExamIds(ctx context.Context, ids []int64) ([]*exam.AiExam, error) {
	var out []*exam.AiExam
	for _, id := range ids {
		if e, ok := f.sets[id]; ok {
			out = append(out, e)
		}
	}
	return out, nil
}

func sitting(id, aiExamID int64, journey *int64, typ enum.ExamType) *exam.UserAiExam {
	a := exam.NewUserAiExam()
	a.SetUserAiExamId(id)
	a.SetAiExamId(aiExamID)
	a.SetUserExamId(journey)
	a.SetReqExamType(string(typ))
	st := string(enum.UserAiExamStatusInProgress)
	a.SetUserAiExamStatus(&st)
	return a
}

func set(id int64) *exam.AiExam {
	e := exam.NewAiExam()
	e.SetAiExamId(id)
	return e
}

func i64(v int64) *int64 { return &v }

func newStatsHandler(repo *fakeJourneyRepo, open ...*exam.UserAiExam) *query.GetExamStatsQueryHandler {
	sets := map[int64]*exam.AiExam{}
	for _, a := range open {
		sets[a.AiExamId()] = set(a.AiExamId())
	}
	return query.NewGetExamStatsQueryHandler(repo, &fakeAttemptRepo{open: open}, &fakeAiExamRepo{sets: sets})
}

func row(id int64, typ enum.ExamType) *exam.UserExam {
	r := exam.NewUserExam()
	r.SetUserExamId(id)
	r.SetReqExamType(string(typ))
	return r
}

func strPtr(s string) *string { return &s }

// TestStatsNestsPracticeUnderItsJourney: the table holds the PRACTICE row
// as a sibling sharing the id; the dashboard must see it as a child of
// its ASSESSMENT journey, and never as an entry of its own.
func TestStatsNestsPracticeUnderItsJourney(t *testing.T) {
	repo := &fakeJourneyRepo{rows: []*exam.UserExam{
		row(1, enum.ExamTypeAssessment),
		row(2, enum.ExamTypeAssessment), // no practice yet
		row(3, enum.ExamTypeGrade),
		row(1, enum.ExamTypePractice),
	}}
	h := newStatsHandler(repo)

	t.Run("every journey", func(t *testing.T) {
		res, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1})
		if err != nil {
			t.Fatal(err)
		}
		got := res.Journeys
		if len(got) != 3 {
			t.Fatalf("%d entries, want 3 (two ASSESSMENT + one GRADE; PRACTICE nested)", len(got))
		}
		byID := map[int64]exam.JourneyStats{}
		for _, j := range got {
			if j.Journey.ReqExamType() == string(enum.ExamTypePractice) {
				t.Fatal("a PRACTICE row surfaced at the top level")
			}
			byID[j.Journey.UserExamId()] = j
		}
		if byID[1].Practice == nil || byID[1].Practice.ReqExamType() != string(enum.ExamTypePractice) {
			t.Error("journey 1 must carry its PRACTICE row")
		}
		if byID[2].Practice != nil {
			t.Error("journey 2 has no practice and must say so")
		}
		if byID[3].Practice != nil {
			t.Error("a GRADE journey never carries practice")
		}
	})

	t.Run("ASSESSMENT only still nests practice", func(t *testing.T) {
		res, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1, ExamType: strPtr("ASSESSMENT")})
		if err != nil {
			t.Fatal(err)
		}
		got := res.Journeys
		if repo.lastFilter.ExamType != nil {
			t.Error("an ASSESSMENT read must fetch every type so the PRACTICE rows can nest")
		}
		if len(got) != 2 {
			t.Fatalf("%d entries, want the 2 ASSESSMENT journeys", len(got))
		}
		if got[0].Practice == nil {
			t.Error("journey 1's practice row was dropped")
		}
	})

	t.Run("another type is fetched alone", func(t *testing.T) {
		res, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1, ExamType: strPtr("GRADE")})
		if err != nil {
			t.Fatal(err)
		}
		got := res.Journeys
		if repo.lastFilter.ExamType == nil || *repo.lastFilter.ExamType != "GRADE" {
			t.Error("a GRADE read has no practice to nest and should filter at the repository")
		}
		if len(got) != 1 || got[0].Practice != nil {
			t.Fatalf("got %d entries, want the one GRADE journey with no practice", len(got))
		}
	})
}

// TestStatsAttachesUnfinishedSittings: a sitting handed out and never
// submitted hangs under its journey — whichever type it is — with its
// question set hydrated, so the app can offer to resume it. A sitting
// with no journey (never happens after hand-out opens one; legacy rows)
// or whose journey is not on the page is left out.
//
// A child may be handed a new exam while one is open, so a journey can
// carry several, oldest first.
func TestStatsAttachesUnfinishedSittings(t *testing.T) {
	repo := &fakeJourneyRepo{rows: []*exam.UserExam{
		row(1, enum.ExamTypeAssessment),
		row(2, enum.ExamTypeAssessment),
		row(3, enum.ExamTypeAssessment),
	}}
	h := newStatsHandler(repo,
		sitting(11, 501, i64(1), enum.ExamTypeAssessment),
		sitting(12, 502, i64(3), enum.ExamTypePractice),
		sitting(13, 503, nil, enum.ExamTypeAssessment),     // legacy: no journey
		sitting(14, 504, i64(99), enum.ExamTypeAssessment), // journey not on this page
		sitting(15, 505, i64(1), enum.ExamTypeAssessment),  // a second open one in journey 1
	)

	res, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1})
	if err != nil {
		t.Fatal(err)
	}
	byID := map[int64]exam.JourneyStats{}
	for _, j := range res.Journeys {
		byID[j.Journey.UserExamId()] = j
	}

	if got := byID[1].InProgress; len(got) != 2 || got[0].UserAiExamId() != 11 || got[1].UserAiExamId() != 15 {
		t.Fatalf("journey 1 in-progress = %v, want sittings 11 then 15", ids(got))
	}
	if len(byID[2].InProgress) != 0 {
		t.Error("journey 2 has nothing unfinished")
	}
	if got := byID[3].InProgress; len(got) != 1 || got[0].ReqExamType() != string(enum.ExamTypePractice) {
		t.Fatalf("journey 3 in-progress = %v, want its PRACTICE sitting", ids(got))
	}
	for _, id := range []int64{501, 502, 505} {
		if res.AiExams[id] == nil {
			t.Errorf("question set %d was not hydrated", id)
		}
	}
	if len(res.AiExams) != 3 {
		t.Errorf("hydrated %d sets, want only the 3 that were attached", len(res.AiExams))
	}
}

func ids(as []*exam.UserAiExam) []int64 {
	out := make([]int64, 0, len(as))
	for _, a := range as {
		out = append(out, a.UserAiExamId())
	}
	return out
}
