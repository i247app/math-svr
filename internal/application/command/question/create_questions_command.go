package command

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/question"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// CreateQuestionsCommand stores a batch of questions in the bank, all or
// none. Questions come checked and hashed (module/question
// ValidateCreateQuestions), without ids.
type CreateQuestionsCommand struct {
	Questions []*question.Question
	ActorUID  *int64 // create_id; nil = no uid (API-key admin)
}

type CreateQuestionsCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateQuestionsCommandHandler(uow transaction.UnitOfWork) *CreateQuestionsCommandHandler {
	return &CreateQuestionsCommandHandler{uow: uow}
}

// Handle runs in one transaction: refuse what the bank already holds
// (naming the questions), mint one id per question, insert them in one
// statement, read them back. uk_live_question_hash is the real guard: a
// copy committed by another admin between the read and the INSERT still
// fails the INSERT (1062), and the whole batch rolls back.
//
// The result is in the order of cmd.Questions.
func (h *CreateQuestionsCommandHandler) Handle(ctx context.Context, cmd CreateQuestionsCommand) ([]*question.Question, error) {
	var saved []*question.Question
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		hashes := make([]string, 0, len(cmd.Questions))
		for _, q := range cmd.Questions {
			hashes = append(hashes, q.QuestionHash())
		}
		held, err := repos.Question.FindLiveHashes(ctx, hashes)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(held) > 0 {
			var at []string
			for i, h := range hashes {
				if slices.Contains(held, h) {
					at = append(at, fmt.Sprintf("questions[%d]", i))
				}
			}
			return errs.NewError(ctx, status.QUESTION_DUPLICATE, nil,
				fmt.Errorf("already in the question bank: %s", strings.Join(at, ", ")))
		}

		ids := make([]int64, 0, len(cmd.Questions))
		for _, q := range cmd.Questions {
			id, err := seqgen.Next(ctx, repos.Seq, seq.NameQuestion)
			if err != nil {
				return err
			}
			q.SetQuestionId(id)
			ids = append(ids, id)
		}
		if err := repos.Question.CreateMany(ctx, cmd.Questions, cmd.ActorUID); err != nil {
			if errors.Is(err, question.ErrDuplicate) {
				return errs.NewError(ctx, status.QUESTION_DUPLICATE, nil,
					fmt.Errorf("already in the question bank (stored by another request meanwhile): %w", err))
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		rows, err := repos.Question.ListByIds(ctx, ids)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		byId := make(map[int64]*question.Question, len(rows))
		for _, q := range rows {
			byId[q.QuestionId()] = q
		}
		saved = make([]*question.Question, 0, len(ids))
		for _, id := range ids {
			q, ok := byId[id]
			if !ok {
				return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("question: inserted %d but cannot read it back", id))
			}
			saved = append(saved, q)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
