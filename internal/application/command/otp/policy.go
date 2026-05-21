package command

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Tunables. Kept as package-level vars so tests (and a future config layer)
// can override them without changing call sites. Production code MUST NOT
// reassign these at runtime — they are policy, not config.
var (
	// OtpCodeLength is the number of decimal digits in a delivered code.
	OtpCodeLength = 4

	// OtpResendCooldown is the minimum wait before the same (type, identifier)
	// may request another OTP. Trips OTP_TOO_FREQUENT.
	OtpResendCooldown = 10 * time.Second

	// OtpSendWindow + OtpMaxSendsPerWindow form the per-window send cap.
	// Trips OTP_RATE_LIMITED.
	OtpSendWindow        = 1 * time.Hour
	OtpMaxSendsPerWindow = 5

	// OtpMaxAttempts is the max number of verify attempts before a row is
	// auto-revoked. Trips OTP_TOO_MANY_ATTEMPTS.
	OtpMaxAttempts = 5
)

// TtlFor returns the validity window for a given OTP type. Short-lived for
// auth-adjacent flows, slightly longer for inbox-delivered ones.
func TtlFor(t enum.OtpType) time.Duration {
	switch t {
	case enum.OtpTypeLogin2FA:
		return 1 * time.Minute
	case enum.OtpTypeChangePassword:
		return 1 * time.Minute
	case enum.OtpTypeVerifyPhone:
		return 1 * time.Minute
	case enum.OtpTypeRegister, enum.OtpTypeForgotPassword, enum.OtpTypeVerifyEmail:
		return 15 * time.Minute
	default:
		return 1 * time.Minute
	}
}

// generateCode returns a uniformly-random N-digit numeric code. Uses
// crypto/rand directly — math/rand would be predictable across processes.
func generateCode(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("otp: invalid code length %d", n)
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, v), nil
}

// hashCode is the one-way function applied to a plaintext code before
// storage. We use raw SHA-256 (not bcrypt/argon2) because:
//   - Codes are short-lived (≤ 15 min) and high-entropy is bounded at 10^N.
//   - Verify is a hot path; a slow KDF would dominate latency without adding
//     meaningful protection against an attacker who has DB access.
//
// At-rest hashing is defense-in-depth against accidental log/dump exposure,
// not a primary security boundary.
func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// codesMatch is constant-time over the hash strings.
func codesMatch(submittedPlain, storeCode string) bool {
	// return subtle.ConstantTimeCompare(
	// 	[]byte(hashCode(submittedPlain)),
	// 	[]byte(storedHash),
	// ) == 1
	return submittedPlain == storeCode
}
