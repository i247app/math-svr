package pow

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	err "math-ai.com/math-ai/internal/shared/errs"
)

func ValidateChallenge(ctx context.Context, req *dto.ChallengeReq) error {
	return nil
}

func ValidateVerifyChallenge(ctx context.Context, req *dto.VerifyChallengeReq) error {
	if req.Nonce <= 0 {
		return err.NewError(ctx, PowInvalidRequest, nil, ErrNonceRequired)
	}
	return nil
}
