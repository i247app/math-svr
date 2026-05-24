package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type GetDeviceByIdQuery struct {
	DeviceID string
}

type GetDeviceByIdQueryHandler struct {
	repo device.IRepository
}

func NewGetDeviceByIdQueryHandler(repo device.IRepository) *GetDeviceByIdQueryHandler {
	return &GetDeviceByIdQueryHandler{repo: repo}
}

func (h *GetDeviceByIdQueryHandler) Handle(ctx context.Context, q GetDeviceByIdQuery) (*device.Device, error) {
	d, err := h.repo.FindByDeviceId(ctx, q.DeviceID)
	if err != nil {
		return nil, errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
	}
	return d, nil
}
