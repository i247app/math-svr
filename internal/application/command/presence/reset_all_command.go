package presence

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type ResetAllCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewResetAllCommandHandler(uow transaction.UnitOfWork) *ResetAllCommandHandler {
	return &ResetAllCommandHandler{uow: uow}
}

// Handle clears every stale counter. Run once at boot, before the server
// starts accepting connections: the Hub's registry is process memory, so every
// non-zero counter left in the table refers to a connection that died with the
// previous process.
func (h *ResetAllCommandHandler) Handle(ctx context.Context) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Presence.ResetAll(ctx); err != nil {
			return errs.NewError(ctx, status.PRESENCE_UPDATE_FAILED, nil, err)
		}
		return nil
	})
}
