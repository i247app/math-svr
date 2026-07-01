package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type GetDeviceByUserDeviceQuery struct {
	UserID     int64
	DeviceUUID string
}

type GetDeviceByUserDeviceQueryHandler struct {
	repo device.IRepository
}

func NewGetDeviceByUserDeviceQueryHandler(repo device.IRepository) *GetDeviceByUserDeviceQueryHandler {
	return &GetDeviceByUserDeviceQueryHandler{repo: repo}
}

func (h *GetDeviceByUserDeviceQueryHandler) Handle(ctx context.Context, q GetDeviceByUserDeviceQuery) (*device.Device, error) {
	d, err := h.repo.FindByUserDevice(ctx, q.UserID, q.DeviceUUID)
	if err != nil {
		return nil, errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
	}
	return d, nil
}
