package command

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateQuestionCommand edits one question. Apply turns the stored row into
// the edited one — merging the request and checking the result — and
// returns a MathError when the edit is refused. It runs on the row read
// under lock, so it always merges onto the latest version.
type UpdateQuestionCommand struct {
	QuestionID int64
	ActorUID   *int64 // modify_id; nil = no uid (API-key admin)
	Apply      func(stored *question.Question) (*question.Question, error)
}

type UpdateQuestionCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateQuestionCommandHandler(uow transaction.UnitOfWork) *UpdateQuestionCommandHandler {
	return &UpdateQuestionCommandHandler{uow: uow}
}

// Handle locks the row first, so two admins editing the same question are
// serialised and the second merges onto the first's result instead of
// overwriting it. A new content that another live question already holds
// is refused (uk_live_question_hash).
func (h *UpdateQuestionCommandHandler) Handle(ctx context.Context, cmd UpdateQuestionCommand) (*question.Question, error) {
	var saved *question.Question
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		stored, err := repos.Question.LockById(ctx, cmd.QuestionID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if stored == nil {
			return errs.NewError(ctx, status.QUESTION_NOT_FOUND, nil,
				fmt.Errorf("question: no live question %d", cmd.QuestionID))
		}
		edited, err := cmd.Apply(stored)
		if err != nil {
			return err
		}
		edited.SetQuestionId(stored.QuestionId())
		if err := repos.Question.UpdateContent(ctx, edited, cmd.ActorUID); err != nil {
			if errors.Is(err, question.ErrDuplicate) {
				return errs.NewError(ctx, status.QUESTION_DUPLICATE, nil,
					fmt.Errorf("another question in the bank already has this content: %w", err))
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved, err = repos.Question.LockById(ctx, cmd.QuestionID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("question: updated %d but cannot read it back", cmd.QuestionID))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
