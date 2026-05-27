package quiz

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// DefaultNumQuestions is what GenerateQuiz uses when the request omits
// num_questions (or sends 0). MaxNumQuestions is a soft cap to keep
// prompt size, token cost, and bot latency bounded; values above it are
// silently clamped rather than rejected, since the caller's intent
// ("a longer quiz") is preserved.
const (
	DefaultNumQuestions = 5
	MaxNumQuestions     = 20
)

func validateLanguage(ctx context.Context, lang enum.LanguageType) error {
	if lang == "" {
		return nil
	}
	switch lang {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeEnglish:
		return nil
	default:
		return errs.NewError(ctx, status.QUIZ_INVALID_LANGUAGE, nil,
			errors.New("language must be 'vn' or 'en'"))
	}
}

// validateQuizPurpose normalises the request's purpose to upper-case
// before checking it against the enum. The status code name still reads
// "TYPE" — kept for backward compatibility with existing mobile clients
// that key off the numeric code (11003 / 11004).
func validateQuizPurpose(ctx context.Context, p string) (enum.QuizPurpose, error) {
	upper := enum.QuizPurpose(strings.ToUpper(strings.TrimSpace(p)))
	if upper == "" {
		return "", errs.NewError(ctx, status.QUIZ_MISSING_TYPE, nil,
			errors.New("purpose is required"))
	}
	if !upper.IsValid() {
		return "", errs.NewError(ctx, status.QUIZ_INVALID_TYPE, nil,
			errors.New("purpose must be one of ASSESSMENT, PRACTICE, EXAM"))
	}
	return upper, nil
}

// resolveTypeOfQuiz is intentionally permissive. When the request omits
// type_of_quiz, the service infers it from PreviousQuizID: a previous
// quiz implies REINFORCEMENT, no previous quiz implies GENERAL. When
// supplied, the value is validated against the enum; an unknown value is
// rejected loudly so drift surfaces immediately.
func resolveTypeOfQuiz(ctx context.Context, raw string, hasPrevious bool) (enum.QuizTypeOfQuiz, error) {
	upper := enum.QuizTypeOfQuiz(strings.ToUpper(strings.TrimSpace(raw)))
	if upper == "" {
		if hasPrevious {
			return enum.QuizTypeOfQuizReinforcement, nil
		}
		return enum.QuizTypeOfQuizGeneral, nil
	}
	if !upper.IsValid() {
		return "", errs.NewError(ctx, status.QUIZ_INVALID_TYPE_OF_QUIZ, nil,
			errors.New("type_of_quiz must be one of GENERAL, REINFORCEMENT"))
	}
	return upper, nil
}

// validatedGenerateQuiz bundles the normalised values the service needs
// downstream so callers don't have to re-parse req strings.
type validatedGenerateQuiz struct {
	Purpose    enum.QuizPurpose
	TypeOfQuiz enum.QuizTypeOfQuiz
}

func ValidateGenerateQuiz(ctx context.Context, req *dto.GenerateQuizReq) (validatedGenerateQuiz, error) {
	if req.ProfileID != nil && *req.ProfileID == "" {
		req.ProfileID = nil
	}
	if err := validateLanguage(ctx, req.Language); err != nil {
		return validatedGenerateQuiz{}, err
	}
	purpose, err := validateQuizPurpose(ctx, req.Purpose)
	if err != nil {
		return validatedGenerateQuiz{}, err
	}
	typeOfQuiz, err := resolveTypeOfQuiz(ctx, req.TypeOfQuiz, req.PreviousQuizID != nil)
	if err != nil {
		return validatedGenerateQuiz{}, err
	}
	if req.NumQuestions <= 0 {
		req.NumQuestions = DefaultNumQuestions
	} else if req.NumQuestions > MaxNumQuestions {
		req.NumQuestions = MaxNumQuestions
	}
	return validatedGenerateQuiz{Purpose: purpose, TypeOfQuiz: typeOfQuiz}, nil
}

func ValidateSubmitAnswers(ctx context.Context, req *dto.SubmitQuizAnswersReq) error {
	if req.QuizID == "" {
		return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			errors.New("quiz_id is required"))
	}
	if err := validateLanguage(ctx, req.Language); err != nil {
		return err
	}
	if len(req.Answers) == 0 {
		return errs.NewError(ctx, status.QUIZ_MISSING_ANSWERS, nil,
			errors.New("answers are required"))
	}
	for i, a := range req.Answers {
		if a.QuestionNumber <= 0 {
			return errs.NewError(ctx, status.QUIZ_INVALID_ANSWERS,
				map[string]any{"index": i}, errors.New("question_number must be positive"))
		}
		if strings.TrimSpace(a.Label) == "" {
			return errs.NewError(ctx, status.QUIZ_INVALID_ANSWERS,
				map[string]any{"index": i}, errors.New("label is required"))
		}
	}
	return nil
}

func ValidateGetQuiz(ctx context.Context, req *dto.GetQuizByQuizIdReq) error {
	if req.QuizID == "" {
		return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			errors.New("quiz_id is required"))
	}
	return nil
}

func ValidateListQuizzes(ctx context.Context, req *dto.ListQuizzesReq) error {
	if req.ProfileID != nil && *req.ProfileID == "" {
		req.ProfileID = nil
	}
	if req.UserID != nil && *req.UserID == "" {
		req.UserID = nil
	}
	if req.ProfileID == nil && req.UserID == nil {
		return errs.NewError(ctx, status.QUIZ_MISSING_PROFILE_ID, nil,
			errors.New("at least one of profile_id or user_id is required"))
	}
	return nil
}

func ValidateDeleteQuiz(ctx context.Context, req *dto.DeleteQuizReq) error {
	if req.QuizID == "" {
		return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			errors.New("quiz_id is required"))
	}
	return nil
}
