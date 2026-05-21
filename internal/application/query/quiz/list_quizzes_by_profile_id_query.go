package query

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListQuizzesByProfileIdQuery struct {
	ProfileID uuid.UUID
	Page      int64
	Limit     int64
}

type ListQuizzesByProfileIdQueryHandler struct {
	quizRepo quiz.IRepository
}

func NewListQuizzesByProfileIdQueryHandler(quizRepo quiz.IRepository) *ListQuizzesByProfileIdQueryHandler {
	return &ListQuizzesByProfileIdQueryHandler{quizRepo: quizRepo}
}

func (h *ListQuizzesByProfileIdQueryHandler) Handle(ctx context.Context, q ListQuizzesByProfileIdQuery) ([]*quiz.Quiz, *pagination.Pagination, error) {
	return h.quizRepo.ListByProfileId(ctx, q.ProfileID, q.Page, q.Limit)
}
