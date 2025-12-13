package di

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type ILoginService interface {
	Login(ctx context.Context, sess *session.AppSession, req *dto.LoginRequest) (status.Code, *dto.LoginResponse, error)
	Logout(ctx context.Context, session *session.AppSession) (status.Code, error)
}
