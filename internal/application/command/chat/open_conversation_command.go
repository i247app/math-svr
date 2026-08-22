// Package chat holds the write side of messaging.
package chat

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// OpenConversationCommand carries both identities for both sides: the profile
// is who is talking, the user is where realtime events and push are delivered.
// Callers resolve the user ids before reaching here.
type OpenConversationCommand struct {
	ActorProfileID  int64
	ActorUserID     int64
	TargetProfileID int64
	TargetUserID    int64
}

type OpenConversationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewOpenConversationCommandHandler(uow transaction.UnitOfWork) *OpenConversationCommandHandler {
	return &OpenConversationCommandHandler{uow: uow}
}

// Handle resolves the 1-1 thread for the pair, creating it on first contact.
//
// The interesting case is two people tapping "message" at the same instant.
// Both miss the initial lookup, both try to insert, and the UNIQUE index on
// dm_key rejects the loser. That is the desired outcome, not an error to
// surface: the loser re-reads and gets the winner's thread, so both clients
// end up in the same conversation.
func (h *OpenConversationCommandHandler) Handle(ctx context.Context, cmd *OpenConversationCommand) (*domain.Conversation, error) {
	conv, err := h.create(ctx, cmd)
	if err == nil {
		return conv, nil
	}
	if !errors.Is(err, domain.ErrDuplicateConversation) {
		return nil, err
	}

	// Lost the race. The winner's row is committed by now, so a plain read
	// resolves it.
	existing, findErr := h.findByPair(ctx, cmd)
	if findErr != nil {
		return nil, findErr
	}
	if existing == nil {
		// The unique key fired but no visible row matches — the only way this
		// happens is a soft-deleted thread holding the key, which the active
		// filter hides. Report it rather than looping.
		return nil, errs.NewError(ctx, status.CHAT_CONVERSATION_CREATE_FAILED, nil, err)
	}
	return existing, nil
}

func (h *OpenConversationCommandHandler) findByPair(ctx context.Context, cmd *OpenConversationCommand) (*domain.Conversation, error) {
	dmKey := domain.BuildDmKey(cmd.ActorProfileID, cmd.TargetProfileID)

	var found *domain.Conversation
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		c, err := repos.ChatConversation.FindByDmKey(ctx, dmKey)
		if err != nil {
			return errs.NewError(ctx, status.CHAT_CONVERSATION_NOT_FOUND, nil, err)
		}
		found = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

func (h *OpenConversationCommandHandler) create(ctx context.Context, cmd *OpenConversationCommand) (*domain.Conversation, error) {
	dmKey := domain.BuildDmKey(cmd.ActorProfileID, cmd.TargetProfileID)

	var result *domain.Conversation
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ChatConversation.FindByDmKey(ctx, dmKey)
		if err != nil {
			return errs.NewError(ctx, status.CHAT_CONVERSATION_NOT_FOUND, nil, err)
		}
		if existing != nil {
			result = existing
			return nil
		}

		conversationID, err := seqgen.Next(ctx, repos.Seq, seq.NameChatConversation)
		if err != nil {
			return err
		}

		active := string(enum.ChatConversationStatusActive)
		c := domain.NewConversation()
		c.SetConversationId(conversationID)
		c.SetConversationType(string(enum.ChatConversationTypeDirect))
		c.SetDmKey(&dmKey)
		// classroom_id stays NULL: a direct thread is global (decision D2), so
		// the same conversation is reached from every classroom the pair share.
		c.SetParticipantCount(2)
		c.SetConversationStatus(&active)
		c.SetStatus(string(enum.StatusActive))
		c.SetCreateId(&cmd.ActorUserID)

		created, err := repos.ChatConversation.Create(ctx, c)
		if err != nil {
			// Passed through unwrapped so Handle can detect the lost race.
			if errors.Is(err, domain.ErrDuplicateConversation) {
				return err
			}
			return errs.NewError(ctx, status.CHAT_CONVERSATION_CREATE_FAILED, nil, err)
		}

		// Both participant rows are inserted in the same transaction as the
		// conversation: a thread that exists with only one side in it would be
		// invisible to the other person and impossible to repair from the API.
		for _, side := range []struct{ profileID, userID int64 }{
			{cmd.ActorProfileID, cmd.ActorUserID},
			{cmd.TargetProfileID, cmd.TargetUserID},
		} {
			participantID, err := seqgen.Next(ctx, repos.Seq, seq.NameChatParticipant)
			if err != nil {
				return err
			}

			activeParticipant := string(enum.ChatParticipantStatusActive)
			p := domain.NewParticipant()
			p.SetParticipantId(participantID)
			p.SetConversationId(conversationID)
			p.SetProfileId(side.profileID)
			p.SetUserId(side.userID)
			p.SetParticipantRole(string(enum.ChatParticipantRoleMember))
			p.SetJoinedDt(mtime.Now())
			p.SetParticipantStatus(&activeParticipant)
			p.SetStatus(string(enum.StatusActive))
			p.SetCreateId(&cmd.ActorUserID)

			if _, err := repos.ChatParticipant.Create(ctx, p); err != nil {
				return errs.NewError(ctx, status.CHAT_CONVERSATION_CREATE_FAILED, nil, err)
			}
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
