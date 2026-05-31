package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type ListDevicesByUserIdQuery struct {
	UserID int64
}

type ListDevicesByUserIdQueryHandler struct {
	repo device.IRepository
}

func NewListDevicesByUserIdQueryHandler(repo device.IRepository) *ListDevicesByUserIdQueryHandler {
	return &ListDevicesByUserIdQueryHandler{repo: repo}
}

func (h *ListDevicesByUserIdQueryHandler) Handle(ctx context.Context, q ListDevicesByUserIdQuery) ([]*device.Device, error) {
	devices, err := h.repo.ListByUserId(ctx, q.UserID)
	if err != nil {
		return nil, errs.NewError(ctx, status.DEVICE_REGISTRATION_FAIL, nil, err)
	}
	return devices, nil
}
