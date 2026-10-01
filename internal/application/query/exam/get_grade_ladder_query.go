package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

// gradeLadderReader is the slice of the journey repository this query
// needs, so a hand-rolled fake can stand in for it.
type gradeLadderReader interface {
	ListLatestCompletedByLevel(ctx context.Context, uid, profileId int64, grade int) ([]*exam.ExamSession, error)
	FindLatestJourney(ctx context.Context, filter exam.LatestJourneyFilter) (*exam.ExamSession, error)
}

type GetGradeLadderQuery struct {
	UID       int64
	ProfileID int64
	Grade     int
}

// GradeLadderEntry is one rung the client draws: the latest COMPLETED
// journey of its level (passed or not — see its esess_flag), or, at most
// once and only at the end, the latest journey of the grade when that one
// is not a level entry. IsLatest marks whichever entry is the journey the
// child worked last.
type GradeLadderEntry struct {
	Journey  *exam.ExamSession
	IsLatest bool
}

// GetGradeLadderQueryHandler builds one grade's level ladder as a single
// list, one entry per level: every level the child has completed, with
// its most recently COMPLETED journey whether it passed or not
// (esess_flag tells), lowest level first — a level completed, then
// stepped down from, stays on the ladder.
//
// The latest journey of the grade ("latest" as FindLatestJourney has it,
// the journey /exams/sessions/latest and latest_level name) is flagged,
// never repeated: when it is one of the level entries, that entry gets
// IsLatest. Only when it is not — it is still ACTIVE, or was CANCELLED,
// so no level holds it — is it appended as one extra entry at the end.
// A journey therefore appears at most once.
type GetGradeLadderQueryHandler struct {
	reader gradeLadderReader
}

func NewGetGradeLadderQueryHandler(reader gradeLadderReader) *GetGradeLadderQueryHandler {
	return &GetGradeLadderQueryHandler{reader: reader}
}

func (h *GetGradeLadderQueryHandler) Handle(ctx context.Context, q GetGradeLadderQuery) ([]GradeLadderEntry, error) {
	completed, err := h.reader.ListLatestCompletedByLevel(ctx, q.UID, q.ProfileID, q.Grade)
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

	out := make([]GradeLadderEntry, 0, len(completed)+1)
	flagged := false
	for _, j := range completed {
		isLatest := latest != nil && j.EsessId() == latest.EsessId()
		flagged = flagged || isLatest
		out = append(out, GradeLadderEntry{Journey: j, IsLatest: isLatest})
	}
	if latest != nil && !flagged {
		out = append(out, GradeLadderEntry{Journey: latest, IsLatest: true})
	}
	return out, nil
}
