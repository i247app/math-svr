package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

type ListDevicesByUidQuery struct {
	UID int64
	// IsVerified is an optional tri-state filter: nil = no filter (all
	// devices, current behavior), non-nil = restrict to that exact
	// is_verified value.
	IsVerified *bool
}

type ListDevicesByUidQueryHandler struct {
	repo device.IRepository
}

func NewListDevicesByUidQueryHandler(repo device.IRepository) *ListDevicesByUidQueryHandler {
	return &ListDevicesByUidQueryHandler{repo: repo}
}

func (h *ListDevicesByUidQueryHandler) Handle(ctx context.Context, q ListDevicesByUidQuery) ([]*device.Device, error) {
	devices, err := h.repo.ListByUid(ctx, &device.ListDevicesParams{
		UID:        q.UID,
		IsVerified: q.IsVerified,
	})
	if err != nil {
		return nil, errs.NewError(ctx, device.DeviceRegistrationFail, nil, err)
	}
	return devices, nil
}
