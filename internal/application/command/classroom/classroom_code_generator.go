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

// maxClassroomCodeAttempts bounds how many times mintUniqueClassroomCode
// will retry on a uniqueness collision. The keyspace is 26*26*10000 ≈
// 6.76M, so even with hundreds of thousands of live codes a handful of
// tries is more than enough; the cap is here to keep a pathological DB
// state from looping forever rather than to plan for collisions.
const maxClassroomCodeAttempts = 5

// classroomCodeLetters is the alphabet for the two-letter prefix.
// Limited to the 26 uppercase ASCII letters; ambiguity-prone characters
// are not stripped because the prefix is only two chars and parents
// type it on a normal keyboard, not a numeric pad.
const classroomCodeLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateClassroomCode returns a single random "AA-1111" candidate
// (two uppercase letters, dash, four digits). Uses crypto/rand directly
// — math/rand would be predictable across processes and a guessable
// code defeats the purpose of a join-code gate. Caller is responsible
// for uniqueness; collisions are extremely unlikely (≈ 6.76M space) but
// the create path retries on conflict to make that guarantee hard.
func generateClassroomCode() (string, error) {
	letterMax := big.NewInt(int64(len(classroomCodeLetters)))

	l1, err := rand.Int(rand.Reader, letterMax)
	if err != nil {
		return "", fmt.Errorf("classroom code: letter1: %w", err)
	}
	l2, err := rand.Int(rand.Reader, letterMax)
	if err != nil {
		return "", fmt.Errorf("classroom code: letter2: %w", err)
	}
	digitMax := big.NewInt(10000)
	d, err := rand.Int(rand.Reader, digitMax)
	if err != nil {
		return "", fmt.Errorf("classroom code: digits: %w", err)
	}

	return fmt.Sprintf("%c%c-%04d",
		classroomCodeLetters[l1.Int64()],
		classroomCodeLetters[l2.Int64()],
		d.Int64()), nil
}

// mintUniqueClassroomCode picks a fresh AA-1111 code that does not yet
// exist in ma_classrooms. Runs inside the caller's UoW so the
// FindByClassroomCode probe sees the same snapshot the subsequent
// INSERT will commit against. After maxClassroomCodeAttempts collisions
// or any crypto/rand error it returns CLASSROOM_CODE_GENERATION_FAILED
// so the caller surfaces a stable status code instead of a generic
// failure.
func mintUniqueClassroomCode(ctx context.Context, repos transaction.Repositories) (string, error) {
	for attempt := 0; attempt < maxClassroomCodeAttempts; attempt++ {
		candidate, err := generateClassroomCode()
		if err != nil {
			return "", errs.NewError(ctx, status.CLASSROOM_CODE_GENERATION_FAILED, nil, err)
		}
		existing, err := repos.Classroom.FindByClassroomCode(ctx, candidate)
		if err != nil {
			return "", errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return candidate, nil
		}
	}
	return "", errs.NewError(ctx, status.CLASSROOM_CODE_GENERATION_FAILED, nil,
		errors.New("could not mint a unique classroom code"))
}
