package loginlog

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindByLoginLogId(ctx context.Context, loginLogId string) (*LoginLog, error)
	FindActiveByToken(ctx context.Context, token string) (*LoginLog, error)
	FindActiveByUserDevice(ctx context.Context, userId string, deviceUUID string) (*LoginLog, error)
	ListByUserId(ctx context.Context, userId string) ([]*LoginLog, error)
	Create(ctx context.Context, loginLog *LoginLog) (*LoginLog, error)
	MarkStatusByLoginLogId(ctx context.Context, loginLogId string, status enum.LoginLogStatusType) error
	MarkStatusByUserDevice(ctx context.Context, userId string, deviceUUID string, status enum.LoginLogStatusType) error
	SoftDeleteByLoginLogId(ctx context.Context, loginLogId string) error
}
