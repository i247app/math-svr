package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/otp"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type GetOtpByIdQuery struct {
	OtpID int64
}

type GetOtpByIdQueryHandler struct {
	repo otp.IRepository
}

func NewGetOtpByIdQueryHandler(repo otp.IRepository) *GetOtpByIdQueryHandler {
	return &GetOtpByIdQueryHandler{repo: repo}
}

func (h *GetOtpByIdQueryHandler) Handle(ctx context.Context, q GetOtpByIdQuery) (*otp.Otp, error) {
	o, err := h.repo.FindByOtpId(ctx, q.OtpID)
	if err != nil {
		return nil, errs.NewError(ctx, status.OTP_NOT_FOUND, nil, err)
	}
	return o, nil
}
