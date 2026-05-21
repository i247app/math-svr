package quiz

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
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

func validateQuizType(ctx context.Context, t string) (enum.QuizType, error) {
	upper := enum.QuizType(strings.ToUpper(strings.TrimSpace(t)))
	if upper == "" {
		return "", errs.NewError(ctx, status.QUIZ_MISSING_TYPE, nil,
			errors.New("type is required"))
	}
	if !upper.IsValid() {
		return "", errs.NewError(ctx, status.QUIZ_INVALID_TYPE, nil,
			errors.New("type must be one of ASSESSMENT, PRACTICE"))
	}
	return upper, nil
}

func ValidateGenerateQuiz(ctx context.Context, req *dto.GenerateQuizReq) (enum.QuizType, error) {
	if req.ProfileID == uuid.Nil {
		return "", errs.NewError(ctx, status.QUIZ_MISSING_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if err := validateLanguage(ctx, req.Language); err != nil {
		return "", err
	}
	typ, err := validateQuizType(ctx, req.Type)
	if err != nil {
		return "", err
	}
	return typ, nil
}

func ValidateSubmitAnswers(ctx context.Context, req *dto.SubmitQuizAnswersReq) error {
	if req.QuizID == uuid.Nil {
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
	if req.QuizID == uuid.Nil {
		return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			errors.New("quiz_id is required"))
	}
	return nil
}

func ValidateListQuizzes(ctx context.Context, req *dto.ListQuizzesByProfileIdReq) error {
	if req.ProfileID == uuid.Nil {
		return errs.NewError(ctx, status.QUIZ_MISSING_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	return nil
}

func ValidateDeleteQuiz(ctx context.Context, req *dto.DeleteQuizReq) error {
	if req.QuizID == uuid.Nil {
		return errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			errors.New("quiz_id is required"))
	}
	return nil
}
