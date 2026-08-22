// Package chat is the presentation/orchestration layer for messaging:
// permission gates, validation, hydration of display data, and best-effort
// realtime fan-out.
package chat

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/chat"
	dto "math-ai.com/math-ai/internal/application/dto/chat"
	query "math-ai.com/math-ai/internal/application/query/chat"
	"math-ai.com/math-ai/internal/application/transaction"
	chatDomain "math-ai.com/math-ai/internal/domain/chat"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	presenceDomain "math-ai.com/math-ai/internal/domain/presence"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// presignTTL bounds avatar URL validity. Mirrors the classroom and home
// modules so a client can cache one screen's URLs for the same period
// everywhere.
const presignTTL = 1 * time.Hour

type Service struct {
	openConversationCmd *command.OpenConversationCommandHandler
	sendMessageCmd      *command.SendMessageCommandHandler
	markReadCmd         *command.MarkReadCommandHandler

	listConversationsQuery    *query.ListConversationsQueryHandler
	listMessagesQuery         *query.ListMessagesQueryHandler
	listClassroomMembersQuery *query.ListClassroomMembersQueryHandler
	unreadCountQuery          *query.UnreadCountQueryHandler

	conversationRepo    chatDomain.IRepository
	participantRepo     chatDomain.IParticipantRepository
	profileRepo         profileDomain.IRepository
	classroomMemberRepo classroomDomain.IMemberRepository
	presenceRepo        presenceDomain.IRepository
	storageProvider     *storage.Adapter
	realtime            MessagePublisher
}

// NewService wires the module. storageProvider and realtime may be nil —
// storage disabled means no presigned avatar URLs, realtime disabled means
// clients fall back to polling the list endpoints.
func NewService(
	uow transaction.UnitOfWork,
	conversationRepo chatDomain.IRepository,
	participantRepo chatDomain.IParticipantRepository,
	messageRepo chatDomain.IMessageRepository,
	profileRepo profileDomain.IRepository,
	classroomMemberRepo classroomDomain.IMemberRepository,
	presenceRepo presenceDomain.IRepository,
	storageProvider *storage.Adapter,
	realtime MessagePublisher,
) *Service {
	return &Service{
		openConversationCmd: command.NewOpenConversationCommandHandler(uow),
		sendMessageCmd:      command.NewSendMessageCommandHandler(uow),
		markReadCmd:         command.NewMarkReadCommandHandler(uow),

		listConversationsQuery: query.NewListConversationsQueryHandler(conversationRepo, participantRepo),
		listMessagesQuery:      query.NewListMessagesQueryHandler(messageRepo, participantRepo),
		listClassroomMembersQuery: query.NewListClassroomMembersQueryHandler(
			classroomMemberRepo, profileRepo, presenceRepo, conversationRepo, participantRepo),
		unreadCountQuery: query.NewUnreadCountQueryHandler(participantRepo),

		conversationRepo:    conversationRepo,
		participantRepo:     participantRepo,
		profileRepo:         profileRepo,
		classroomMemberRepo: classroomMemberRepo,
		presenceRepo:        presenceRepo,
		storageProvider:     storageProvider,
		realtime:            realtime,
	}
}

// ListClassroomMembers backs the classroom message tab: every member, their
// online dot, and the existing thread with each so a tap opens into history.
func (s *Service) ListClassroomMembers(ctx context.Context, req *dto.ListClassroomMembersReq, sessionUserID int64) (*dto.ListClassroomMembersRes, error) {
	if err := ValidateListClassroomMembers(ctx, req); err != nil {
		return nil, err
	}
	if _, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID); err != nil {
		return nil, err
	}
	// The caller must be in the classroom to see its roster.
	if _, err := s.requireActiveMember(ctx, req.ClassroomID, req.ProfileID, ErrNotClassroomMember); err != nil {
		return nil, err
	}

	result, err := s.listClassroomMembersQuery.Handle(ctx, &query.ListClassroomMembersQuery{
		ClassroomID:    req.ClassroomID,
		ActorProfileID: req.ProfileID,
		Page:           req.Page,
		Limit:          req.Limit,
	})
	if err != nil {
		return nil, err
	}

	members := make([]*dto.ClassroomChatMember, 0, len(result.Rows))
	for _, row := range result.Rows {
		item := &dto.ClassroomChatMember{
			ProfileID:  row.Member.ProfileId(),
			MemberRole: row.Member.MemberRole(),
		}
		if row.Profile != nil {
			item.Name = row.Profile.Name()
			item.AvatarKey = row.Profile.AvatarKey()
			item.AvatarURL = s.presign(ctx, row.Profile.AvatarKey())
		}
		if row.Presence != nil {
			item.IsOnline = row.Presence.IsOnline()
			item.LastSeenDt = row.Presence.LastSeenDt()
		}
		if row.Conversation != nil {
			id := row.Conversation.ConversationId()
			item.ConversationID = &id
			item.LastMessagePreview = row.Conversation.LastMessagePreview()
			item.LastMessageDt = row.Conversation.LastMessageDt()
		}
		if row.Participant != nil {
			item.UnreadCount = row.Participant.UnreadCount()
		}
		members = append(members, item)
	}

	return &dto.ListClassroomMembersRes{Members: members, Pagination: result.Pagination}, nil
}

