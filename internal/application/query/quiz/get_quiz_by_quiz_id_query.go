package query

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/domain/quiz"
)

type GetQuizByQuizIdQuery struct {
	QuizID uuid.UUID
}

type GetQuizByQuizIdQueryHandler struct {
	quizRepo quiz.IRepository
}

func NewGetQuizByQuizIdQueryHandler(quizRepo quiz.IRepository) *GetQuizByQuizIdQueryHandler {
	return &GetQuizByQuizIdQueryHandler{quizRepo: quizRepo}
}

func (h *GetQuizByQuizIdQueryHandler) Handle(ctx context.Context, q GetQuizByQuizIdQuery) (*quiz.Quiz, error) {
	return h.quizRepo.FindByQuizId(ctx, q.QuizID)
}
