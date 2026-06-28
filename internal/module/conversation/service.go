package conversation

import (
	"context"
	"strings"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/conversation"
	dto "math-ai.com/math-ai/internal/application/dto/conversation"
	query "math-ai.com/math-ai/internal/application/query/conversation"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/conversation"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
)

// conversationGetMessageLimit caps how many messages GetConversationById
// returns. Larger than the chat window since reading history is a cheap
// local DB read with no token cost.
const conversationGetMessageLimit int64 = 100

// chat sampling for conversational replies — a bit more creative than quiz
// generation, but JSON mode is off (free-form prose).
const (
	chatTemperature = 0.7
	chatTopP        = 0.95
)

// windowConfig is the sanitized (clamped) view of ConversationConfig the
// service actually uses on the hot path.
type windowConfig struct {
	enabled  bool
	size     int64
	maxChars int
}

func sanitizeConfig(c config.ConversationConfig) windowConfig {
	size := c.HistoryWindowSize
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	maxChars := c.MaxMessageChars
	if maxChars <= 0 {
		maxChars = 4000
	}
	if maxChars > 60000 {
		maxChars = 60000
	}
	return windowConfig{enabled: c.HistoryWindowEnabled, size: int64(size), maxChars: maxChars}
}

// Service is the conversation module façade. It orchestrates two UoW writes
// around one out-of-tx bot call (the LLM must not hold a transaction open).
type Service struct {
	appendUserCmd      *command.AppendUserMessageCommandHandler
	appendAssistantCmd *command.AppendAssistantMessageCommandHandler
	softDeleteCmd      *command.SoftDeleteConversationCommandHandler
	getQuery           *query.GetConversationByIdQueryHandler
	listQuery          *query.ListConversationsQueryHandler
	bot                *botAdapter.Adapter
	cfg                windowConfig
}

// NewService wires the conversation module. bot may be nil when the deploy
// runs with BOT_PROVIDER=""/"disabled"; SendMessage then returns
// AI_CONVERSATION_DISABLED so the caller sees a uniform error shape.
func NewService(
	uow transaction.UnitOfWork,
	convRepo domain.IRepository,
	msgRepo domain.IMessageRepository,
	bot *botAdapter.Adapter,
	cfg config.ConversationConfig,
) *Service {
	return &Service{
		appendUserCmd:      command.NewAppendUserMessageCommandHandler(uow),
		appendAssistantCmd: command.NewAppendAssistantMessageCommandHandler(uow),
		softDeleteCmd:      command.NewSoftDeleteConversationCommandHandler(uow),
		getQuery:           query.NewGetConversationByIdQueryHandler(convRepo, msgRepo),
		listQuery:          query.NewListConversationsQueryHandler(convRepo),
		bot:                bot,
		cfg:                sanitizeConfig(cfg),
	}
}

