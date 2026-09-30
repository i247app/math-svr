package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
)

// Defaults for a grade the child has no standing in. LatestLevel 0 says
// "never worked this grade"; MaxLevel 1 is the first lock, which is open
// to everyone — so the client can render the ladder without a special case.
const (
	NoLatestLevel = 0
	FirstLevel    = 1
)

// gradeLevelsReader is the slice of the journey repository this query
// needs, so a hand-rolled fake can stand in for it.
type gradeLevelsReader interface {
	FindGradeLevels(ctx context.Context, uid, profileId int64, grade int) (exam.GradeLevels, error)
}

type GetGradeLevelsQuery struct {
	UID       int64
	ProfileID int64
	Grade     int
}

type GetGradeLevelsResult struct {
	LatestLevel int
	MaxLevel    int
}

// GetGradeLevelsQueryHandler reports where a child stands on one grade's
// GRADE level ladder (exam.GradeLevels): the level of the journey worked
// last, and the highest level ever reached.
type GetGradeLevelsQueryHandler struct {
	reader gradeLevelsReader
}

func NewGetGradeLevelsQueryHandler(reader gradeLevelsReader) *GetGradeLevelsQueryHandler {
	return &GetGradeLevelsQueryHandler{reader: reader}
}

func (h *GetGradeLevelsQueryHandler) Handle(ctx context.Context, q GetGradeLevelsQuery) (*GetGradeLevelsResult, error) {
	levels, err := h.reader.FindGradeLevels(ctx, q.UID, q.ProfileID, q.Grade)
	if err != nil {
		return nil, err
	}

	// Each end falls back on its own, though in practice both are set or
	// neither is: they read the same rows.
	res := &GetGradeLevelsResult{LatestLevel: NoLatestLevel, MaxLevel: FirstLevel}
	if levels.Latest != nil {
		res.LatestLevel = *levels.Latest
	}
	if levels.Max != nil {
		res.MaxLevel = *levels.Max
	}
	return res, nil
}
