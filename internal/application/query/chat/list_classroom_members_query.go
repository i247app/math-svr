package chat

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/chat"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	presenceDomain "math-ai.com/math-ai/internal/domain/presence"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListClassroomMembersQuery is the message tab: everyone in the classroom,
// their online state, and the thread the caller already has with each of them.
type ListClassroomMembersQuery struct {
	ClassroomID    int64
	ActorProfileID int64
	Page           int64
	Limit          int64
}

// ClassroomMemberRow is one assembled row. Conversation and Participant are
// nil for a member the caller has never messaged — the common case on first
// open, and the reason the thread is created lazily rather than for every pair
// in the class.
type ClassroomMemberRow struct {
	Member       *classroomDomain.Member
	Profile      *profileDomain.Profile
	Presence     *presenceDomain.Presence
	Conversation *domain.Conversation
	Participant  *domain.Participant
}

type ListClassroomMembersResult struct {
	Rows       []*ClassroomMemberRow
	Pagination *pagination.Pagination
}

type ListClassroomMembersQueryHandler struct {
	memberRepo       classroomDomain.IMemberRepository
	profileRepo      profileDomain.IRepository
	presenceRepo     presenceDomain.IRepository
	conversationRepo domain.IRepository
	participantRepo  domain.IParticipantRepository
}

func NewListClassroomMembersQueryHandler(
	memberRepo classroomDomain.IMemberRepository,
	profileRepo profileDomain.IRepository,
	presenceRepo presenceDomain.IRepository,
	conversationRepo domain.IRepository,
	participantRepo domain.IParticipantRepository,
) *ListClassroomMembersQueryHandler {
	return &ListClassroomMembersQueryHandler{
		memberRepo:       memberRepo,
		profileRepo:      profileRepo,
		presenceRepo:     presenceRepo,
		conversationRepo: conversationRepo,
		participantRepo:  participantRepo,
	}
}

// Handle assembles the screen from five batched reads — members, profiles,
// conversations, the caller's participant rows, presence — rather than looping
// per member. A 40-person class would otherwise cost over a hundred queries
// every time someone opens the tab.
func (h *ListClassroomMembersQueryHandler) Handle(ctx context.Context, q *ListClassroomMembersQuery) (*ListClassroomMembersResult, error) {
	activeStatus := string(enum.ClassroomMemberStatusTypeActive)
	members, pg, err := h.memberRepo.ListMembers(ctx, &classroomDomain.ListMembersParams{
		ClassroomId: &q.ClassroomID,
		Status:      &activeStatus,
		Page:        q.Page,
		Limit:       q.Limit,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	// Drop the caller: you do not message yourself, and showing your own row
	// with an "open chat" affordance is the kind of thing that ships.
	profileIDs := make([]int64, 0, len(members))
	kept := make([]*classroomDomain.Member, 0, len(members))
	for _, m := range members {
		if m.ProfileId() == q.ActorProfileID {
			continue
		}
		kept = append(kept, m)
		profileIDs = append(profileIDs, m.ProfileId())
	}
	if len(kept) == 0 {
		return &ListClassroomMembersResult{Rows: nil, Pagination: pg}, nil
	}

	profiles, err := h.profileRepo.ListByProfileIds(ctx, profileIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	profileByID := make(map[int64]*profileDomain.Profile, len(profiles))
	userIDs := make([]int64, 0, len(profiles))
	for _, p := range profiles {
		profileByID[p.ProfileId()] = p
		userIDs = append(userIDs, p.UserId())
	}

	// Presence is keyed by user, not profile: a WebSocket belongs to an
	// account, so every profile a person holds shares one online state.
	presenceByUser, err := h.presenceRepo.ListByUserIds(ctx, userIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	// One lookup for every pair at once, using the deterministic key.
	dmKeys := make([]string, 0, len(profileIDs))
	keyByProfile := make(map[int64]string, len(profileIDs))
	for _, pid := range profileIDs {
		key := domain.BuildDmKey(q.ActorProfileID, pid)
		dmKeys = append(dmKeys, key)
		keyByProfile[pid] = key
	}
	conversationByKey, err := h.conversationRepo.ListByDmKeys(ctx, dmKeys)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	conversationIDs := make([]int64, 0, len(conversationByKey))
	for _, c := range conversationByKey {
		conversationIDs = append(conversationIDs, c.ConversationId())
	}
	participantByConversation, err := h.participantRepo.ListByProfileAndConversationIds(ctx, q.ActorProfileID, conversationIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	rows := make([]*ClassroomMemberRow, 0, len(kept))
	for _, m := range kept {
		row := &ClassroomMemberRow{Member: m}

		if p, ok := profileByID[m.ProfileId()]; ok {
			row.Profile = p
			if pr, ok := presenceByUser[p.UserId()]; ok {
				row.Presence = pr
			}
		}
		if c, ok := conversationByKey[keyByProfile[m.ProfileId()]]; ok {
			row.Conversation = c
			if part, ok := participantByConversation[c.ConversationId()]; ok {
				row.Participant = part
			}
		}
		rows = append(rows, row)
	}

	return &ListClassroomMembersResult{Rows: rows, Pagination: pg}, nil
}
