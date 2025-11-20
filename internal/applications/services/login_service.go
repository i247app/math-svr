package services

import "math-ai.com/math-ai/internal/core/di/repositories"

type LoginService struct {
	loginRepo repositories.ILoginRepository
}

func NewLoginService(loginRepo repositories.ILoginRepository) *LoginService {
	return &LoginService{
		loginRepo: loginRepo,
	}
}
