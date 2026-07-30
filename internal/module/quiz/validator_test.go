package quiz

import (
	"context"
	"testing"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func codeOf(t *testing.T, err error) status.StatusCode {
	t.Helper()
	mErr, ok := errs.IsMathError(err)
	if !ok {
		t.Fatalf("expected MathError, got %v", err)
	}
	return mErr.GetStatusCode()
}

func TestValidateQuizProgress(t *testing.T) {
	ctx := context.Background()
	practice := "practice"
	bad := "SHOPPING"

	t.Run("missing profile", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{ProfileID: 0})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_MISSING_PROFILE {
			t.Fatal("wrong code")
		}
	})

	t.Run("defaults applied", func(t *testing.T) {
		req := &dto.QuizProgressReq{ProfileID: 1, Purpose: &practice}
		if err := ValidateQuizProgress(ctx, req); err != nil {
			t.Fatal(err)
		}
		if req.Limit != ProgressLimitDefault {
			t.Errorf("limit = %d want %d", req.Limit, ProgressLimitDefault)
		}
		if req.Tz != "+07:00" {
			t.Errorf("tz = %q want +07:00", req.Tz)
		}
		if req.Purpose == nil || *req.Purpose != "PRACTICE" {
			t.Errorf("purpose not upper-cased: %v", req.Purpose)
		}
	})

	t.Run("limit clamp", func(t *testing.T) {
		req := &dto.QuizProgressReq{ProfileID: 1, Limit: 9999}
		_ = ValidateQuizProgress(ctx, req)
		if req.Limit != ProgressLimitMax {
			t.Errorf("limit = %d want %d", req.Limit, ProgressLimitMax)
		}
	})

	t.Run("bad purpose", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{ProfileID: 1, Purpose: &bad})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_INVALID_PURPOSE {
			t.Fatal("wrong code")
		}
	})

	t.Run("bad tz", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{ProfileID: 1, Tz: "Asia/Ho_Chi_Minh"})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_INVALID_TZ {
			t.Fatal("wrong code")
		}
	})

	t.Run("reversed range", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{
			ProfileID: 1, FromDt: "2026-06-24 00:00:00", ToDt: "2026-05-01 00:00:00",
		})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_INVALID_DATE_RANGE {
			t.Fatal("wrong code")
		}
	})
}
