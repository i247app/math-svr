package query

import (
	"context"

	"github.com/google/uuid"

	"math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type ListDevicesByUserIdQuery struct {
	UserID uuid.UUID
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
