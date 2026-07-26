package command

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/application/command/shared/scorer"
	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/exercise"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SubmitExerciseAnswersV2Command is the deterministic counterpart of
// SubmitExerciseAnswersCommand. It scores the student's answers entirely
// in process — no bot adapter call — and persists the row through the
// same repo write path as v1, including the seq-minted external ID.
//
// Compared to v1, the bot grading inputs (AIReview / TotalQuestions /
// ...) are computed inline from the exercise row's `questions` JSON so
// the caller doesn't have to plumb them in.
//
// QuestionsJSON is the exercise row's `questions` column verbatim;
// pass-through so this command stays I/O-free and the service layer can
// keep the FindBy* call where it already happens for the v1 path.
type SubmitExerciseAnswersV2Command struct {
	ActorID             *int64
	ClassroomExerciseID int64
	ClassroomID         int64
	ProfileID           int64
	QuestionsJSON       string
	Answers             []quizDto.QuizStudentAnswer
	Note                *string
	Language            enum.LanguageType
}

type SubmitExerciseAnswersV2CommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSubmitExerciseAnswersV2CommandHandler(uow transaction.UnitOfWork) *SubmitExerciseAnswersV2CommandHandler {
	return &SubmitExerciseAnswersV2CommandHandler{uow: uow}
}

// Handle scores the submission deterministically and inserts the row in
// a single UoW. The seq mint and the insert run on the same tx
// connection so the external ID never leaks past a failure, mirroring
// the v1 command's behaviour.
//
// The lifecycle column lands on GRADED because the deterministic scorer
// always produces a non-empty review — same convention as v1's
// "non-nil AIReview ⇒ GRADED" branch.
func (h *SubmitExerciseAnswersV2CommandHandler) Handle(ctx context.Context, cmd SubmitExerciseAnswersV2Command) (*domain.Submission, error) {
	log := logger.From(ctx)

	answersJSON, err := json.Marshal(cmd.Answers)
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_INVALID_ANSWERS, nil,
			fmt.Errorf("submission v2: marshal answers: %w", err))
	}

	if strings.TrimSpace(cmd.QuestionsJSON) == "" {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED, nil,
			fmt.Errorf("submission v2: exercise has no questions to grade"))
	}

	score, err := scorer.Score(cmd.QuestionsJSON, cmd.Answers, cmd.Language)
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED, nil, err)
	}

	var created *domain.Submission

	err = h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		submissionID, err := seqgen.Next(ctx, repos.Seq, seq.NameClassroomExerciseSubmission)
		if err != nil {
			return err
		}

		sub := domain.NewSubmission()
		sub.SetClassroomExerciseSubmissionId(submissionID)
		sub.SetClassroomExerciseId(cmd.ClassroomExerciseID)
		sub.SetClassroomId(cmd.ClassroomID)
		sub.SetProfileId(cmd.ProfileID)

		ans := string(answersJSON)
		sub.SetAnswers(&ans)

		review := score.AIReview
		sub.SetAIReview(&review)

		total := int64(score.TotalQuestions)
		correct := int64(score.CorrectNumber)
		percentage := int64(score.ScorePercentage)
		sub.SetTotalQuestions(&total)
		sub.SetCorrectNumber(&correct)
		sub.SetScorePercentage(&percentage)

		now := mtime.Now()
		sub.SetSubmittedDt(now)
		// Deterministic scoring always produces a review, so we always
		// land on GRADED with graded_dt set — matching v1's success
		// branch (which writes GRADED when AIReview != "").
		statusVal := string(enum.ClassroomExerciseSubmissionStatusGraded)
		sub.SetSubmissionStatus(&statusVal)
		sub.SetGradedDt(now)

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

	log.Infof("exercise.v2.scored classroom_exercise_id=%d profile_id=%d total=%d correct=%d pct=%d review_source=%s",
		cmd.ClassroomExerciseID, cmd.ProfileID, score.TotalQuestions, score.CorrectNumber, score.ScorePercentage, scorer.ReviewSourceMarker)

	return created, nil
}
