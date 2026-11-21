package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/shared/errs"
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
		return nil, errs.NewError(ctx, otp.OTPNotFound, nil, err)
	}
	return o, nil
}
