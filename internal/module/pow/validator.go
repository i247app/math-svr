package pow

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/pow"
	err "math-ai.com/math-ai/internal/domain/shared/error"
	status "math-ai.com/math-ai/internal/domain/shared/status"
)

func ValidateChallenge(ctx context.Context, req *dto.ChallengeReq) error {
	return nil
}

func ValidateVerifyChallenge(ctx context.Context, req *dto.VerifyChallengeReq) error {
	if req.Nonce <= 0 {
		return err.NewError(ctx, status.POW_INVALID_REQUEST, nil, ErrNonceRequired)
	}
	return nil
}
