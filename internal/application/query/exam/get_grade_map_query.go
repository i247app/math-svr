package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

// gradeMapReader is the slice of the journey repository this query
// needs, so a hand-rolled fake can stand in for it.
type gradeMapReader interface {
	ListLatestCompletedByLevel(ctx context.Context, uid, profileId int64, grade int) ([]*exam.ExamSession, error)
	FindLatestJourney(ctx context.Context, filter exam.LatestJourneyFilter) (*exam.ExamSession, error)
}

type GetGradeMapQuery struct {
	UID       int64
	ProfileID int64
	Grade     int
}

// GradeMapEntry is one rung the client draws: the latest COMPLETED
// journey of its level (passed or not — see its esess_flag), or, at most
// once and only at the end, the latest journey of the grade when that one
// is not a level entry. IsLatest marks whichever entry is the journey the
// child worked last.
type GradeMapEntry struct {
	Journey  *exam.ExamSession
	IsLatest bool
}

// GetGradeMapQueryHandler builds one grade's level ladder as a single
// list, one entry per level: every level the child has completed, with
// its most recently COMPLETED journey whether it passed or not
// (esess_flag tells), lowest level first — a level completed, then
// stepped down from, stays on the map.
//
// The latest journey of the grade ("latest" as FindLatestJourney has it,
// the journey /exams/sessions/latest and latest_level name) is flagged,
// never repeated: when it is one of the level entries, that entry gets
// IsLatest. Only when it is not — it is still ACTIVE, or was CANCELLED,
// so no level holds it — is it appended as one extra entry at the end.
// A journey therefore appears at most once.
type GetGradeMapQueryHandler struct {
	reader gradeMapReader
}

func NewGetGradeMapQueryHandler(reader gradeMapReader) *GetGradeMapQueryHandler {
	return &GetGradeMapQueryHandler{reader: reader}
}

func (h *GetGradeMapQueryHandler) Handle(ctx context.Context, q GetGradeMapQuery) ([]GradeMapEntry, error) {
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

	out := make([]GradeMapEntry, 0, len(completed)+1)
	flagged := false
	for _, j := range completed {
		isLatest := latest != nil && j.EsessId() == latest.EsessId()
		flagged = flagged || isLatest
		out = append(out, GradeMapEntry{Journey: j, IsLatest: isLatest})
	}
	if latest != nil && !flagged {
		out = append(out, GradeMapEntry{Journey: latest, IsLatest: true})
	}
	return out, nil
}
