package utils

import "regexp"

func ValidateEmail(email string) bool {
	r, err := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		return false
	}

	return r.MatchString(email)
}

func ValidatePhone(phone string) bool {
	r, err := regexp.Compile(`^[0-9]{10}$`)
	if err != nil {
		return false
	}

	return r.MatchString(phone)
}
