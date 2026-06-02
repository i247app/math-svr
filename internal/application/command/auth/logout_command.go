package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type LogoutCommand struct {
	UserID     int64
	DeviceUUID string
}

type LogoutCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewLogoutCommandHandler(uow transaction.UnitOfWork) *LogoutCommandHandler {
	return &LogoutCommandHandler{uow: uow}
}

func (h *LogoutCommandHandler) Handle(ctx context.Context, cmd LogoutCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		ll, err := repos.LoginLog.FindActiveByUserDevice(ctx, cmd.UserID, cmd.DeviceUUID)
		if err != nil {
			return errs.NewError(ctx, status.AUTH_LOGOUT_FAILED, nil, err)
		}
		if ll == nil {
			return errs.NewError(ctx, status.AUTH_INVALID_TOKEN, nil, ErrTokenNotFoundOrRevoked)
		}

		if err := repos.LoginLog.MarkStatusByLoginLogId(
			ctx, ll.LoginLogId(), enum.LoginLogStatusTypeRevoked,
		); err != nil {
			return errs.NewError(ctx, status.AUTH_LOGOUT_FAILED, nil, err)
		}
		return nil
	})
}
