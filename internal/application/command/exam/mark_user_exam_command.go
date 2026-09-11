package command

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// MarkUserExamCommand ends one journey.
//
// Ending is the only mutation a client can make to a journey directly,
// and it is one-way: an ended journey is never reopened. The next
// submission of that exam type opens a fresh row instead, which is what
// lets a child start over without losing the record of the last run.
type MarkUserExamCommand struct {
	UserExamID int64
	UserID     int64
	ProfileID  int64
	Status     enum.UserExamStatusType
}

type MarkUserExamCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkUserExamCommandHandler(uow transaction.UnitOfWork) *MarkUserExamCommandHandler {
	return &MarkUserExamCommandHandler{uow: uow}
}

func (h *MarkUserExamCommandHandler) Handle(ctx context.Context, cmd MarkUserExamCommand) (*exam.UserExam, error) {
	log := logger.From(ctx)

	if !cmd.Status.IsEnding() {
		return nil, errs.NewError(ctx, status.EXAM_INVALID_JOURNEY_STATUS, nil,
			fmt.Errorf("exam: %q is not a status a journey can be marked with", string(cmd.Status)))
	}

	var updated *exam.UserExam

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		journey, err := repos.UserExam.FindByUserExamId(ctx, cmd.UserExamID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if journey == nil {
			return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil,
				fmt.Errorf("exam: journey %d not found", cmd.UserExamID))
		}
		// Ownership on BOTH ids: a profile id alone is guessable, and the
		// session only proves the user.
		if journey.UserId() != cmd.UserID || journey.ProfileId() != cmd.ProfileID {
			return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_OWNED, nil,
				fmt.Errorf("exam: journey %d belongs to another profile", cmd.UserExamID))
		}
		if s := journey.UserExamStatus(); s == nil || *s != string(enum.UserExamStatusActive) {
			return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ENDED, nil,
				fmt.Errorf("exam: journey %d is not active", cmd.UserExamID))
		}

		if err := repos.UserExam.MarkStatus(ctx, journey.UserExamId(), string(cmd.Status), mtime.Now()); err != nil {
			// The read above and this write are not atomic with respect to
			// another mark; the WHERE clause in MarkStatus is. This is how
			// the loser of that race finds out.
			if errors.Is(err, exam.ErrJourneyNotActive) {
				return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ENDED, nil, err)
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		fresh, err := repos.UserExam.FindByUserExamId(ctx, journey.UserExamId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if fresh == nil {
			return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil, nil)
		}
		updated = fresh

		log.Infof("exam.journey.marked user_exam_id=%d profile=%d type=%s status=%s",
			journey.UserExamId(), cmd.ProfileID, journey.ReqExamType(), cmd.Status)
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return updated, nil
}
