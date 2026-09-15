package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Shared plumbing for the prompt builders in this package. The exam
// prompts (exam_prompts.go) and the classroom-exercise prompts
// (exercise_prompts.go) each own their templates; what they have in
// common — the language switch, the required-field check, the default
// length — lives here.
//
// The QuizLanguage name predates the exam module: it was the response
// language of the old quiz prompts, and the exercise and grade-profile
// code adopted it. The quiz prompts are gone; the type stays under its
// old name because a rename would touch every builder for no behaviour.

// QuizLanguage is the response language requested from the model. The
// model is instructed to write prose in this language; JSON keys are
// always English regardless.
type QuizLanguage string

const (
	QuizLanguageVietnamese QuizLanguage = "vn"
	QuizLanguageEnglish    QuizLanguage = "en"
)

func normalizeLanguage(lang QuizLanguage) (QuizLanguage, error) {
	switch strings.ToLower(strings.TrimSpace(string(lang))) {
	case string(enum.LanguageTypeVietnamese), string(enum.LanguageTypeVietnameseV2):
		return QuizLanguageVietnamese, nil
	case string(enum.LanguageTypeEnglish), string(enum.LanguageTypeEnglishV2):
		return QuizLanguageEnglish, nil
	default:
		return "", fmt.Errorf("bot: unsupported language %q", string(lang))
	}
}

// requireFields takes alternating (value, name) pairs and returns the
// first missing-field error encountered. Keeps the per-kind validation
// readable at the call site.
func requireFields(pairs ...string) error {
	if len(pairs)%2 != 0 {
		return fmt.Errorf("bot: requireFields expects (value, name) pairs")
	}
	for i := 0; i < len(pairs); i += 2 {
		if strings.TrimSpace(pairs[i]) == "" {
			return fmt.Errorf("bot: %s is required", pairs[i+1])
		}
	}
	return nil
}

// defaultNumQuestions is the fallback used when a caller leaves the
// question count at zero. Kept in sync with the module validators'
// DefaultNumQuestions (10), but duplicated here so the domain builders
// remain usable without going through the module layer.
const defaultNumQuestions = 10

func resolveNumQuestions(n int) int {
	if n <= 0 {
		return defaultNumQuestions
	}
	return n
}
