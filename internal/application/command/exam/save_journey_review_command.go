package command

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// SaveJourneyReviewCommand stores the AI review of one journey row. The
// model call happens before, outside any transaction; this only writes
// its result — and only while the journey is still COMPLETE: one reopened
// during the model call is being worked again, and the review would
// describe a run that is no longer finished. The caller has already
// proved the journey belongs to the child, but ownership is part of the
// row read here too, so the write can never land on another family's
// journey.
type SaveJourneyReviewCommand struct {
	EsessID     int64
	ExamType    string // the journey's own type (ASSESSMENT or GRADE)
	UID         int64
	ProfileID   int64
	ReviewShort string
	ReviewLong  string
}

type SaveJourneyReviewCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSaveJourneyReviewCommandHandler(uow transaction.UnitOfWork) *SaveJourneyReviewCommandHandler {
	return &SaveJourneyReviewCommandHandler{uow: uow}
}

// Handle writes the review and returns the journey as it now reads.
func (h *SaveJourneyReviewCommandHandler) Handle(ctx context.Context, cmd SaveJourneyReviewCommand) (*exam.ExamSession, error) {
	var updated *exam.ExamSession

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		journey, err := repos.ExamSession.FindByEsessIdAndType(ctx, cmd.EsessID, cmd.ExamType)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if journey == nil {
			return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil,
				fmt.Errorf("exam: journey %d (%s) not found", cmd.EsessID, cmd.ExamType))
		}
		if journey.Uid() != cmd.UID || journey.ProfileId() != cmd.ProfileID {
			return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_OWNED, nil,
				fmt.Errorf("exam: journey %d belongs to another profile", cmd.EsessID))
		}
		if st := utils.DerefString(journey.EsessStatus()); st != string(enum.EsessStatusComplete) {
			return errs.NewError(ctx, status.EXAM_REVIEW_JOURNEY_NOT_COMPLETE, nil,
				fmt.Errorf("exam: journey %d is %s; only a COMPLETE journey is reviewed", cmd.EsessID, st))
		}

		if err := repos.ExamSession.SetAiReview(ctx, cmd.EsessID, cmd.ExamType, cmd.ReviewShort, cmd.ReviewLong); err != nil {
			if errors.Is(err, exam.ErrJourneyNotFound) {
				return errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil, err)
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		short, long := cmd.ReviewShort, cmd.ReviewLong
		journey.SetAiReviewShort(&short)
		journey.SetAiReviewLong(&long)
		updated = journey
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
