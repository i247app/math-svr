package locales

import (
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

func GetArgsByStatatus(statusCode status.Code) []string {
	switch statusCode {
	case status.USER_INVALID_ROLE:
		return []string{string(enum.RoleUser), string(enum.RoleAdmin), string(enum.RoleGuest)}
	default:
		return []string{}
	}
}
