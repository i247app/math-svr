package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// nextSeqID mints a new external id via the central ma_seqs registry.
// Wraps repos.Seq.Next so command handlers don't import seq directly.
func nextSeqID(ctx context.Context, repos transaction.Repositories, name string) (int64, error) {
	id, err := repos.Seq.Next(ctx, name)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if id == 0 {
		return 0, errs.NewError(ctx, status.FAIL, nil, ErrSeqReturnedZeroID)
	}
	return id, nil
}
