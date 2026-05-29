package utils

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidPhone is returned by NormalizePhone when the input cannot be
// resolved to a canonical E.164 number under the rules below.
var ErrInvalidPhone = errors.New("invalid phone number")

// phoneSeparatorsRe matches characters that may decorate a human-typed
// phone number but carry no semantic value. They're stripped before any
// format check. URL-decoding (e.g. "+" arriving as " " from form-encoded
// bodies) is the caller's job — see auth/service.go's url.QueryUnescape.
var phoneSeparatorsRe = regexp.MustCompile(`[\s\-().]`)

// vnLocalRe — Vietnamese local format: leading 0 then 9 digits (10 total).
var vnLocalRe = regexp.MustCompile(`^0[0-9]{9}$`)

// vnE164BodyRe — Vietnamese E.164 without the leading "+": 84 + 9 digits.
var vnE164BodyRe = regexp.MustCompile(`^84[0-9]{9}$`)

// generalE164Re — any other E.164 number with explicit "+": 8–15 digits.
var generalE164Re = regexp.MustCompile(`^\+[0-9]{8,15}$`)

func ValidateEmail(email string) bool {
	r, err := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		return false
	}

	return r.MatchString(email)
}

func ValidatePhone(phone string) bool {
	phone = strings.TrimSpace(phone)

	r, err := regexp.Compile(`^\+?[0-9]{8,15}$`)
	if err != nil {
		return false
	}
	return r.MatchString(phone)
}

// NormalizePhone converts a client-supplied phone string into canonical
// E.164 form (e.g. "+84702465222"). It first trims whitespace and strips
// common human-typed separators (spaces, dashes, parens, dots), then
// classifies the cleaned digits:
//
//	"0702465222"      -> "+84702465222"   (VN local — leading 0 swapped for +84)
//	"84702465222"     -> "+84702465222"   (VN E.164 missing the +)
//	"+84702465222"    -> "+84702465222"   (already canonical)
//	"+15551234567"    -> "+15551234567"   (any other E.164)
//	"070 246 5222"    -> "+84702465222"   (separators tolerated everywhere)
//
// Anything else (empty after cleaning, wrong length, non-digit body)
// returns ErrInvalidPhone. Call this at every boundary that ingests a
// phone from the client so the rest of the system sees one format.
func NormalizePhone(phone string) (string, error) {
	cleaned := phoneSeparatorsRe.ReplaceAllString(strings.TrimSpace(phone), "")
	if cleaned == "" {
		return "", ErrInvalidPhone
	}

	switch {
	case vnLocalRe.MatchString(cleaned):
		return "+84" + cleaned[1:], nil
	case vnE164BodyRe.MatchString(cleaned):
		return "+" + cleaned, nil
	case generalE164Re.MatchString(cleaned):
		return cleaned, nil
	default:
		return "", ErrInvalidPhone
	}
}
