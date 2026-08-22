// Package presence holds the write side of realtime availability. Both
// commands are driven by the socket runtime — one per connection open and
// one per connection close — not by an HTTP route.
package presence

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/presence"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type MarkOnlineCommand struct {
	UserId     int64
	DeviceUuid *string
	Platform   *string
}

type MarkOnlineCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkOnlineCommandHandler(uow transaction.UnitOfWork) *MarkOnlineCommandHandler {
	return &MarkOnlineCommandHandler{uow: uow}
}

// Handle adds one live connection for the user and returns the resulting row.
// The caller compares ConnectionCount to decide whether this was the
// off→online transition (count == 1) and therefore worth broadcasting.
func (h *MarkOnlineCommandHandler) Handle(ctx context.Context, cmd *MarkOnlineCommand) (*domain.Presence, error) {
	var result *domain.Presence

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		p, err := repos.Presence.IncrementConnection(ctx, cmd.UserId, cmd.DeviceUuid, cmd.Platform, mtime.Now())
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
