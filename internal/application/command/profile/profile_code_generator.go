package command

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// maxProfileCodeAttempts bounds how many times mintUniqueProfileCode
// will retry on a uniqueness collision. The keyspace is 26*26*10000 ≈
// 6.76M, so even with hundreds of thousands of live profiles a handful
// of tries is more than enough; the cap is here to keep a pathological
// DB state from looping forever rather than to plan for collisions.
const maxProfileCodeAttempts = 5

// profileCodeLetters is the alphabet for the two-letter prefix.
// Limited to the 26 uppercase ASCII letters; ambiguity-prone characters
// are not stripped because the prefix is only two chars and users type
// it on a normal keyboard, not a numeric pad. Format matches the
// classroom_code generator so both surfaces share a familiar shape.
const profileCodeLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateProfileCode returns a single random "AA-1234" candidate (two
// uppercase letters, dash, four digits). Uses crypto/rand directly —
// math/rand would be predictable across processes and a guessable code
// undermines its value as a stable shareable identifier. Caller is
// responsible for uniqueness; collisions are extremely unlikely (≈
// 6.76M space) but the create path retries on conflict to make that
// guarantee hard.
func generateProfileCode() (string, error) {
	letterMax := big.NewInt(int64(len(profileCodeLetters)))

	l1, err := rand.Int(rand.Reader, letterMax)
	if err != nil {
		return "", fmt.Errorf("profile code: letter1: %w", err)
	}
	l2, err := rand.Int(rand.Reader, letterMax)
	if err != nil {
		return "", fmt.Errorf("profile code: letter2: %w", err)
	}
	digitMax := big.NewInt(10000)
	d, err := rand.Int(rand.Reader, digitMax)
	if err != nil {
		return "", fmt.Errorf("profile code: digits: %w", err)
	}

	return fmt.Sprintf("%c%c-%04d",
		profileCodeLetters[l1.Int64()],
		profileCodeLetters[l2.Int64()],
		d.Int64()), nil
}

// mintUniqueProfileCode picks a fresh AA-1234 code that does not yet
// exist in ma_profiles. Runs inside the caller's UoW so the
// FindByProfileCode probe sees the same snapshot the subsequent INSERT
// will commit against; the DB UNIQUE constraint on profile_code is the
// hard backstop for the race window between probe and write. After
// maxProfileCodeAttempts collisions or any crypto/rand error it
// returns PROFILE_CODE_GENERATION_FAILED so the caller surfaces a
// stable status code instead of a generic failure.
func mintUniqueProfileCode(ctx context.Context, repos transaction.Repositories) (string, error) {
	for attempt := 0; attempt < maxProfileCodeAttempts; attempt++ {
		candidate, err := generateProfileCode()
		if err != nil {
			return "", errs.NewError(ctx, status.PROFILE_CODE_GENERATION_FAILED, nil, err)
		}
		existing, err := repos.Profile.FindByProfileCode(ctx, candidate)
		if err != nil {
			return "", errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return candidate, nil
		}
	}
	return "", errs.NewError(ctx, status.PROFILE_CODE_GENERATION_FAILED, nil,
		errors.New("could not mint a unique profile code"))
}
