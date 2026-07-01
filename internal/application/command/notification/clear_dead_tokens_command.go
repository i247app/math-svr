package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ClearDeadTokensCommand prunes device_push_token for tokens FCM reported as
// dead (unregistered / invalid) during a push send, so they are not retried.
// Runs in its own short UoW after the (out-of-tx) push call returns.
type ClearDeadTokensCommand struct {
	Tokens []string
}

type ClearDeadTokensCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewClearDeadTokensCommandHandler(uow transaction.UnitOfWork) *ClearDeadTokensCommandHandler {
	return &ClearDeadTokensCommandHandler{uow: uow}
}

func (h *ClearDeadTokensCommandHandler) Handle(ctx context.Context, cmd ClearDeadTokensCommand) error {
	if len(cmd.Tokens) == 0 {
		return nil
	}
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Device.ClearPushTokens(ctx, cmd.Tokens); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
