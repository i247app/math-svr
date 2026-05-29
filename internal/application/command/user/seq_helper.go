package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// nextSeqID mints the next centralised ID for the given sequence inside
// the surrounding UoW. Wrapping happens here so each command handler
// stays focused on its aggregate logic and the MathError mapping is
// consistent — repo-level seq.ErrNotFound becomes SEQ_NOT_FOUND, any
// other failure becomes SEQ_GENERATION_FAILED.
func nextSeqID(ctx context.Context, repos transaction.Repositories, name string) (string, error) {
	id, err := repos.Seq.Next(ctx, name)
	if err != nil {
		if errors.Is(err, seq.ErrNotFound) {
			return "", errs.NewError(ctx, status.SEQ_NOT_FOUND, map[string]any{"name": name}, err)
		}
		return "", errs.NewError(ctx, status.SEQ_GENERATION_FAILED, map[string]any{"name": name}, err)
	}
	return id, nil
}