// SendMessage runs one chat turn: persist the user message (UoW #1), call
// the LLM with the windowed history (outside any tx), persist the assistant
// reply (UoW #2). On LLM failure the user message stays persisted so the
// client can retry.
func (s *Service) SendMessage(ctx context.Context, req *dto.SendMessageReq) (*dto.SendMessageRes, error) {
	log := logger.From(ctx)

	if s.bot == nil {
		return nil, errs.NewError(ctx, status.AI_CONVERSATION_DISABLED, nil, ErrBotDisabled)
	}
	if req.UserID == nil {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	if err := ValidateSend(ctx, req, s.cfg.maxChars); err != nil {
		return nil, err
	}

	// UoW #1 — find/create conversation + persist the user turn.
	userRes, err := s.appendUserCmd.Handle(ctx, command.AppendUserMessageCommand{
		ConversationID: req.ConversationID,
		UserID:         *req.UserID,
		Content:        strings.TrimSpace(req.Message),
	})
	if err != nil {
		return nil, err
	}
	conversationID := userRes.Conversation.ConversationId()

	// Build the prompt: system + (windowed history | just this turn).
	messages, err := s.buildPromptMessages(ctx, conversationID, userRes.UserMessage)
	if err != nil {
		return nil, err
	}

	// LLM call — OUTSIDE any tx.
	out, err := s.bot.Chat(ctx, botAdapter.ChatRequest{
		Messages:    messages,
		Temperature: chatTemperature,
		TopP:        chatTopP,
	})
	if err != nil {
		log.Warnf("conversation.generation_failed conversation_id=%d uid=%d err=%v",
			conversationID, *req.UserID, err)
		return nil, errs.NewError(ctx, status.AI_CONVERSATION_GENERATION_FAILED, nil, err)
	}
	reply := strings.TrimSpace(out.Content)

	// UoW #2 — persist the assistant turn.
	assistant, err := s.appendAssistantCmd.Handle(ctx, command.AppendAssistantMessageCommand{
		ConversationID: conversationID,
		Content:        reply,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("conversation.turn conversation_id=%d uid=%d window=%t reply_len=%d",
		conversationID, *req.UserID, s.cfg.enabled, len(reply))

	return &dto.SendMessageRes{
		ConversationID: conversationID,
		Reply:          reply,
		Message:        dto.MessageDomainToResponse(assistant),
	}, nil
}

// buildPromptMessages assembles the system prompt plus, when the history
// window is enabled, the recent turns (already ending with the just-saved
// user message). When disabled it sends only the system prompt + this turn.
func (s *Service) buildPromptMessages(ctx context.Context, conversationID int64, userMessage *domain.Message) ([]botAdapter.Message, error) {
	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	messages := []botAdapter.Message{
		{Role: botAdapter.RoleSystem, Content: systemPrompt(lang)},
	}

	if !s.cfg.enabled {
		messages = append(messages, botAdapter.Message{
			Role:    botAdapter.RoleUser,
			Content: userMessage.Content(),
		})
		return messages, nil
	}

	res, err := s.getQuery.Handle(ctx, query.GetConversationByIdQuery{
		ConversationID: conversationID,
		MessageLimit:   s.cfg.size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if res == nil {
		// Should not happen (we just wrote a row), but degrade to this turn.
		messages = append(messages, botAdapter.Message{
			Role:    botAdapter.RoleUser,
			Content: userMessage.Content(),
		})
		return messages, nil
	}
	for _, m := range res.Messages {
		messages = append(messages, botAdapter.Message{
			Role:    botAdapter.Role(m.Role()),
			Content: m.Content(),
		})
	}
	return messages, nil
}

func (s *Service) ListConversations(ctx context.Context, req *dto.ListConversationsReq) (*dto.ListConversationsRes, error) {
	if req.UserID == nil {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	page := int64(req.Page)
	if page <= 0 {
		page = 1
	}
	size := int64(req.Size)
	if size <= 0 {
		size = 20
	}

	conversations, pg, err := s.listQuery.Handle(ctx, query.ListConversationsQuery{
		UserID: *req.UserID,
		Page:   page,
		Limit:  size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &dto.ListConversationsRes{
		Conversations: dto.ConversationListToResponse(conversations),
		Pagination:    pg,
	}, nil
}

func (s *Service) GetConversationById(ctx context.Context, req *dto.GetConversationByIdReq) (*dto.GetConversationByIdRes, error) {
	if req.UserID == nil {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}

	res, err := s.getQuery.Handle(ctx, query.GetConversationByIdQuery{
		ConversationID: req.ConversationID,
		MessageLimit:   conversationGetMessageLimit,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if res == nil {
		return nil, errs.NewError(ctx, status.AI_CONVERSATION_NOT_FOUND, nil, ErrConversationNotFound)
	}
	if res.Conversation.UserId() != *req.UserID {
		return nil, errs.NewError(ctx, status.AI_CONVERSATION_NOT_OWNED, nil, ErrConversationNotOwned)
	}

	return &dto.GetConversationByIdRes{
		Conversation: dto.ConversationDomainToResponse(res.Conversation),
		Messages:     dto.MessageListToResponse(res.Messages),
	}, nil
}

func (s *Service) SoftDeleteConversation(ctx context.Context, req *dto.DeleteConversationReq) (*dto.DeleteConversationRes, error) {
	if req.UserID == nil {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	if err := s.softDeleteCmd.Handle(ctx, command.SoftDeleteConversationCommand{
		ConversationID: req.ConversationID,
		UserID:         *req.UserID,
	}); err != nil {
		return nil, err
	}
	return &dto.DeleteConversationRes{}, nil
}
