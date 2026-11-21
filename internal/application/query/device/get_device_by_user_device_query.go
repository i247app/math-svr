package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/shared/errs"
)

type GetDeviceByUserDeviceQuery struct {
	UID        int64
	DeviceUUID string
}

type GetDeviceByUserDeviceQueryHandler struct {
	repo device.IRepository
}

func NewGetDeviceByUserDeviceQueryHandler(repo device.IRepository) *GetDeviceByUserDeviceQueryHandler {
	return &GetDeviceByUserDeviceQueryHandler{repo: repo}
}

func (h *GetDeviceByUserDeviceQueryHandler) Handle(ctx context.Context, q GetDeviceByUserDeviceQuery) (*device.Device, error) {
	d, err := h.repo.FindByUserDevice(ctx, q.UID, q.DeviceUUID)
	if err != nil {
		return nil, errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
	}
	return d, nil
}
