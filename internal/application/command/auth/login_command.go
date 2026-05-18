package command

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/loginlog"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

const loginTokenBytes = 32

type LoginCommand struct {
	Phone           string
	DeviceUUID      string
	IPAddress       string
	DevicePushToken string
}

type LoginCommandResult struct {
	UserID     uuid.UUID
	LoginLogID uuid.UUID
}

type LoginCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewLoginCommandHandler(uow transaction.UnitOfWork) *LoginCommandHandler {
	return &LoginCommandHandler{uow: uow}
}

// Handle resolves the user by phone, revokes any prior active session for the
// same (user, device), and inserts a fresh login_log row carrying the new
// opaque token. The entire sequence runs inside one UoW so a partial failure
// cannot leave two active sessions for the device.
func (h *LoginCommandHandler) Handle(ctx context.Context, cmd LoginCommand) (*LoginCommandResult, error) {
	var result *LoginCommandResult

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		u, err := repos.User.FindByPhone(ctx, cmd.Phone)
		if err != nil {
			return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		}
		if u == nil {
			return errs.NewError(ctx, status.USER_NOT_FOUND, nil, errors.New("user not found"))
		}

		if err := repos.LoginLog.MarkStatusByUserDevice(
			ctx, u.UserId(), cmd.DeviceUUID, enum.LoginLogStatusTypeRevoked,
		); err != nil {
			return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		}

		ll := BuildLoginLog(u.UserId(), cmd)

		_, err = repos.LoginLog.Create(ctx, ll)
		if err != nil {
			return errs.NewError(ctx, status.AUTH_LOGIN_FAILED, nil, err)
		}

		result = &LoginCommandResult{
			UserID: u.UserId(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func BuildLoginLog(userId uuid.UUID, cmd LoginCommand) *loginlog.LoginLog {
	ll := loginlog.NewLoginLog()
	ll.SetLoginLogId(utils.GenerateUUID())
	ll.SetUserId(userId)
	ll.SetDeviceUUID(cmd.DeviceUUID)
	ll.SetIpAddress(cmd.IPAddress)
	ll.SetToken(cmd.DevicePushToken)
	active := enum.LoginLogStatusTypeActive.String()
	ll.SetLoginLogStatus(&active)
	ll.SetStatus(enum.StatusActive.String())
	return ll
}

func ValidateLoginCommand(ctx context.Context, cmd LoginCommand) error {
	if strings.TrimSpace(cmd.Phone) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_PHONE, nil, errors.New("phone is required"))
	}
	if strings.TrimSpace(cmd.DeviceUUID) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_DEVICE_UUID, nil, errors.New("device_uuid is required"))
	}
	if strings.TrimSpace(cmd.IPAddress) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_IP_ADDRESS, nil, errors.New("ip_address is required"))
	}
	if strings.TrimSpace(cmd.DevicePushToken) == "" {
		return errs.NewError(ctx, status.AUTH_MISSING_DEVICE_PUSH_TOKEN, nil, errors.New("device_push_token is required"))
	}
	return nil
}
