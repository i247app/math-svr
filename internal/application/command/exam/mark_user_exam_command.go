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
	"math-ai.com/math-ai/internal/shared/utils"
)

// MarkUserExamCommand moves one journey between lifecycle states.
//
// Two moves exist, and they are each other's inverse:
//
//   - COMPLETE / CANCEL end an open journey. The next submission of that
//     exam type opens a fresh row, which is what lets a child start over
//     without losing the record of the last run.
//   - ACTIVE reopens an ended journey, for the child who changes their
//     mind. Only one journey of a type may be open at a time, so this is
//     refused while another is — end that one first.
//
// Both moves address the ASSESSMENT row and carry the PRACTICE row that
// shares its id along with it; a PRACTICE row is never marked on its own.
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

	if !cmd.Status.IsMarkable() {
		return nil, errs.NewError(ctx, status.EXAM_INVALID_JOURNEY_STATUS, nil,
			fmt.Errorf("exam: %q is not a status a journey can be marked with", string(cmd.Status)))
	}

	var updated *exam.UserExam

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		journey, err := h.loadOwnedJourney(ctx, repos, cmd)
		if err != nil {
			return err
		}

		if cmd.Status == enum.UserExamStatusActive {
			err = h.reopen(ctx, repos, cmd, journey)
		} else {
			err = h.end(ctx, repos, cmd, journey)
		}
		if err != nil {
			return err
		}

		fresh, err := repos.UserExam.FindByUserExamIdAndType(ctx, journey.UserExamId(), journey.ReqExamType())
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

// loadOwnedJourney reads the ASSESSMENT row — the one that owns the
// lifecycle — and proves it belongs to this child. Ownership is checked
// on BOTH ids: a profile id alone is guessable, and the session only
// proves the user.
func (h *MarkUserExamCommandHandler) loadOwnedJourney(ctx context.Context, repos transaction.Repositories, cmd MarkUserExamCommand) (*exam.UserExam, error) {
	journey, err := repos.UserExam.FindByUserExamIdAndType(ctx, cmd.UserExamID, string(enum.ExamTypeAssessment))
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if journey == nil {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil,
			fmt.Errorf("exam: journey %d not found", cmd.UserExamID))
	}
	if journey.UserId() != cmd.UserID || journey.ProfileId() != cmd.ProfileID {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_OWNED, nil,
			fmt.Errorf("exam: journey %d belongs to another profile", cmd.UserExamID))
	}
	return journey, nil
}

// end closes an open journey.
func (h *MarkUserExamCommandHandler) end(ctx context.Context, repos transaction.Repositories, cmd MarkUserExamCommand, journey *exam.UserExam) error {
	if !hasStatus(journey, enum.UserExamStatusActive) {
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
	return nil
}

// reopen puts an ended journey back in play.
//
// The "no other journey open" rule is checked twice on purpose. The read
// here gives the client a clean answer in the ordinary case; the unique
// key the repository trips under a race is what actually holds the rule,
// and it reports back through ErrJourneyConflict.
func (h *MarkUserExamCommandHandler) reopen(ctx context.Context, repos transaction.Repositories, cmd MarkUserExamCommand, journey *exam.UserExam) error {
	if hasStatus(journey, enum.UserExamStatusActive) {
		return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ACTIVE, nil,
			fmt.Errorf("exam: journey %d is already active", cmd.UserExamID))
	}
	if !enum.UserExamStatusType(utils.DerefString(journey.UserExamStatus())).IsEnding() {
		return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil,
			fmt.Errorf("exam: journey %d is in state %q and cannot be reopened", cmd.UserExamID, utils.DerefString(journey.UserExamStatus())))
	}

	open, err := repos.UserExam.FindActiveByUserProfileType(ctx, cmd.UserID, cmd.ProfileID, journey.ReqExamType())
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open != nil {
		return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ACTIVE, nil,
			fmt.Errorf("exam: journey %d is open; end it before reopening %d", open.UserExamId(), cmd.UserExamID))
	}

	if err := repos.UserExam.Reopen(ctx, journey.UserExamId()); err != nil {
		switch {
		case errors.Is(err, exam.ErrJourneyConflict):
			// Another journey took the open slot between the read above
			// and this write.
			return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ACTIVE, nil, err)
		case errors.Is(err, exam.ErrJourneyNotEnded):
			// The row changed under us — most likely reopened by a
			// concurrent mark, so "already active" is the honest answer.
			return errs.NewError(ctx, status.EXAM_JOURNEY_ALREADY_ACTIVE, nil, err)
		}
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	return nil
}

func hasStatus(journey *exam.UserExam, want enum.UserExamStatusType) bool {
	return utils.DerefString(journey.UserExamStatus()) == string(want)
}
