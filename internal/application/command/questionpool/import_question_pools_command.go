package command

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ImportQuestionPoolsCommand stores the questions of an imported file.
// They come checked, hashed and free of duplicates among themselves
// (module/questionpool import.go), without ids.
type ImportQuestionPoolsCommand struct {
	QuestionPools []*questionpool.QuestionPool
	ActorUID      *int64 // create_id; nil = no uid (API-key admin)
}

// ImportQuestionPoolsResult says, by position in cmd.QuestionPools, what
// was stored (with its new id set on the question) and what was skipped
// because the pool already held it when the transaction ran.
type ImportQuestionPoolsResult struct {
	Imported []int
	Skipped  []int
}

type ImportQuestionPoolsCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewImportQuestionPoolsCommandHandler(uow transaction.UnitOfWork) *ImportQuestionPoolsCommandHandler {
	return &ImportQuestionPoolsCommandHandler{uow: uow}
}

// Handle runs in ONE transaction, so a file is stored whole or not at all:
//
//  1. lock the ma_question_pools seq row first, so two imports queue and
//     the second sees what the first committed (its copies are skipped
//     instead of failing the INSERT);
//  2. skip what a live row already holds (checked again here: the module's
//     check ran outside the transaction);
//  3. mint every id in one step and insert in chunks (CreateMany).
//
// /question-pools/create takes no seq lock, so the same content created
// there at the same moment still fails the INSERT (1062): the whole import
// rolls back with QUESTION_POOL_DUPLICATE, and running it again skips it.
func (h *ImportQuestionPoolsCommandHandler) Handle(ctx context.Context, cmd ImportQuestionPoolsCommand) (*ImportQuestionPoolsResult, error) {
	var res ImportQuestionPoolsResult
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		res = ImportQuestionPoolsResult{}
		if err := seqgen.Lock(ctx, repos.Seq, seq.NameQuestionPool); err != nil {
			return err
		}

		hashes := make([]string, 0, len(cmd.QuestionPools))
		for _, q := range cmd.QuestionPools {
			hashes = append(hashes, q.QuestionHash())
		}
		held, err := repos.QuestionPool.FindLiveHashes(ctx, hashes)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		isHeld := make(map[string]bool, len(held))
		for _, h := range held {
			isHeld[h] = true
		}
		toInsert := make([]*questionpool.QuestionPool, 0, len(cmd.QuestionPools))
		for i, q := range cmd.QuestionPools {
			if isHeld[q.QuestionHash()] {
				res.Skipped = append(res.Skipped, i)
				continue
			}
			res.Imported = append(res.Imported, i)
			toInsert = append(toInsert, q)
		}
		if len(toInsert) == 0 {
			return nil
		}

		ids, err := seqgen.NextN(ctx, repos.Seq, seq.NameQuestionPool, len(toInsert))
		if err != nil {
			return err
		}
		for i, q := range toInsert {
			q.SetQuestionPoolId(ids[i])
		}
		if err := repos.QuestionPool.CreateMany(ctx, toInsert, cmd.ActorUID); err != nil {
			if errors.Is(err, questionpool.ErrDuplicate) {
				return errs.NewError(ctx, status.QUESTION_POOL_DUPLICATE, nil,
					fmt.Errorf("already in the question bank (stored by another request meanwhile): %w", err))
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}
