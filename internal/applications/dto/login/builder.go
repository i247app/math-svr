package dto

import domain "math-ai.com/math-ai/internal/core/domain/login"

func BuildLoginDomain(uid string, password string) *domain.Login {
	loginDomain := domain.NewLoginDomain()
	loginDomain.GenerateID()
	loginDomain.SetUID(uid)
	loginDomain.SetHassPass(password)

	return loginDomain
}
