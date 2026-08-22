package chat

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	dto "math-ai.com/math-ai/internal/application/dto/chat"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// maxMessageRunes bounds one message. Counted in runes because the column is
// TEXT and Vietnamese is multi-byte — a byte limit would cut a typical
// Vietnamese message roughly a third shorter than the same English one.
const maxMessageRunes = 4000

func ValidateListClassroomMembers(ctx context.Context, req *dto.ListClassroomMembersReq) error {
	if req == nil {
		return errs.NewError(ctx, status.CHAT_MISSING_CLASSROOM_ID, nil, errors.New("request body is required"))
	}
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_PROFILE_ID, nil, errors.New("profile_id is required"))
	}
	if req.ClassroomID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_CLASSROOM_ID, nil, errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateOpenConversation(ctx context.Context, req *dto.OpenConversationReq) error {
	if req == nil {
		return errs.NewError(ctx, status.CHAT_MISSING_TARGET_PROFILE_ID, nil, errors.New("request body is required"))
	}
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_PROFILE_ID, nil, errors.New("profile_id is required"))
	}
	if req.ClassroomID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_CLASSROOM_ID, nil, errors.New("classroom_id is required"))
	}
	if req.TargetProfileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_TARGET_PROFILE_ID, nil, errors.New("target_profile_id is required"))
	}
	if req.ProfileID == req.TargetProfileID {
		return errs.NewError(ctx, status.CHAT_CANNOT_MESSAGE_SELF, nil, ErrCannotMessageSelf)
	}
	return nil
}

func ValidateSendMessage(ctx context.Context, req *dto.SendMessageReq) error {
	if req == nil {
		return errs.NewError(ctx, status.CHAT_MESSAGE_EMPTY, nil, errors.New("request body is required"))
	}
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_PROFILE_ID, nil, errors.New("profile_id is required"))
	}
	if req.ConversationID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_CONVERSATION_ID, nil, errors.New("conversation_id is required"))
	}
	// Whitespace-only is empty as far as a reader is concerned; the stored
	// content keeps the user's original spacing.
	if strings.TrimSpace(req.Content) == "" {
		return errs.NewError(ctx, status.CHAT_MESSAGE_EMPTY, nil, errors.New("content is required"))
	}
	if utf8.RuneCountInString(req.Content) > maxMessageRunes {
		return errs.NewError(ctx, status.CHAT_MESSAGE_TOO_LONG, nil, errors.New("content exceeds the maximum length"))
	}
	return nil
}

func ValidateListMessages(ctx context.Context, req *dto.ListMessagesReq) error {
	if req == nil {
		return errs.NewError(ctx, status.CHAT_MISSING_CONVERSATION_ID, nil, errors.New("request body is required"))
	}
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_PROFILE_ID, nil, errors.New("profile_id is required"))
	}
	if req.ConversationID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_CONVERSATION_ID, nil, errors.New("conversation_id is required"))
	}
	return nil
}

func ValidateMarkRead(ctx context.Context, req *dto.MarkReadReq) error {
	if req == nil {
		return errs.NewError(ctx, status.CHAT_MISSING_CONVERSATION_ID, nil, errors.New("request body is required"))
	}
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_PROFILE_ID, nil, errors.New("profile_id is required"))
	}
	if req.ConversationID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_CONVERSATION_ID, nil, errors.New("conversation_id is required"))
	}
	if req.SeqNo <= 0 {
		return errs.NewError(ctx, status.FAIL, nil, errors.New("seq_no must be positive"))
	}
	return nil
}

func ValidateProfileOnly(ctx context.Context, profileID int64) error {
	if profileID <= 0 {
		return errs.NewError(ctx, status.CHAT_MISSING_PROFILE_ID, nil, errors.New("profile_id is required"))
	}
	return nil
}
