package conversation

import (
	"context"
	"strings"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/conversation"
	appconv "math-ai.com/math-ai/internal/application/conversation"
	dto "math-ai.com/math-ai/internal/application/dto/conversation"
	query "math-ai.com/math-ai/internal/application/query/conversation"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/conversation"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/libs/langchain"
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
	createConvCmd      *command.CreateConversationCommandHandler
	appendUserCmd      *command.AppendUserMessageCommandHandler
	appendAssistantCmd *command.AppendAssistantMessageCommandHandler
	softDeleteCmd      *command.SoftDeleteConversationCommandHandler
	getQuery           *query.GetConversationByIdQueryHandler
	listQuery          *query.ListConversationsQueryHandler
	convRepo           domain.IRepository
	msgRepo            domain.IMessageRepository
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
		createConvCmd:      command.NewCreateConversationCommandHandler(uow),
		appendUserCmd:      command.NewAppendUserMessageCommandHandler(uow),
		appendAssistantCmd: command.NewAppendAssistantMessageCommandHandler(uow),
		softDeleteCmd:      command.NewSoftDeleteConversationCommandHandler(uow),
		getQuery:           query.NewGetConversationByIdQueryHandler(convRepo, msgRepo),
		listQuery:          query.NewListConversationsQueryHandler(convRepo),
		convRepo:           convRepo,
		msgRepo:            msgRepo,
		bot:                bot,
		cfg:                sanitizeConfig(cfg),
	}
}

// SendMessage runs one chat turn through langchaingo's memory framework:
// resolve (find/create) the conversation, let a ConversationWindowBuffer
// (backed by the MySQL ChatMessageHistory) load the windowed prior turns,
// call the model on our client, then let the same buffer SAVE the new turn
// back to MySQL. The LLM call stays outside any tx.
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
	uid := *req.UserID
	input := strings.TrimSpace(req.Message)

	// Resolve the thread (ownership-checked) BEFORE any LLM cost.
	conversationID, err := s.resolveConversation(ctx, req.ConversationID, uid)
	if err != nil {
		return nil, err
	}

	// Framework memory: window buffer over the MySQL chat-history store.
	windowSize := int(s.cfg.size)
	history := appconv.NewChatMessageHistory(
		s.appendUserCmd, s.appendAssistantCmd, s.msgRepo, conversationID, uid, s.cfg.size*2)
	mem := langchain.NewWindowMemory(history, windowSize)

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	messages := []botAdapter.Message{
		{Role: botAdapter.RoleSystem, Content: systemPrompt(lang)},
	}
	if s.cfg.enabled {
		prior, err := langchain.LoadHistoryString(ctx, mem)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if strings.TrimSpace(prior) != "" {
			messages = append(messages, botAdapter.Message{
				Role:    botAdapter.RoleSystem,
				Content: priorContextHeader(lang) + "\n" + prior,
			})
		}
	}
	messages = append(messages, botAdapter.Message{Role: botAdapter.RoleUser, Content: input})

	// LLM call — OUTSIDE any tx.
	out, err := s.bot.Chat(ctx, botAdapter.ChatRequest{
		Messages:    messages,
		Temperature: chatTemperature,
		TopP:        chatTopP,
	})
	if err != nil {
		log.Warnf("conversation.generation_failed conversation_id=%d uid=%d err=%v",
			conversationID, uid, err)
		return nil, errs.NewError(ctx, status.AI_CONVERSATION_GENERATION_FAILED, nil, err)
	}
	reply := strings.TrimSpace(out.Content)

	// Framework memory persists BOTH turns (user input + AI reply) through the
	// MySQL history store.
	if err := langchain.SaveTurn(ctx, mem, input, reply); err != nil {
		return nil, err
	}

	log.Infof("conversation.turn conversation_id=%d uid=%d window=%t reply_len=%d",
		conversationID, uid, s.cfg.enabled, len(reply))

	// Surface the just-saved assistant turn in the response.
	var msgResp dto.MessageResponse
	if recent, rerr := s.msgRepo.ListRecentByConversationId(ctx, conversationID, 1); rerr == nil && len(recent) > 0 {
		msgResp = dto.MessageDomainToResponse(recent[len(recent)-1])
	}

	return &dto.SendMessageRes{
		ConversationID: conversationID,
		Reply:          reply,
		Message:        msgResp,
	}, nil
}

// resolveConversation creates a new chat thread when id is nil, otherwise
// loads the existing one and enforces ownership — all before the LLM call so
// a not-found / not-owned request never burns vendor quota.
func (s *Service) resolveConversation(ctx context.Context, id *int64, uid int64) (int64, error) {
	if id == nil {
		conv, err := s.createConvCmd.Handle(ctx, command.CreateConversationCommand{
			UserID:  uid,
			Purpose: domain.PurposeChat,
		})
		if err != nil {
			return 0, err
		}
		return conv.ConversationId(), nil
	}

	conv, err := s.convRepo.FindByConversationId(ctx, *id)
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if conv == nil {
		return 0, errs.NewError(ctx, status.AI_CONVERSATION_NOT_FOUND, nil, ErrConversationNotFound)
	}
	if conv.UserId() != uid {
		return 0, errs.NewError(ctx, status.AI_CONVERSATION_NOT_OWNED, nil, ErrConversationNotOwned)
	}
	return conv.ConversationId(), nil
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
