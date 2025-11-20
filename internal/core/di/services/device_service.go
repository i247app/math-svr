package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type IDeviceService interface {
	GetDeviceByDeviceUUID(ctx context.Context, deviceUUID string) (status.Code, *dto.DeviceResponse, error)
	CreateDevice(ctx context.Context, req *dto.CreateDeviceReq) (status.Code, error)
	UpdateDevice(ctx context.Context, req *dto.UpdateDeviceReq) (status.Code, error)
	MarkVerifiedDevice(ctx context.Context, uid string, deviceUUID string) (status.Code, error)
}
