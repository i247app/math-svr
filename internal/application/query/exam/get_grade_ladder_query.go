package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

// gradeLadderReader is the slice of the journey repository this query
// needs, so a hand-rolled fake can stand in for it.
type gradeLadderReader interface {
	ListLatestPassedByLevel(ctx context.Context, uid, profileId int64, grade int) ([]*exam.ExamSession, error)
	FindLatestJourney(ctx context.Context, filter exam.LatestJourneyFilter) (*exam.ExamSession, error)
}

type GetGradeLadderQuery struct {
	UID       int64
	ProfileID int64
	Grade     int
}

// GradeLadderEntry is one rung the client draws. IsLatest marks the entry
// for the journey the child worked last; every other entry is the latest
// PASS of its level.
type GradeLadderEntry struct {
	Journey  *exam.ExamSession
	IsLatest bool
}

// GetGradeLadderQueryHandler builds one grade's level ladder as a single
// list: every level the child has passed (its most recent pass — a level
// passed, then stepped down from, stays on the ladder), lowest level
// first, then the latest journey of the grade as its own entry with
// IsLatest set.
//
// The latest entry is always separate, by decision: when the latest
// journey is itself the latest pass of its level it appears twice — once
// as that level's pass, once as the latest — and the client tells them
// apart by IsLatest. "Latest" is FindLatestJourney's, the same journey
// /exams/sessions/latest and latest_level name.
type GetGradeLadderQueryHandler struct {
	reader gradeLadderReader
}

func NewGetGradeLadderQueryHandler(reader gradeLadderReader) *GetGradeLadderQueryHandler {
	return &GetGradeLadderQueryHandler{reader: reader}
}

func (h *GetGradeLadderQueryHandler) Handle(ctx context.Context, q GetGradeLadderQuery) ([]GradeLadderEntry, error) {
	passed, err := h.reader.ListLatestPassedByLevel(ctx, q.UID, q.ProfileID, q.Grade)
	if err != nil {
		return nil, err
	}

	gradeType := string(enum.ExamTypeGrade)
	grade := q.Grade
	latest, err := h.reader.FindLatestJourney(ctx, exam.LatestJourneyFilter{
		Uid: q.UID, ProfileId: q.ProfileID, ExamType: &gradeType, Grade: &grade,
	})
	if err != nil {
		return nil, err
	}

	out := make([]GradeLadderEntry, 0, len(passed)+1)
	for _, j := range passed {
		out = append(out, GradeLadderEntry{Journey: j})
	}
	if latest != nil {
		out = append(out, GradeLadderEntry{Journey: latest, IsLatest: true})
	}
	return out, nil
}
