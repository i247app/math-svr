package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"math-ai.com/math-ai/internal/shared/utils"
)

// Solves reports whether SHA-256(seed + nonce), in hex, starts with
// difficulty zeros.
func Solves(seed string, nonce int64, difficulty int) bool {
	sum := sha256.Sum256([]byte(seed + utils.Int64ToString(nonce)))
	return strings.HasPrefix(hex.EncodeToString(sum[:]), strings.Repeat("0", difficulty))
}
