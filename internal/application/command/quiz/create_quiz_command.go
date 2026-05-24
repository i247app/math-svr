package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// CreateQuizCommand inserts a fresh quiz row. Generation of questions
// (which is the expensive bot call) happens BEFORE this command — the
// command is a thin DB write so the transaction is short and the bot
// I/O doesn't hold a tx open.
//
// UserID and ProfileID are optional: an anonymous / profile-less quiz
// is persisted with both columns NULL and is only reachable via its
// quiz_id thereafter.
type CreateQuizCommand struct {
	UserID         *string
	ProfileID      *string
	QuizType       enum.QuizType
	QuestionsJSON  string
	PreviousQuizID *string
}

type CreateQuizCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateQuizCommandHandler(uow transaction.UnitOfWork) *CreateQuizCommandHandler {
	return &CreateQuizCommandHandler{uow: uow}
}

func (h *CreateQuizCommandHandler) Handle(ctx context.Context, cmd CreateQuizCommand) (*quiz.Quiz, error) {
	var created *quiz.Quiz

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		userUUID, err := utils.PtrStringToUUID(cmd.UserID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		profileUUID, err := utils.PtrStringToUUID(cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		q := quiz.NewQuiz()
		q.SetQuizId(utils.GenerateUUID())
		q.SetUserId(&userUUID)
		q.SetProfileId(&profileUUID)
		q.SetQuizType(string(cmd.QuizType))
		questions := cmd.QuestionsJSON
		q.SetQuestions(&questions)
		generated := string(enum.QuizStatusTypeGenerated)
		q.SetQuizStatus(&generated)
		if cmd.PreviousQuizID != nil {
			prevQuizUUID, err := utils.PtrStringToUUID(cmd.PreviousQuizID)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			q.SetPreviousQuizId(&prevQuizUUID)
		}

		saved, err := repos.Quiz.Create(ctx, q)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		created = saved
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return created, nil
}
