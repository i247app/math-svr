package loginlog

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByLoginLogId(ctx context.Context, loginLogId int64) (*LoginLog, error)
	FindActiveByToken(ctx context.Context, token string) (*LoginLog, error)
	FindActiveByUserDevice(ctx context.Context, userId int64, deviceUUID string) (*LoginLog, error)
	ListByUserId(ctx context.Context, userId int64) ([]*LoginLog, error)
	Create(ctx context.Context, loginLog *LoginLog) (*LoginLog, error)
	MarkStatusByLoginLogId(ctx context.Context, loginLogId int64, status enum.LoginLogStatusType) error
	MarkStatusByUserDevice(ctx context.Context, userId int64, deviceUUID string, status enum.LoginLogStatusType) error
	SoftDeleteByLoginLogId(ctx context.Context, loginLogId int64) error
}
