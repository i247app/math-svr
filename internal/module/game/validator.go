package game

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// ValidateStartGame checks the request and fills the default number of
// questions. A number outside the range is refused, not clamped.
func ValidateStartGame(ctx context.Context, req *dto.StartGameReq) error {
	if err := validateProfileID(ctx, req.ProfileID); err != nil {
		return err
	}
	if g := req.Grade; g != nil && (*g < enum.ExamGradeMin || *g > enum.ExamGradeMax) {
		return errs.NewError(ctx, domain.GameInvalidGrade, nil, ErrGradeOutOfRange)
	}
	if l := req.Level; l != nil && (*l < enum.ExamLevelMin || *l > enum.ExamLevelMax) {
		return errs.NewError(ctx, domain.GameInvalidLevel, nil, ErrLevelOutOfRange)
	}
	if req.NumQuestions == nil {
		n := enum.GameNumQuestionsDefault
		req.NumQuestions = &n
	}
	if n := *req.NumQuestions; n < enum.GameNumQuestionsMin || n > enum.GameNumQuestionsMax {
		return errs.NewError(ctx, domain.GameInvalidNumQuestions, nil, ErrNumQuestionsOutOfRange)
	}
	return nil
}

// ValidateGameIDs checks the ids of a submit or a detail. The answers of a
// submit are checked by its command, against the round's own questions.
func ValidateGameIDs(ctx context.Context, profileID, gsessID int64) error {
	if err := validateProfileID(ctx, profileID); err != nil {
		return err
	}
	if gsessID <= 0 {
		return errs.NewError(ctx, domain.GameMissingSessionID, nil, ErrGsessIDRequired)
	}
	return nil
}

func ValidateListGames(ctx context.Context, req *dto.ListGamesReq) error {
	if err := validateProfileID(ctx, req.ProfileID); err != nil {
		return err
	}
	return req.Request.Validate(ctx)
}

func validateProfileID(ctx context.Context, profileID int64) error {
	if profileID <= 0 {
		return errs.NewError(ctx, domain.GameMissingProfileID, nil, ErrProfileIDRequired)
	}
	return nil
}
