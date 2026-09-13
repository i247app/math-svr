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
	h := query.NewGetExamStatsQueryHandler(repo)

	t.Run("every journey", func(t *testing.T) {
		got, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1})
		if err != nil {
			t.Fatal(err)
		}
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
		got, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1, ExamType: strPtr("ASSESSMENT")})
		if err != nil {
			t.Fatal(err)
		}
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
		got, err := h.Handle(context.Background(), query.GetExamStatsQuery{UserID: 1, ProfileID: 1, ExamType: strPtr("GRADE")})
		if err != nil {
			t.Fatal(err)
		}
		if repo.lastFilter.ExamType == nil || *repo.lastFilter.ExamType != "GRADE" {
			t.Error("a GRADE read has no practice to nest and should filter at the repository")
		}
		if len(got) != 1 || got[0].Practice != nil {
			t.Fatalf("got %d entries, want the one GRADE journey with no practice", len(got))
		}
	})
}
