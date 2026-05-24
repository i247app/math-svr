package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/quiz"
)

type GetQuizByQuizIdQuery struct {
	QuizID string
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
