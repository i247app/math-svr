package command

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SubmitQuizAnswersCommand stores the student's answers and the bot's
// grading verdict in a single update. The bot call happens at the
// service layer BEFORE this command, so the transaction stays short.
type SubmitQuizAnswersCommand struct {
	QuizID      uuid.UUID
	AnswersJSON string
	Grading     quiz.GradingUpdate
}

type SubmitQuizAnswersCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSubmitQuizAnswersCommandHandler(uow transaction.UnitOfWork) *SubmitQuizAnswersCommandHandler {
	return &SubmitQuizAnswersCommandHandler{uow: uow}
}

func (h *SubmitQuizAnswersCommandHandler) Handle(ctx context.Context, cmd SubmitQuizAnswersCommand) (*quiz.Quiz, error) {
	var updated *quiz.Quiz

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Quiz.UpdateAnswersAndGrading(ctx, cmd.QuizID,
			cmd.AnswersJSON, cmd.Grading, string(enum.QuizStatusTypeSubmitted)); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		// Read-after-write so the caller hands the client the same shape
		// the listing endpoints will surface — including the freshly
		// written grading counts and SUBMITTED status.
		fresh, err := repos.Quiz.FindByQuizId(ctx, cmd.QuizID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if fresh == nil {
			return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil, nil)
		}
		updated = fresh
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return updated, nil
}