// OpenConversation resolves (creating on first contact) the 1-1 thread with a
// classmate. Idempotent: calling it repeatedly returns the same thread.
func (s *Service) OpenConversation(ctx context.Context, req *dto.OpenConversationReq, sessionUserID int64) (*dto.OpenConversationRes, error) {
	if err := ValidateOpenConversation(ctx, req); err != nil {
		return nil, err
	}
	actor, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	target, err := s.requireCanMessage(ctx, req.ClassroomID, req.ProfileID, req.TargetProfileID)
	if err != nil {
		return nil, err
	}

	conversation, err := s.openConversationCmd.Handle(ctx, &command.OpenConversationCommand{
		ActorProfileID:  actor.ProfileId(),
		ActorUserID:     actor.UserId(),
		TargetProfileID: target.ProfileId(),
		TargetUserID:    target.UserId(),
	})
	if err != nil {
		return nil, err
	}

	res := dto.DomainToConversationResponse(conversation)
	s.attachCounterpart(ctx, res, target)
	s.attachOwnState(ctx, res, req.ProfileID)
	return &dto.OpenConversationRes{Conversation: res}, nil
}

// ListConversations is the inbox screen.
func (s *Service) ListConversations(ctx context.Context, req *dto.ListConversationsReq, sessionUserID int64) (*dto.ListConversationsRes, error) {
	if err := ValidateProfileOnly(ctx, req.ProfileID); err != nil {
		return nil, err
	}
	if _, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID); err != nil {
		return nil, err
	}

	result, err := s.listConversationsQuery.Handle(ctx, &chatDomain.ListConversationsParams{
		ProfileId:  req.ProfileID,
		UnreadOnly: req.UnreadOnly,
		Page:       req.Page,
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*dto.ConversationResponse, 0, len(result.Conversations))
	for _, c := range result.Conversations {
		item := dto.DomainToConversationResponse(c)
		if p, ok := result.Participants[c.ConversationId()]; ok {
			item.UnreadCount = p.UnreadCount()
			item.LastReadSeqNo = p.LastReadSeqNo()
		}
		s.hydrateCounterpart(ctx, item, req.ProfileID)
		out = append(out, item)
	}

	return &dto.ListConversationsRes{Conversations: out, Pagination: result.Pagination}, nil
}

// ListMessages pages a thread's history. Authorization happens inside the
// query, on the participant row.
func (s *Service) ListMessages(ctx context.Context, req *dto.ListMessagesReq, sessionUserID int64) (*dto.ListMessagesRes, error) {
	if err := ValidateListMessages(ctx, req); err != nil {
		return nil, err
	}
	if _, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID); err != nil {
		return nil, err
	}

	messages, err := s.listMessagesQuery.Handle(ctx, &query.ListMessagesQuery{
		ConversationID: req.ConversationID,
		ProfileID:      req.ProfileID,
		BeforeSeqNo:    req.BeforeSeqNo,
		AfterSeqNo:     req.AfterSeqNo,
		Limit:          req.Limit,
	})
	if err != nil {
		return nil, err
	}
	return &dto.ListMessagesRes{Messages: dto.DomainListToMessageResponse(messages)}, nil
}

// SendMessage persists the message, then pushes it to the other participants.
//
// Persistence is authoritative and realtime is best-effort: the publish
// happens after the transaction commits and its failure is logged, never
// returned. A recipient who was offline picks the message up from
// /chats/messages/list on next open.
func (s *Service) SendMessage(ctx context.Context, req *dto.SendMessageReq, sessionUserID int64) (*dto.SendMessageRes, error) {
	if err := ValidateSendMessage(ctx, req); err != nil {
		return nil, err
	}
	actor, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	result, err := s.sendMessageCmd.Handle(ctx, &command.SendMessageCommand{
		ConversationID:   req.ConversationID,
		SenderProfileID:  actor.ProfileId(),
		SenderUserID:     actor.UserId(),
		Content:          req.Content,
		ClientMsgID:      req.ClientMsgID,
		ReplyToMessageID: req.ReplyToMessageID,
	})
	if err != nil {
		return nil, err
	}

	res := dto.DomainToMessageResponse(result.Message)
	// A retry resolved to an existing message: it was already delivered, and
	// re-publishing would make it appear twice in an open thread.
	if !result.Duplicate {
		s.publishMessage(ctx, result.Message, res)
	}
	return &dto.SendMessageRes{Message: res}, nil
}

