package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListQuizzesQuery filters the quiz history by any combination of
// ProfileID / UserID. Either or both may be set; the repository AND's
// the populated filters together.
type ListQuizzesQuery struct {
	ProfileID *string
	UserID    *string
	Page      int64
	Limit     int64
}

type ListQuizzesQueryHandler struct {
	quizRepo quiz.IRepository
}

func NewListQuizzesQueryHandler(quizRepo quiz.IRepository) *ListQuizzesQueryHandler {
	return &ListQuizzesQueryHandler{quizRepo: quizRepo}
}

func (h *ListQuizzesQueryHandler) Handle(ctx context.Context, q ListQuizzesQuery) ([]*quiz.Quiz, *pagination.Pagination, error) {
	return h.quizRepo.ListQuizzes(ctx, quiz.ListQuizzesFilter{
		ProfileID: q.ProfileID,
		UserID:    q.UserID,
	}, q.Page, q.Limit)
}
