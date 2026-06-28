package conversation

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/conversation"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ValidateSend checks the per-turn request. maxChars is the (already
// clamped) limit from ConversationConfig.
func ValidateSend(ctx context.Context, req *dto.SendMessageReq, maxChars int) error {
	if strings.TrimSpace(req.Message) == "" {
		return errs.NewError(ctx, status.AI_CONVERSATION_MISSING_MESSAGE, nil, ErrMessageRequired)
	}
	if len([]rune(req.Message)) > maxChars {
		return errs.NewError(ctx, status.AI_CONVERSATION_MESSAGE_TOO_LONG,
			map[string]any{"max": maxChars}, ErrMessageTooLong)
	}
	if req.ConversationID != nil && *req.ConversationID <= 0 {
		return errs.NewError(ctx, status.AI_CONVERSATION_NOT_FOUND, nil, ErrConversationIDInvalid)
	}
	return nil
}
