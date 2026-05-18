package loginlog

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByLoginLogId(ctx context.Context, loginLogId uuid.UUID) (*LoginLog, error)
	FindActiveByToken(ctx context.Context, token string) (*LoginLog, error)
	FindActiveByUserDevice(ctx context.Context, userId uuid.UUID, deviceUUID string) (*LoginLog, error)
	ListByUserId(ctx context.Context, userId uuid.UUID) ([]*LoginLog, error)
	Create(ctx context.Context, loginLog *LoginLog) (*LoginLog, error)
	MarkStatusByLoginLogId(ctx context.Context, loginLogId uuid.UUID, status enum.LoginLogStatusType) error
	MarkStatusByUserDevice(ctx context.Context, userId uuid.UUID, deviceUUID string, status enum.LoginLogStatusType) error
	SoftDeleteByLoginLogId(ctx context.Context, loginLogId uuid.UUID) error
}
