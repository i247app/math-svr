package quiz

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/dto/question"
	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// The JSON parsing, salvage and normalization that used to live in this
// file now belong to internal/application/dto/question — the neutral home
// for the MCQ contract, so the exercise module (and the exam module that
// replaces this one) can share it without importing a sibling module.
// What stays here is the thin module-layer glue: turning the parser's
// structured warnings into log lines, and the DB-column clamps.

// normalizeGeneratedQuestions clamps each question's render type and logs
// any icon-token drift the parser reported. Drift never fails generation —
// grading is label-based and unaffected by icon content.
func normalizeGeneratedQuestions(ctx context.Context, questions []quizDto.QuizQuestion) []quizDto.QuizQuestion {
	out, warnings := question.Normalize(questions)
	if len(warnings) > 0 {
		log := logger.From(ctx)
		for _, w := range warnings {
			log.Warnf("quiz.normalize.%s value=%q q=%d", w.Kind, w.Value, w.QuestionNumber)
		}
	}
	return out
}

func normalizeLanguage(lang enum.LanguageType) string {
	s := strings.ToLower(strings.TrimSpace(string(lang)))
	if s == "" {
		return string(enum.LanguageTypeEnglish)
	}
	return s
}

// parseGeneration adapts the shared parser to the tuple this module's bot
// client consumes.
func parseGeneration(content string) (title string, shortText string, assessmentGrade string, questions []quizDto.QuizQuestion, err error) {
	gen, err := question.ParseGeneration(content)
	if err != nil {
		return "", "", "", nil, err
	}
	return gen.Title, gen.ShortText, gen.AssessmentGrade, gen.Questions, nil
}

func parseGradedQuiz(content string) (*quizDto.QuizGradingResult, error) {
	return question.ParseGrading(content)
}

// ParseGradedQuiz is kept exported for grading_contract_test.go, which
// pins the "review" JSON key against the prompt templates.
func ParseGradedQuiz(content string) (*quizDto.QuizGradingResult, error) {
	return parseGradedQuiz(content)
}

// maxQuizTextLen mirrors the VARCHAR(255) limit on ma_quizzes.title and
// ma_quizzes.short_text. The prompt asks the model for <= 80 characters;
// the clamp here is a defensive backstop against drift, not a primary
// enforcement point.
const maxQuizTextLen = 255

// sanitizeQuizText trims whitespace and clamps an AI-generated text field
// (title or short_text) to the DB column's rune budget. Returns nil when
// nothing usable is left so the row stores a real NULL (which
// DomainToResponse then omits).
func sanitizeQuizText(text string) *string {
	t := strings.TrimSpace(text)
	if t == "" {
		return nil
	}
	if runes := []rune(t); len(runes) > maxQuizTextLen {
		t = string(runes[:maxQuizTextLen])
	}
	return &t
}
