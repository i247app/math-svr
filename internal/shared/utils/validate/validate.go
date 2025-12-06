package validate

import (
	"regexp"

	"math-ai.com/math-ai/internal/shared/constant/enum"
)

func IsValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func IsValidPhoneNumber(phone string) bool {
	r, e := regexp.Compile(`^\+?[\d\s()-]{7,20}$`)
	if e != nil {
		return false
	}

	return r.MatchString(phone)
}

func IsValidRole(role enum.ERole) bool {
	validRoles := []enum.ERole{enum.RoleAdmin, enum.RoleUser, enum.RoleGuest}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}
