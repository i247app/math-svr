package prompt

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/prompt"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// MaxPromptRunes caps a stored prompt. The column is MEDIUMTEXT, so the
// limit is about cost, not storage: every call of that kind sends the
// whole text.
const MaxPromptRunes = 50000

// ValidateUpdatePrompt checks what needs no stored row.
func ValidateUpdatePrompt(ctx context.Context, req *dto.UpdatePromptReq) error {
	if req.PromptID <= 0 {
		return errs.NewError(ctx, domain.PromptMissingID, nil, ErrPromptIDRequired)
	}
	if strings.TrimSpace(req.SystemPrompt) == "" {
		return errs.NewError(ctx, domain.PromptEmpty, nil, ErrPromptEmpty)
	}
	if utf8.RuneCountInString(req.SystemPrompt) > MaxPromptRunes {
		return errs.NewError(ctx, domain.PromptTooLong, nil, ErrPromptTooLong)
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
		return errs.NewError(ctx, domain.PromptMissingJSONKey, nil,
			fmt.Errorf("system_prompt does not name the JSON key(s) %s", strings.Join(missing, ", ")))
	}
	return nil
}
