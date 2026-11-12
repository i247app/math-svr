package sms

import (
	"errors"
	"regexp"
	"strings"
)

var (
	nonDigitRegex = regexp.MustCompile(`[^\d+]`)

	vnMobileRegex = regexp.MustCompile(`^[35789]\d{8}$`)

	usPhoneRegex = regexp.MustCompile(`^[2-9]\d{2}[2-9]\d{2}\d{4}$`)
)

func NormalizePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "", errors.New("phone number cannot be empty")
	}

	if strings.HasPrefix(phone, "+84") {
		nationalNumber := strings.TrimPrefix(phone, "+84")
		nationalNumber = strings.TrimPrefix(nationalNumber, "0")
		if vnMobileRegex.MatchString(nationalNumber) {
			return "+84" + nationalNumber, nil
		}
	}

	if strings.HasPrefix(phone, "+1") {
		nationalNumber := strings.TrimPrefix(phone, "+1")
		nationalNumber = nonDigitRegex.ReplaceAllString(nationalNumber, "")
		if usPhoneRegex.MatchString(nationalNumber) {
			return "+1" + nationalNumber, nil
		}
	}

	digitsOnly := nonDigitRegex.ReplaceAllString(phone, "")

	if len(digitsOnly) == 10 && strings.HasPrefix(digitsOnly, "0") {
		nationalNumber := digitsOnly[1:]
		if vnMobileRegex.MatchString(nationalNumber) {
			return "+84" + nationalNumber, nil
		}
	}

	if len(digitsOnly) == 11 && strings.HasPrefix(digitsOnly, "84") {
		nationalNumber := digitsOnly[2:]
		if vnMobileRegex.MatchString(nationalNumber) {
			return "+84" + nationalNumber, nil
		}
	}

	if len(digitsOnly) == 11 && strings.HasPrefix(digitsOnly, "1") {
		nationalNumber := digitsOnly[1:]
		if usPhoneRegex.MatchString(nationalNumber) {
			return "+1" + nationalNumber, nil
		}
	}

	if len(digitsOnly) == 10 {
		if usPhoneRegex.MatchString(digitsOnly) {
			return "+1" + digitsOnly, nil
		}
	}

	return "", errors.New("The phone number you provided appears to be invalid. Please check the number and try again.")
}
