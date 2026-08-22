package chat

import (
	"context"
	"strings"
	"testing"

	dto "math-ai.com/math-ai/internal/application/dto/chat"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func wantCode(t *testing.T, err error, code status.StatusCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error with code %d, got nil", code)
	}
	mErr, ok := errs.IsMathError(err)
	if !ok {
		t.Fatalf("expected a MathError, got %T: %v", err, err)
	}
	if got := mErr.GetStatusCode(); got != code {
		t.Errorf("status code = %d, want %d", got, code)
	}
}

func TestValidateSendMessage(t *testing.T) {
	ctx := context.Background()
	longEnough := strings.Repeat("a", maxMessageRunes)
	tooLong := strings.Repeat("a", maxMessageRunes+1)

	tests := []struct {
		name     string
		req      *dto.SendMessageReq
		wantCode status.StatusCode
		wantOK   bool
	}{
		{
			name:   "valid",
			req:    &dto.SendMessageReq{ProfileID: 1, ConversationID: 2, Content: "chào em"},
			wantOK: true,
		},
		{
			name:   "content exactly at the limit is accepted",
			req:    &dto.SendMessageReq{ProfileID: 1, ConversationID: 2, Content: longEnough},
			wantOK: true,
		},
		{
			name:     "missing profile",
			req:      &dto.SendMessageReq{ConversationID: 2, Content: "hi"},
			wantCode: status.CHAT_MISSING_PROFILE_ID,
		},
		{
			name:     "missing conversation",
			req:      &dto.SendMessageReq{ProfileID: 1, Content: "hi"},
			wantCode: status.CHAT_MISSING_CONVERSATION_ID,
		},
		{
			name:     "empty content",
			req:      &dto.SendMessageReq{ProfileID: 1, ConversationID: 2, Content: ""},
			wantCode: status.CHAT_MESSAGE_EMPTY,
		},
		{
			// A message of only spaces reads as blank to the recipient, so it
			// is rejected the same way an empty one is.
			name:     "whitespace-only content",
			req:      &dto.SendMessageReq{ProfileID: 1, ConversationID: 2, Content: "   \n\t "},
			wantCode: status.CHAT_MESSAGE_EMPTY,
		},
		{
			name:     "content over the limit",
			req:      &dto.SendMessageReq{ProfileID: 1, ConversationID: 2, Content: tooLong},
			wantCode: status.CHAT_MESSAGE_TOO_LONG,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSendMessage(ctx, tc.req)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			wantCode(t, err, tc.wantCode)
		})
	}
}

// The limit is in runes, not bytes: Vietnamese characters are multi-byte, so a
// byte-based limit would cut a Vietnamese message far shorter than an English
// one of the same visible length.
func TestValidateSendMessageCountsRunesNotBytes(t *testing.T) {
	ctx := context.Background()
	// Each "ữ" is 3 bytes in UTF-8, so this is ~3x the rune limit in bytes but
	// exactly at the limit in runes.
	content := strings.Repeat("ữ", maxMessageRunes)
	if len(content) <= maxMessageRunes {
		t.Fatalf("test setup wrong: expected multi-byte content, got %d bytes", len(content))
	}
	if err := ValidateSendMessage(ctx, &dto.SendMessageReq{
		ProfileID: 1, ConversationID: 2, Content: content,
	}); err != nil {
		t.Fatalf("multi-byte content at the rune limit must be accepted, got %v", err)
	}
}

func TestValidateOpenConversation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		req      *dto.OpenConversationReq
		wantCode status.StatusCode
		wantOK   bool
	}{
		{
			name:   "valid",
			req:    &dto.OpenConversationReq{ProfileID: 1, ClassroomID: 5, TargetProfileID: 2},
			wantOK: true,
		},
		{
			name:     "missing classroom",
			req:      &dto.OpenConversationReq{ProfileID: 1, TargetProfileID: 2},
			wantCode: status.CHAT_MISSING_CLASSROOM_ID,
		},
		{
			name:     "missing target",
			req:      &dto.OpenConversationReq{ProfileID: 1, ClassroomID: 5},
			wantCode: status.CHAT_MISSING_TARGET_PROFILE_ID,
		},
		{
			name:     "messaging yourself",
			req:      &dto.OpenConversationReq{ProfileID: 1, ClassroomID: 5, TargetProfileID: 1},
			wantCode: status.CHAT_CANNOT_MESSAGE_SELF,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOpenConversation(ctx, tc.req)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			wantCode(t, err, tc.wantCode)
		})
	}
}
