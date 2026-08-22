package chat

import (
	"context"
	"errors"
	"unicode/utf8"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// previewRuneLimit keeps the denormalised preview inside the column's 255
// bytes with room to spare. Counted in runes, not bytes: Vietnamese text is
// multi-byte, and cutting on a byte boundary would split a character and
// produce invalid UTF-8 in the inbox.
const previewRuneLimit = 60

type SendMessageCommand struct {
	ConversationID   int64
	SenderProfileID  int64
	SenderUserID     int64
	MessageType      string
	Content          string
	ClientMsgID      *string
	ReplyToMessageID *int64
}

// SendMessageResult reports whether the row was newly written. Duplicate is
// true when an idempotent retry resolved to an existing message, which the
// caller uses to skip re-publishing the realtime event.
type SendMessageResult struct {
	Message   *domain.Message
	Duplicate bool
}

type SendMessageCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSendMessageCommandHandler(uow transaction.UnitOfWork) *SendMessageCommandHandler {
	return &SendMessageCommandHandler{uow: uow}
}

// Handle writes one message and everything derived from it, in a single
// transaction: allocate the sequence number, insert the row, refresh the
// conversation's preview, and bump every other participant's unread badge.
//
// They belong together because each of the last three is a lie without the
// first: a preview for a message that rolled back, or an unread badge for a
// message nobody can open.
func (h *SendMessageCommandHandler) Handle(ctx context.Context, cmd *SendMessageCommand) (*SendMessageResult, error) {
	result := &SendMessageResult{}

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// Idempotency first: a client on a flaky network resends the same
		// composed message, and it must not land twice.
		if cmd.ClientMsgID != nil && *cmd.ClientMsgID != "" {
			existing, err := repos.ChatMessage.FindByClientMsgId(ctx, cmd.ConversationID, cmd.SenderProfileID, *cmd.ClientMsgID)
			if err != nil {
				return errs.NewError(ctx, status.CHAT_MESSAGE_CREATE_FAILED, nil, err)
			}
			if existing != nil {
				result.Message = existing
				result.Duplicate = true
				return nil
			}
		}

		// The sender must still be an active participant. Checked inside the
		// transaction so a concurrent "leave" cannot slip a message in behind
		// it.
		participant, err := repos.ChatParticipant.FindByConversationAndProfile(ctx, cmd.ConversationID, cmd.SenderProfileID)
		if err != nil {
			return errs.NewError(ctx, status.CHAT_NOT_PARTICIPANT, nil, err)
		}
		if participant == nil || participant.ParticipantStatus() == nil ||
			*participant.ParticipantStatus() != string(enum.ChatParticipantStatusActive) {
			return errs.NewError(ctx, status.CHAT_NOT_PARTICIPANT, nil, errNotParticipant)
		}

		seqNo, err := repos.ChatConversation.NextSeqNo(ctx, cmd.ConversationID)
		if err != nil {
			return errs.NewError(ctx, status.CHAT_SEQ_ALLOCATION_FAILED, nil, err)
		}

		messageID, err := seqgen.Next(ctx, repos.Seq, seq.NameChatMessage)
		if err != nil {
			return err
		}

		messageType := cmd.MessageType
		if messageType == "" {
			messageType = string(enum.ChatMessageTypeText)
		}
		content := cmd.Content
		sentStatus := string(enum.ChatMessageStatusSent)

		m := domain.NewMessage()
		m.SetMessageId(messageID)
		m.SetConversationId(cmd.ConversationID)
		m.SetSeqNo(seqNo)
		m.SetSenderProfileId(&cmd.SenderProfileID)
		m.SetSenderUserId(&cmd.SenderUserID)
		m.SetMessageType(messageType)
		m.SetContent(&content)
		m.SetReplyToMessageId(cmd.ReplyToMessageID)
		m.SetClientMsgId(cmd.ClientMsgID)
		m.SetSentDt(mtime.Now())
		m.SetMessageStatus(&sentStatus)
		m.SetStatus(string(enum.StatusActive))
		m.SetCreateId(&cmd.SenderUserID)

		created, err := repos.ChatMessage.Create(ctx, m)
		if err != nil {
			// Two sends of the same client_msg_id can interleave past the
			// lookup above; the UNIQUE index is the real guard.
			if errors.Is(err, domain.ErrDuplicateMessage) {
				return err
			}
			return errs.NewError(ctx, status.CHAT_MESSAGE_CREATE_FAILED, nil, err)
		}

		if err := repos.ChatConversation.UpdateLastMessage(ctx, cmd.ConversationID, created, buildPreview(content)); err != nil {
			return errs.NewError(ctx, status.CHAT_MESSAGE_CREATE_FAILED, nil, err)
		}

		if err := repos.ChatParticipant.IncUnreadExcept(ctx, cmd.ConversationID, cmd.SenderProfileID, seqNo); err != nil {
			return errs.NewError(ctx, status.CHAT_MESSAGE_CREATE_FAILED, nil, err)
		}

		result.Message = created
		return nil
	})

	if err != nil {
		// The interleaved-retry case: the row the first attempt wrote is
		// committed now, so read it back and report success.
		if errors.Is(err, domain.ErrDuplicateMessage) && cmd.ClientMsgID != nil {
			return h.findExisting(ctx, cmd)
		}
		return nil, err
	}
	return result, nil
}

func (h *SendMessageCommandHandler) findExisting(ctx context.Context, cmd *SendMessageCommand) (*SendMessageResult, error) {
	var found *domain.Message
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		m, err := repos.ChatMessage.FindByClientMsgId(ctx, cmd.ConversationID, cmd.SenderProfileID, *cmd.ClientMsgID)
		if err != nil {
			return errs.NewError(ctx, status.CHAT_MESSAGE_CREATE_FAILED, nil, err)
		}
		found = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.CHAT_MESSAGE_CREATE_FAILED, nil, errDuplicateUnresolved)
	}
	return &SendMessageResult{Message: found, Duplicate: true}, nil
}

// buildPreview truncates on a rune boundary and appends an ellipsis so the
// inbox shows "Chào em, hôm nay…" rather than a mangled final character.
func buildPreview(content string) string {
	if utf8.RuneCountInString(content) <= previewRuneLimit {
		return content
	}
	runes := []rune(content)
	return string(runes[:previewRuneLimit]) + "…"
}

var (
	errNotParticipant      = errors.New("sender is not an active participant of the conversation")
	errDuplicateUnresolved = errors.New("duplicate message reported but no existing row found")
)
