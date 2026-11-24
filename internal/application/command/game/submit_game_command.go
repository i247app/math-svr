package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/application/command/shared/scorer"
	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/utils"
)

// SubmitGameCommand grades a round, once.
type SubmitGameCommand struct {
	GsessID   int64
	UID       int64
	ProfileID int64
	Answers   []dto.QuestionAnswer
}

type SubmitGameCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSubmitGameCommandHandler(uow transaction.UnitOfWork) *SubmitGameCommandHandler {
	return &SubmitGameCommandHandler{uow: uow}
}

func (h *SubmitGameCommandHandler) Handle(ctx context.Context, cmd SubmitGameCommand) (*game.GameSession, error) {
	var submitted *game.GameSession
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		s, err := repos.GameSession.FindByGsessId(ctx, cmd.GsessID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if s == nil {
			return errs.NewError(ctx, game.GameSessionNotFound, nil, fmt.Errorf("game: session %d not found", cmd.GsessID))
		}
		if !s.OwnedBy(cmd.UID, cmd.ProfileID) {
			return errs.NewError(ctx, game.GameSessionNotOwned, nil, fmt.Errorf("game: session %d belongs to another profile", cmd.GsessID))
		}
		if utils.DerefString(s.GsessStatus()) != enum.GsessStatusInProgress.String() {
			return errs.NewError(ctx, game.GameAlreadySubmitted, nil, fmt.Errorf("game: session %d is not in progress", cmd.GsessID))
		}

		res, err := grade(ctx, s, cmd.Answers)
		if err != nil {
			return err
		}
		res.ModifyId = &cmd.UID

		// The state guard lives in the UPDATE: of two racing submits the
		// loser ends here and rolls back.
		if err := repos.GameSession.MarkComplete(ctx, s.GsessId(), res); err != nil {
			if errors.Is(err, game.ErrSessionNotInProgress) {
				return errs.NewError(ctx, game.GameAlreadySubmitted, nil, err)
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		submitted, err = repos.GameSession.FindByGsessId(ctx, s.GsessId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return submitted, nil
}

// grade scores the answers over ALL the round's questions: a skipped one
// counts as not correct.
func grade(ctx context.Context, s *game.GameSession, answers []dto.QuestionAnswer) (game.Result, error) {
	var wire []dto.Question
	if err := json.Unmarshal([]byte(s.QuestionsJson()), &wire); err != nil {
		return game.Result{}, errs.NewError(ctx, game.GameGradingFailed, nil, fmt.Errorf("game: parse stored questions: %w", err))
	}
	if err := dto.CheckSet(wire); err != nil {
		return game.Result{}, errs.NewError(ctx, game.GameGradingFailed, nil, err)
	}
	invalid := func(err error) error {
		return errs.NewError(ctx, game.GameInvalidAnswers, map[string]any{"reason": err.Error()}, err)
	}
	if err := checkAnswers(wire, answers); err != nil {
		return game.Result{}, invalid(err)
	}

	// The language only shapes the scorer's review text, which a game drops.
	scored, err := scorer.ScoreQuestions(wire, answers,
		enum.LanguageTypeVietnamese, scorer.DenominatorAllQuestions)
	if err != nil {
		if errors.Is(err, scorer.ErrAnswerShape) {
			return game.Result{}, invalid(err)
		}
		return game.Result{}, errs.NewError(ctx, game.GameGradingFailed, nil, err)
	}

	graded := make([]dto.GameAnswer, 0, len(scored.Outcomes))
	for _, o := range scored.Outcomes {
		graded = append(graded, dto.GameAnswer{
			QuestionNumber:  o.Question.QuestionNumber,
			ChildOptionIDs:  dto.SplitIDs(o.ChildOptionIDs),
			ChildAnswerText: o.ChildAnswerText,
			IsCorrect:       o.IsCorrect,
		})
	}
	raw, err := json.Marshal(graded)
	if err != nil {
		return game.Result{}, errs.NewError(ctx, game.GameGradingFailed, nil, err)
	}
	return game.Result{
		AnswersJson:     string(raw),
		TotalQuestions:  scored.TotalQuestions,
		CorrectNumber:   scored.CorrectNumber,
		SkippedNumber:   scored.SkippedNumber,
		ScorePercentage: scored.ScorePercentage,
		SubmittedDt:     time.Now().UTC(),
	}, nil
}

// checkAnswers refuses an answer to a question the round does not have, a
// second answer to the same question, and an option the question was not
// served with. The scorer checks none of these but the duplicate: it would
// drop the first silently and grade the last as wrong.
func checkAnswers(questions []dto.Question, answers []dto.QuestionAnswer) error {
	byNumber := make(map[int]dto.Question, len(questions))
	for _, q := range questions {
		byNumber[q.QuestionNumber] = q
	}
	seen := make(map[int]bool, len(answers))
	for _, a := range answers {
		q, ok := byNumber[a.QuestionNumber]
		if !ok {
			return fmt.Errorf("game: the round has no question %d", a.QuestionNumber)
		}
		if seen[a.QuestionNumber] {
			return fmt.Errorf("game: question %d is answered twice", a.QuestionNumber)
		}
		seen[a.QuestionNumber] = true
		for _, id := range a.OptionIDs {
			handler := func(o dto.QuestionOption) bool {
				return strings.EqualFold(o.ID, strings.TrimSpace(id))
			}

			if !slices.ContainsFunc(q.Options, handler) {
				return fmt.Errorf("game: option %q is not an option of question %d", id, a.QuestionNumber)
			}
		}
	}
	return nil
}
