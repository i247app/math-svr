package home

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/home"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ValidateHomeLayout enforces the single hard requirement: a profile_id
// must be supplied so the service knows which role to render for.
func ValidateHomeLayout(ctx context.Context, req *dto.HomeLayoutReq) error {
	if req == nil || req.ProfileID == 0 {
		return errs.NewError(ctx, status.HOME_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	return nil
}
