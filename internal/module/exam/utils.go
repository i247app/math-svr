package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"

	command "math-ai.com/math-ai/internal/application/command/exam"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/utils"
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

// newContentFrom packages a fresh generation for storage. extras is the
// cache tag the set is filed under, or nil for a set that must never be
// served to anyone else.
func newContentFrom(ctx context.Context, req *dto.GenerateExamReq, generated *generateExamOutput, extras *string, level *int) (*command.NewAiExamContent, error) {
	questionsJSON, err := marshalQuestions(ctx, generated.Questions)
	if err != nil {
		return nil, err
	}
	return &command.NewAiExamContent{
		NumQues:       req.NumQuestions,
		Level:         level,
		Semester:      utils.ToStringPtr(req.Semester),
		Program:       utils.ToStringPtr(req.Program),
		Extras:        extras,
		Title:         sanitizeExamText(generated.Title),
		ShortText:     sanitizeExamText(generated.ShortText),
		QuestionsJSON: questionsJSON,
	}, nil
}

// systemRand adapts the process-wide generator to question.Shuffler. The
// top-level math/rand/v2 source is seeded by the runtime and safe for
// concurrent use, which is all a shuffle needs; nothing here is security
// sensitive.
type systemRand struct{}

func (systemRand) Shuffle(n int, swap func(i, j int)) { rand.Shuffle(n, swap) }

// drawShuffle picks this sitting's ordering and returns it in stored form.
//
// nil means "no ordering of its own", which is the identity: the sitting
// is served in stored order. That is both the empty-set case and what
// EXAM_SHUFFLE_ENABLED=false produces, so switching the feature off needs
// no special case anywhere downstream — ParseShuffle already treats a
// missing ordering as stored order, and sittings handed out while it was
// on keep the ordering they stored.
func (s *Service) drawShuffle(ctx context.Context, canonical []question.Question) (*string, error) {
	if !s.shuffleEnabled || len(canonical) == 0 {
		return nil, nil
	}
	raw, err := question.NewShuffle(canonical, systemRand{}).JSON()
	if err != nil {
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil, err)
	}
	return &raw, nil
}

// decodeStoredQuestions reads a cached set back into the shared shape so
// a shuffle can be drawn against it.
func decodeStoredQuestions(ctx context.Context, raw string) ([]question.Question, error) {
	var stored []dto.ExamQuestion
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil,
			fmt.Errorf("exam: decode cached questions: %w", err))
	}
	return dto.ToSharedQuestions(stored), nil
}
