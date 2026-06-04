package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/exercise"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SubmitExerciseAnswersCommand persists a freshly graded submission.
// The expensive bot grading call happens BEFORE this command — the
// AIReview / score numbers arrive already computed so the UoW stays a
// short DB-only insert.
//
// The DB UNIQUE (classroom_exercise_id, profile_id) protects against
// races; a duplicate insert surfaces a generic FAIL — the module layer
// pre-checks for an existing row and maps the duplicate to
// CLASSROOM_EXERCISE_SUBMISSION_ALREADY_EXISTS, but the constraint is
// still the hard backstop.
type SubmitExerciseAnswersCommand struct {
	ActorID             *int64
	ClassroomExerciseID int64
	ClassroomID         int64
	ProfileID           int64
	AnswersJSON         string
	AIReview            *string
	TotalQuestions      *int64
	CorrectNumber       *int64
	ScorePercentage     *int64
	Note                *string
}

type SubmitExerciseAnswersCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSubmitExerciseAnswersCommandHandler(uow transaction.UnitOfWork) *SubmitExerciseAnswersCommandHandler {
	return &SubmitExerciseAnswersCommandHandler{uow: uow}
}

func (h *SubmitExerciseAnswersCommandHandler) Handle(ctx context.Context, cmd SubmitExerciseAnswersCommand) (*domain.Submission, error) {
	var created *domain.Submission

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		submissionID, err := nextSeqID(ctx, repos, seq.NameClassroomExerciseSubmission)
		if err != nil {
			return err
		}

		sub := domain.NewSubmission()
		sub.SetClassroomExerciseSubmissionId(submissionID)
		sub.SetClassroomExerciseId(cmd.ClassroomExerciseID)
		sub.SetClassroomId(cmd.ClassroomID)
		sub.SetProfileId(cmd.ProfileID)
		answers := cmd.AnswersJSON
		sub.SetAnswers(&answers)
		sub.SetAIReview(cmd.AIReview)
		sub.SetTotalQuestions(cmd.TotalQuestions)
		sub.SetCorrectNumber(cmd.CorrectNumber)
		sub.SetScorePercentage(cmd.ScorePercentage)

		now := mtime.Now()
		sub.SetSubmittedDt(now)
		// Treat a non-nil AIReview as the marker that grading happened
		// at submit time. The row goes straight to GRADED so the student
		// sees the result on the same response; a regrade flow can roll
		// it back later if needed.
		statusVal := string(enum.ClassroomExerciseSubmissionStatusSubmitted)
		if cmd.AIReview != nil && strings.TrimSpace(*cmd.AIReview) != "" {
			statusVal = string(enum.ClassroomExerciseSubmissionStatusGraded)
			sub.SetGradedDt(now)
		}
		sub.SetSubmissionStatus(&statusVal)
		sub.SetNote(cmd.Note)
		sub.SetCreateId(cmd.ActorID)

		saved, err := repos.ExerciseSubmission.Create(ctx, sub)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_NOT_FOUND, nil,
				ErrSubmissionNotFoundAfterInsert)
		}
		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
