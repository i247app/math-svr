package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	"math-ai.com/math-ai/internal/shared/errs"
)

type GetDeviceByIdQuery struct {
	DeviceID int64
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
		return nil, errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
	}
	return d, nil
}
