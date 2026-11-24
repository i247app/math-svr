package game

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// Result is what a submit writes on a round.
type Result struct {
	AnswersJson     string
	TotalQuestions  int
	CorrectNumber   int
	SkippedNumber   int
	ScorePercentage int
	SubmittedDt     time.Time
	ModifyId        *int64
}

// KeysetParams drives IRepository.ListCompleteKeyset. At most one of Older /
// Newer is set; neither = the first page.
type KeysetParams struct {
	Older *int64 // next page: the rounds right below this gsess_id
	Newer *int64 // previous page: the rounds right above this gsess_id
	Limit int64
}

// IRepository owns ma_game_sessions.
type IRepository interface {
	Create(ctx context.Context, s *GameSession) (*GameSession, error)
	// FindByGsessId returns (nil, nil) when there is no such live round.
	FindByGsessId(ctx context.Context, gsessId int64) (*GameSession, error)
	// MarkComplete moves an IN_PROGRESS round to COMPLETE with its result,
	// or fails with ErrSessionNotInProgress.
	MarkComplete(ctx context.Context, gsessId int64, res Result) error

	// The list reads below return a child's COMPLETE rounds, newest first
	// (gsess_id DESC), WITHOUT questions and answers.
	ListCompletePage(ctx context.Context, uid, profileId int64, page, size int64) ([]*GameSession, *pagination.Pagination, error)
	ListCompleteKeyset(ctx context.Context, uid, profileId int64, params KeysetParams) ([]*GameSession, error)
	// ExistsCompleteAtOrNewer / AtOrOlder report whether any round of that
	// list sits at gsessId or above / below it.
	ExistsCompleteAtOrNewer(ctx context.Context, uid, profileId, gsessId int64) (bool, error)
	ExistsCompleteAtOrOlder(ctx context.Context, uid, profileId, gsessId int64) (bool, error)
}
