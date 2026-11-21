package home

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/shared/errs"
)

// ValidateHomeLayout enforces the single hard requirement: a profile_id
// must be supplied so the service knows which role to render for.
func ValidateHomeLayout(ctx context.Context, req *dto.HomeLayoutReq) error {
	if req == nil || req.ProfileID == 0 {
		return errs.NewError(ctx, HomeMissingProfileID, nil, ErrProfileIDRequired)
	}
	return nil
}
