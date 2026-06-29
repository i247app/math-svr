// Package conversation (application) hosts the bridge between langchaingo's
// memory framework and our persistence: a schema.ChatMessageHistory backed
// by MySQL. langchaingo ships no MySQL store, so implementing this interface
// is the framework's official extension point — the memory buffers
// (ConversationWindowBuffer) then layer windowing/load/save on top.
package conversation

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"

	command "math-ai.com/math-ai/internal/application/command/conversation"
	conversationDomain "math-ai.com/math-ai/internal/domain/conversation"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ChatMessageHistory implements schema.ChatMessageHistory over a single
// existing conversation thread (conversationID), persisting through the
// existing append commands and reading through the message repo.
//
// Design notes:
//   - SetMessages is a deliberate NO-OP: ConversationWindowBuffer calls it to
//     prune the store down to the window, which would permanently delete
//     history we want to keep. The window is applied at READ time
//     (LoadMemoryVariables → cutMessages), so the prompt stays bounded while
//     the database retains the full thread.
//   - Clear is a NO-OP for the same safety reason — memory pruning must never
//     destroy persisted history. Use the soft-delete command for real removal.
//   - Reads are capped at readCap rows so a long thread doesn't load entirely;
//     callers set it to at least 2×window.
type ChatMessageHistory struct {
	appendUser      *command.AppendUserMessageCommandHandler
	appendAssistant *command.AppendAssistantMessageCommandHandler
	msgRepo         conversationDomain.IMessageRepository
	conversationID  int64
	userID          int64
	readCap         int64
}

// Statically assert interface conformance.
var _ schema.ChatMessageHistory = (*ChatMessageHistory)(nil)

func NewChatMessageHistory(
	appendUser *command.AppendUserMessageCommandHandler,
	appendAssistant *command.AppendAssistantMessageCommandHandler,
	msgRepo conversationDomain.IMessageRepository,
	conversationID int64,
	userID int64,
	readCap int64,
) *ChatMessageHistory {
	return &ChatMessageHistory{
		appendUser:      appendUser,
		appendAssistant: appendAssistant,
		msgRepo:         msgRepo,
		conversationID:  conversationID,
		userID:          userID,
		readCap:         readCap,
	}
}

func (h *ChatMessageHistory) AddUserMessage(ctx context.Context, text string) error {
	cid := h.conversationID
	_, err := h.appendUser.Handle(ctx, command.AppendUserMessageCommand{
		ConversationID: &cid,
		UserID:         h.userID,
		Content:        text,
	})
	return err
}

func (h *ChatMessageHistory) AddAIMessage(ctx context.Context, text string) error {
	_, err := h.appendAssistant.Handle(ctx, command.AppendAssistantMessageCommand{
		ConversationID: h.conversationID,
		Content:        text,
	})
	return err
}

// AddMessage routes by message type. ConversationBuffer.SaveContext uses the
// typed helpers above, so this is only hit for direct AddMessage calls;
// non-human types fall back to the AI path, anything else to the user path.
func (h *ChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	switch message.GetType() {
	case llms.ChatMessageTypeAI:
		return h.AddAIMessage(ctx, message.GetContent())
	default:
		return h.AddUserMessage(ctx, message.GetContent())
	}
}

func (h *ChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	rows, err := h.msgRepo.ListRecentByConversationId(ctx, h.conversationID, h.readCap)
	if err != nil {
		return nil, err
	}
	out := make([]llms.ChatMessage, 0, len(rows))
	for _, m := range rows {
		out = append(out, toLLMMessage(m.Role(), m.Content()))
	}
	return out, nil
}

// SetMessages is intentionally a no-op — see the type doc.
func (h *ChatMessageHistory) SetMessages(_ context.Context, _ []llms.ChatMessage) error {
	return nil
}

// Clear is intentionally a no-op — see the type doc.
func (h *ChatMessageHistory) Clear(_ context.Context) error {
	return nil
}

func toLLMMessage(role, content string) llms.ChatMessage {
	switch enum.ConversationRoleType(role) {
	case enum.ConversationRoleTypeAssistant:
		return llms.AIChatMessage{Content: content}
	case enum.ConversationRoleTypeSystem:
		return llms.SystemChatMessage{Content: content}
	default:
		return llms.HumanChatMessage{Content: content}
	}
}
