package presence

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/presence"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type MarkOfflineCommand struct {
	UserId int64
}

type MarkOfflineCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkOfflineCommandHandler(uow transaction.UnitOfWork) *MarkOfflineCommandHandler {
	return &MarkOfflineCommandHandler{uow: uow}
}

// Handle removes one live connection. The returned row's ConnectionCount is
// zero when the user's last device disconnected — the only case that changes
// what other people see.
func (h *MarkOfflineCommandHandler) Handle(ctx context.Context, cmd *MarkOfflineCommand) (*domain.Presence, error) {
	var result *domain.Presence

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		p, err := repos.Presence.DecrementConnection(ctx, cmd.UserId, mtime.Now())
		if err != nil {
			return errs.NewError(ctx, status.PRESENCE_UPDATE_FAILED, nil, err)
		}
		result = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