// MarkRead advances the caller's read watermark.
func (s *Service) MarkRead(ctx context.Context, req *dto.MarkReadReq, sessionUserID int64) (*dto.MarkReadRes, error) {
	if err := ValidateMarkRead(ctx, req); err != nil {
		return nil, err
	}
	if _, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID); err != nil {
		return nil, err
	}

	effective, err := s.markReadCmd.Handle(ctx, &command.MarkReadCommand{
		ConversationID: req.ConversationID,
		ProfileID:      req.ProfileID,
		SeqNo:          req.SeqNo,
	})
	if err != nil {
		return nil, err
	}
	return &dto.MarkReadRes{ConversationID: req.ConversationID, LastReadSeqNo: effective}, nil
}

// UnreadCount is the badge on the message tab.
func (s *Service) UnreadCount(ctx context.Context, req *dto.UnreadCountReq, sessionUserID int64) (*dto.UnreadCountRes, error) {
	if err := ValidateProfileOnly(ctx, req.ProfileID); err != nil {
		return nil, err
	}
	if _, err := s.resolveActingProfile(ctx, req.ProfileID, sessionUserID); err != nil {
		return nil, err
	}

	total, err := s.unreadCountQuery.Handle(ctx, req.ProfileID)
	if err != nil {
		return nil, err
	}
	return &dto.UnreadCountRes{UnreadCount: total}, nil
}

// ---------- hydration helpers ----------

// presign turns an S3 key into a short-lived URL. Returns nil when storage is
// disabled or there is no key, so a missing avatar degrades to "no picture"
// rather than an error.
func (s *Service) presign(ctx context.Context, key *string) *string {
	if s.storageProvider == nil || key == nil || *key == "" {
		return nil
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *key,
		Expiration: presignTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("chat.presign_failed key=%s err=%v", *key, err)
		return nil
	}
	return &url
}

func (s *Service) attachCounterpart(ctx context.Context, res *dto.ConversationResponse, peer *profileDomain.Profile) {
	if res == nil || peer == nil {
		return
	}
	item := &dto.ChatPeer{
		ProfileID: peer.ProfileId(),
		Name:      peer.Name(),
		AvatarKey: peer.AvatarKey(),
		AvatarURL: s.presign(ctx, peer.AvatarKey()),
	}
	if pr, err := s.presenceRepo.FindByUserId(ctx, peer.UserId()); err == nil && pr != nil {
		item.IsOnline = pr.IsOnline()
		item.LastSeenDt = pr.LastSeenDt()
	}
	res.Counterpart = item
}

// hydrateCounterpart fills in the other side of a DIRECT thread for the inbox.
//
// This is deliberately per-row rather than batched: the inbox page is small
// (20 by default) and threads are the caller's own, so the extra reads are
// bounded. If the page size ever grows, batch it the way the classroom member
// list already does.
func (s *Service) hydrateCounterpart(ctx context.Context, res *dto.ConversationResponse, actorProfileID int64) {
	if res == nil {
		return
	}
	participants, err := s.participantRepo.ListByConversationId(ctx, &chatDomain.ListParticipantsParams{
		ConversationId: res.ConversationID,
	})
	if err != nil {
		logger.From(ctx).Warnf("chat.counterpart_failed conversation_id=%d err=%v", res.ConversationID, err)
		return
	}
	for _, p := range participants {
		if p.ProfileId() == actorProfileID {
			continue
		}
		peer, err := s.profileRepo.FindByProfileId(ctx, p.ProfileId())
		if err != nil || peer == nil {
			continue
		}
		s.attachCounterpart(ctx, res, peer)
		return
	}
}

func (s *Service) attachOwnState(ctx context.Context, res *dto.ConversationResponse, actorProfileID int64) {
	if res == nil {
		return
	}
	p, err := s.participantRepo.FindByConversationAndProfile(ctx, res.ConversationID, actorProfileID)
	if err != nil || p == nil {
		return
	}
	res.UnreadCount = p.UnreadCount()
	res.LastReadSeqNo = p.LastReadSeqNo()
}
