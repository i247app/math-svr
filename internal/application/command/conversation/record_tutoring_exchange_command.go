package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// RecordTutoringExchangeCommand appends one compact learning exchange to a
// profile's long-lived tutoring thread (find-or-create by profile+purpose),
// all inside one UoW so the two turns get contiguous seq_no values.
//
// UserSummary is a short performance note (NOT the raw quiz JSON);
// AssistantReview is the AI's qualitative feedback. Empty parts are
// skipped. This is what gives later quiz generation cross-quiz memory
// without bloating tokens.
type RecordTutoringExchangeCommand struct {
	UserID          int64
	ProfileID       int64
	Purpose         string
	UserSummary     string
	AssistantReview string
}

type RecordTutoringExchangeCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRecordTutoringExchangeCommandHandler(uow transaction.UnitOfWork) *RecordTutoringExchangeCommandHandler {
	return &RecordTutoringExchangeCommandHandler{uow: uow}
}

func (h *RecordTutoringExchangeCommandHandler) Handle(ctx context.Context, cmd RecordTutoringExchangeCommand) (*conversation.Conversation, error) {
	var result *conversation.Conversation

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		conv, err := h.findOrCreate(ctx, repos, cmd)
		if err != nil {
			return err
		}

		// Contiguous seq_no across both turns; each appendMessage advances
		// message_count by one in the same tx.
		seqNo := conv.MessageCount()
		if s := strings.TrimSpace(cmd.UserSummary); s != "" {
			if _, err := appendMessage(ctx, repos, conv.ConversationId(), seqNo,
				enum.ConversationRoleTypeUser, s, &cmd.UserID); err != nil {
				return err
			}
			seqNo++
		}
		if s := strings.TrimSpace(cmd.AssistantReview); s != "" {
			if _, err := appendMessage(ctx, repos, conv.ConversationId(), seqNo,
				enum.ConversationRoleTypeAssistant, s, nil); err != nil {
				return err
			}
		}

		result = conv
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return result, nil
}

// findOrCreate resolves the profile's tutoring thread, creating a new one
// (tagged with cmd.Purpose) when none exists yet.
func (h *RecordTutoringExchangeCommandHandler) findOrCreate(ctx context.Context, repos transaction.Repositories, cmd RecordTutoringExchangeCommand) (*conversation.Conversation, error) {
	conv, err := repos.Conversation.FindLatestActiveByProfileAndPurpose(ctx, cmd.ProfileID, cmd.Purpose)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if conv != nil {
		return conv, nil
	}

	conversationID, err := nextSeqID(ctx, repos, seq.NameAiConversation)
	if err != nil {
		return nil, err
	}

	c := conversation.NewConversation()
	c.SetConversationId(conversationID)
	c.SetUserId(cmd.UserID)
	profileID := cmd.ProfileID
	c.SetProfileId(&profileID)
	purpose := cmd.Purpose
	c.SetPurpose(&purpose)
	c.SetMessageCount(0)
	active := string(enum.ConversationStatusTypeActive)
	c.SetConversationStatus(&active)
	c.SetCreateId(&cmd.UserID)

	created, err := repos.Conversation.Create(ctx, c)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return created, nil
}
