package command

import (
	"context"
	"encoding/json"
	"fmt"

	"math-ai.com/math-ai/internal/application/command/shared/scorer"
	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// SubmitQuizAnswersV2Command is the deterministic counterpart of
// SubmitQuizAnswersCommand. It scores the student's answers entirely in
// process — no bot adapter call — and persists the result through the
// same repo write path as v1 (UpdateAnswersAndGrading), so downstream
// readers see one row shape regardless of which submit path produced it.
//
// Answers is the parsed student payload from /quizzes/submit/v2; Language
// drives the language of the deterministic review string.
type SubmitQuizAnswersV2Command struct {
	QuizID   int64
	Answers  []quizDto.QuizStudentAnswer
	Language enum.LanguageType
}

type SubmitQuizAnswersV2CommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSubmitQuizAnswersV2CommandHandler(uow transaction.UnitOfWork) *SubmitQuizAnswersV2CommandHandler {
	return &SubmitQuizAnswersV2CommandHandler{uow: uow}
}

// Handle scores the quiz in process and writes the answers + grading
// fields in a single UoW. The scoring step runs INSIDE the tx because it
// is pure CPU work with negligible cost — unlike v1's bot call, there's
// no reason to split the read from the write. Mints nothing new from
// ma_seqs (v1 doesn't either; submit is an update, not a create).
func (h *SubmitQuizAnswersV2CommandHandler) Handle(ctx context.Context, cmd SubmitQuizAnswersV2Command) (*quiz.Quiz, error) {
	log := logger.From(ctx)

	answersJSON, err := json.Marshal(cmd.Answers)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_INVALID_ANSWERS, nil,
			fmt.Errorf("quiz v2: marshal answers: %w", err))
	}

	var updated *quiz.Quiz

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Quiz.FindByQuizId(ctx, cmd.QuizID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil, nil)
		}
		if s := existing.QuizStatus(); s != nil && *s == string(enum.QuizStatusTypeSubmitted) {
			return errs.NewError(ctx, status.QUIZ_ALREADY_SUBMITTED, nil, nil)
		}
		if existing.Questions() == nil || *existing.Questions() == "" {
			return errs.NewError(ctx, status.QUIZ_GRADING_FAILED, nil,
				fmt.Errorf("quiz v2: quiz %d has no questions", cmd.QuizID))
		}

		score, err := scorer.Score(*existing.Questions(), cmd.Answers, cmd.Language)
		if err != nil {
			return errs.NewError(ctx, status.QUIZ_GRADING_FAILED, nil, err)
		}

		log.Infof("quiz.v2.scored quiz_id=%d total=%d correct=%d pct=%d review_source=%s",
			cmd.QuizID, score.TotalQuestions, score.CorrectNumber, score.ScorePercentage, scorer.ReviewSourceMarker)

		grading := quiz.GradingUpdate{
			AIReview:        score.AIReview,
			AIDetectGrade:   score.AIDetectGrade,
			TotalQuestions:  &score.TotalQuestions,
			CorrectNumber:   &score.CorrectNumber,
			ScorePercentage: &score.ScorePercentage,
		}

		if err := repos.Quiz.UpdateAnswersAndGrading(ctx, cmd.QuizID, string(answersJSON), grading, string(enum.QuizStatusTypeSubmitted)); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		fresh, err := repos.Quiz.FindByQuizId(ctx, cmd.QuizID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if fresh == nil {
			return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil, nil)
		}
		updated = fresh
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return updated, nil
}
