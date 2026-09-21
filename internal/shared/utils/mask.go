package utils

import (
	"strings"
)

func MaskIndentifier(s string) string {
	if ValidateEmail(s) {
		return MaskEmail(s)
	}
	if ValidatePhone(s) {
		return MaskPhone(s)
	}
	return s
}

// MaskPhone returns a log-safe rendering of an E.164 phone number that
// preserves the leading "+", the country-code prefix, and the last four
// digits while replacing the middle with "***". Inputs too short to
// mask are returned unchanged — empty input returns empty.
//
// Example: "+15551234567" → "+1***4567".
//
// Lives in the libs package because phone-PII masking is a Twilio-payload
// concern; the adapter layer re-exports it via package-level usage.
func MaskPhone(p string) string {
	if p == "" {
		return ""
	}
	// Need at least "+" + 1 cc digit + 4 trailing digits + 1 hidden char.
	if len(p) < 7 {
		return p
	}
	if !strings.HasPrefix(p, "+") {
		return p
	}
	// Keep the "+", the first cc digit (1–3 chars commonly, but we keep 1
	// for safety since the country code length isn't trivially derivable
	// from the prefix alone), and the last 4 digits.
	keepHead := 2 // "+" + first digit
	keepTail := 4
	if len(p) <= keepHead+keepTail {
		return p
	}
	return p[:keepHead] + "***" + p[len(p)-keepTail:]
}

// MaskEmail returns a log-safe rendering of an email address that
// preserves the local part and the domain suffix while replacing the middle with "***". Inputs too short to
// mask are returned unchanged — empty input returns empty.
//
// Example: "[EMAIL_ADDRESS]" → "[EMAIL_ADDRESS]".
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email // Invalid email format, return as is
	}
	localPart := parts[0]
	domainPart := parts[1]

	// Mask the local part
	if len(localPart) > 2 {
		localPart = localPart[:2] + "***"
	}

	// Mask the domain part (keep the last dot and the TLD)
	lastDotIndex := strings.LastIndex(domainPart, ".")
	if lastDotIndex != -1 && len(domainPart)-lastDotIndex > 1 {
		domainPart = domainPart[:lastDotIndex+1] + "***"
	}

	return localPart + "@" + domainPart
}
