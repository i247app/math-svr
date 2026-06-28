package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// AppendUserMessageCommand is step 1 of a chat turn. It finds (or creates)
// the conversation, then persists the user message with the next seq_no —
// all inside one UoW so the seq_no derived from message_count is race-safe.
// The LLM call happens AFTER this command, outside any tx; the assistant
// reply is persisted by AppendAssistantMessageCommand.
//
// ConversationID nil → start a new conversation owned by UserID. When set,
// the conversation must exist and be owned by UserID.
type AppendUserMessageCommand struct {
	ConversationID *int64
	UserID         int64
	ProfileID      *int64
	Content        string
}

// AppendUserMessageResult carries the (possibly newly created) conversation
// and the persisted user message back to the caller.
type AppendUserMessageResult struct {
	Conversation *conversation.Conversation
	UserMessage  *conversation.Message
}

type AppendUserMessageCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewAppendUserMessageCommandHandler(uow transaction.UnitOfWork) *AppendUserMessageCommandHandler {
	return &AppendUserMessageCommandHandler{uow: uow}
}

func (h *AppendUserMessageCommandHandler) Handle(ctx context.Context, cmd AppendUserMessageCommand) (*AppendUserMessageResult, error) {
	var result AppendUserMessageResult

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		conv, err := h.findOrCreate(ctx, repos, cmd)
		if err != nil {
			return err
		}

		msg, err := appendMessage(ctx, repos, conv.ConversationId(),
			conv.MessageCount(), enum.ConversationRoleTypeUser, cmd.Content, &cmd.UserID)
		if err != nil {
			return err
		}

		result.Conversation = conv
		result.UserMessage = msg
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return &result, nil
}

// findOrCreate resolves the target conversation: a new owned row when
// ConversationID is nil, otherwise the existing row after a not-found /
// ownership check.
func (h *AppendUserMessageCommandHandler) findOrCreate(ctx context.Context, repos transaction.Repositories, cmd AppendUserMessageCommand) (*conversation.Conversation, error) {
	if cmd.ConversationID != nil {
		conv, err := repos.Conversation.FindByConversationId(ctx, *cmd.ConversationID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if conv == nil {
			return nil, errs.NewError(ctx, status.AI_CONVERSATION_NOT_FOUND, nil,
				errors.New("conversation not found"))
		}
		if conv.UserId() != cmd.UserID {
			return nil, errs.NewError(ctx, status.AI_CONVERSATION_NOT_OWNED, nil,
				errors.New("conversation not owned by user"))
		}
		return conv, nil
	}

	conversationID, err := nextSeqID(ctx, repos, seq.NameAiConversation)
	if err != nil {
		return nil, err
	}

	conv := conversation.NewConversation()
	conv.SetConversationId(conversationID)
	conv.SetUserId(cmd.UserID)
	conv.SetProfileId(cmd.ProfileID)
	conv.SetMessageCount(0)
	active := string(enum.ConversationStatusTypeActive)
	conv.SetConversationStatus(&active)
	conv.SetCreateId(&cmd.UserID)

	created, err := repos.Conversation.Create(ctx, conv)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return created, nil
}

// appendMessage mints a message id, persists the turn with seqNo, and
// advances the conversation's message_count by one — all on the supplied
// (tx-bound) repos. Shared by the user and assistant append commands.
func appendMessage(ctx context.Context, repos transaction.Repositories, conversationID, seqNo int64,
	role enum.ConversationRoleType, content string, createID *int64) (*conversation.Message, error) {

	messageID, err := nextSeqID(ctx, repos, seq.NameAiMessage)
	if err != nil {
		return nil, err
	}

	msg := conversation.NewMessage()
	msg.SetMessageId(messageID)
	msg.SetConversationId(conversationID)
	msg.SetRole(string(role))
	msg.SetContent(content)
	msg.SetSeqNo(seqNo)
	msg.SetCreateId(createID)

	saved, err := repos.ConversationMessage.Create(ctx, msg)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	if err := repos.Conversation.IncMessageCount(ctx, conversationID, 1); err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return saved, nil
}
