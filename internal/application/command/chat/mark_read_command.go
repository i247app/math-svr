package chat

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type MarkReadCommand struct {
	ConversationID int64
	ProfileID      int64
	SeqNo          int64
}

type MarkReadCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkReadCommandHandler(uow transaction.UnitOfWork) *MarkReadCommandHandler {
	return &MarkReadCommandHandler{uow: uow}
}

// Handle advances the caller's read watermark. The forward-only guarantee is
// enforced by the repository's SQL predicate, not here — putting it in Go
// would leave a window between the read and the write where a concurrent call
// could still drag the watermark backwards.
//
// It returns the watermark actually in effect afterwards, which may be higher
// than the requested value when a later call already overtook this one.
func (h *MarkReadCommandHandler) Handle(ctx context.Context, cmd *MarkReadCommand) (int64, error) {
	var effective int64

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		participant, err := repos.ChatParticipant.FindByConversationAndProfile(ctx, cmd.ConversationID, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.CHAT_NOT_PARTICIPANT, nil, err)
		}
		if participant == nil || participant.ParticipantStatus() == nil ||
			*participant.ParticipantStatus() != string(enum.ChatParticipantStatusActive) {
			return errs.NewError(ctx, status.CHAT_NOT_PARTICIPANT, nil, errNotParticipant)
		}

		if err := repos.ChatParticipant.MarkRead(ctx, cmd.ConversationID, cmd.ProfileID, cmd.SeqNo, nil, mtime.Now()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		updated, err := repos.ChatParticipant.FindByConversationAndProfile(ctx, cmd.ConversationID, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if updated != nil {
			effective = updated.LastReadSeqNo()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return effective, nil
}
