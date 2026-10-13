package prompt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	dto "math-ai.com/math-ai/internal/application/dto/prompt"
	domain "math-ai.com/math-ai/internal/domain/prompt"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// MaxPromptRunes caps a stored prompt. The column is MEDIUMTEXT, so the
// limit is about cost, not storage: every call of that kind sends the
// whole text.
const MaxPromptRunes = 50000

var (
	ErrPromptIDRequired = errors.New("prompt_id is required")
	ErrPromptEmpty      = errors.New("system_prompt is required")
	ErrPromptTooLong    = errors.New("system_prompt is longer than 50000 characters")
)

// ValidateUpdatePrompt checks what needs no stored row.
func ValidateUpdatePrompt(ctx context.Context, req *dto.UpdatePromptReq) error {
	if req.PromptID <= 0 {
		return errs.NewError(ctx, status.PROMPT_MISSING_ID, nil, ErrPromptIDRequired)
	}
	if strings.TrimSpace(req.SystemPrompt) == "" {
		return errs.NewError(ctx, status.PROMPT_EMPTY, nil, ErrPromptEmpty)
	}
	if utf8.RuneCountInString(req.SystemPrompt) > MaxPromptRunes {
		return errs.NewError(ctx, status.PROMPT_TOO_LONG, nil, ErrPromptTooLong)
	}
	return nil
}

// ValidatePromptKeys refuses a text that does not name every JSON key the
// reply parser of its code reads (catalog Spec.RequiredKeys). The admin owns
// the whole text, output structure included, so a save that drops a key
// would make every later call of that kind unreadable. This proves the text
// NAMES each key; it cannot prove the model obeys.
func ValidatePromptKeys(ctx context.Context, code domain.Code, text string) error {
	spec, ok := domain.SpecOf(code)
	if !ok {
		// A row whose code no reader knows: nothing parses its reply.
		return nil
	}
	var missing []string
	for _, key := range spec.RequiredKeys {
		if !strings.Contains(text, key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return errs.NewError(ctx, status.PROMPT_MISSING_JSON_KEY, nil,
			fmt.Errorf("system_prompt does not name the JSON key(s) %s", strings.Join(missing, ", ")))
	}
	return nil
}
