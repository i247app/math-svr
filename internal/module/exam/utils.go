package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// maxExamTextLen mirrors the VARCHAR(255) limit on ma_ai_exams.ai_title
// and ai_short_text. The prompt asks for far less; this is a backstop
// against drift, not the primary enforcement.
const maxExamTextLen = 255

// sanitizeExamText trims and clamps an AI-generated text field, returning
// nil when nothing usable is left so the column stores a real NULL.
func sanitizeExamText(text string) *string {
	t := strings.TrimSpace(text)
	if t == "" {
		return nil
	}
	if runes := []rune(t); len(runes) > maxExamTextLen {
		t = string(runes[:maxExamTextLen])
	}
	return &t
}

// marshalQuestions stores the round in the EXAM vocabulary, so what sits
// in ai_questions_json is what the client reads and what
// ma_user_exam_details mirrors — one set of names end to end.
func marshalQuestions(ctx context.Context, questions []question.Question) (string, error) {
	raw, err := json.Marshal(dto.FromSharedQuestions(questions))
	if err != nil {
		return "", errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil,
			fmt.Errorf("exam: marshal questions: %w", err))
	}
	return string(raw), nil
}
