package exam

import (
	"context"
	"testing"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

func codeOf(t *testing.T, err error) status.StatusCode {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	mErr, ok := errs.IsMathError(err)
	if !ok {
		t.Fatalf("expected a MathError, got %T", err)
	}
	return mErr.GetStatusCode()
}

func TestValidateGenerateExam(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects a missing profile", func(t *testing.T) {
		req := &dto.GenerateExamReq{ExamType: "PRACTICE"}
		if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_MISSING_PROFILE_ID {
			t.Errorf("code = %d, want EXAM_MISSING_PROFILE_ID", got)
		}
	})

	t.Run("rejects an unknown exam type", func(t *testing.T) {
		req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "HOMEWORK"}
		if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_INVALID_EXAM_TYPE {
			t.Errorf("code = %d, want EXAM_INVALID_EXAM_TYPE", got)
		}
	})

	t.Run("rejects a grade outside the served bands", func(t *testing.T) {
		for _, grade := range []int{-1, 6} {
			g := grade
			req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "PRACTICE", Grade: &g}
			if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_INVALID_GRADE {
				t.Errorf("grade %d: code = %d, want EXAM_INVALID_GRADE", grade, got)
			}
		}
	})

	t.Run("rejects a level outside the scale", func(t *testing.T) {
		for _, level := range []int{0, 11, -3} {
			l := level
			req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "PRACTICE", Level: &l}
			if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_INVALID_LEVEL {
				t.Errorf("level %d: code = %d, want EXAM_INVALID_LEVEL", level, got)
			}
		}
	})

	t.Run("accepts a level inside the scale", func(t *testing.T) {
		for _, level := range []int{1, 5, 10} {
			l := level
			req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "PRACTICE", Level: &l}
			if _, err := ValidateGenerateExam(ctx, req); err != nil {
				t.Errorf("level %d: unexpected error %v", level, err)
			}
		}
	})

	t.Run("normalises the type and clamps the length", func(t *testing.T) {
		req := &dto.GenerateExamReq{ProfileID: 1, ExamType: " assessment ", NumQuestions: 500}
		got, err := ValidateGenerateExam(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ExamType != enum.ExamTypeAssessment {
			t.Errorf("ExamType = %q, want ASSESSMENT", got.ExamType)
		}
		if req.NumQuestions != MaxNumQuestions {
			t.Errorf("NumQuestions = %d, want it clamped to %d", req.NumQuestions, MaxNumQuestions)
		}
	})

	t.Run("fills in the default length", func(t *testing.T) {
		req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "PRACTICE"}
		if _, err := ValidateGenerateExam(ctx, req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.NumQuestions != DefaultNumQuestions {
			t.Errorf("NumQuestions = %d, want %d", req.NumQuestions, DefaultNumQuestions)
		}
	})
}

func mustFail(t *testing.T, ctx context.Context, req *dto.GenerateExamReq) error {
	t.Helper()
	_, err := ValidateGenerateExam(ctx, req)
	return err
}

func TestValidateSubmitExam(t *testing.T) {
	ctx := context.Background()
	answer := func(n int, label string) question.StudentAnswer {
		return question.StudentAnswer{QuestionNumber: n, Label: label}
	}

	tests := []struct {
		name string
		req  *dto.SubmitExamReq
		want status.StatusCode
	}{
		{
			name: "missing profile",
			req:  &dto.SubmitExamReq{UserAiExamID: 1, Answers: []question.StudentAnswer{answer(1, "A")}},
			want: status.EXAM_MISSING_PROFILE_ID,
		},
		{
			name: "missing attempt id",
			req:  &dto.SubmitExamReq{ProfileID: 1, Answers: []question.StudentAnswer{answer(1, "A")}},
			want: status.EXAM_MISSING_ATTEMPT_ID,
		},
		{
			name: "no answers at all",
			req:  &dto.SubmitExamReq{ProfileID: 1, UserAiExamID: 1},
			want: status.EXAM_MISSING_ANSWERS,
		},
		{
			name: "question number must be positive",
			req:  &dto.SubmitExamReq{ProfileID: 1, UserAiExamID: 1, Answers: []question.StudentAnswer{answer(0, "A")}},
			want: status.EXAM_INVALID_ANSWERS,
		},
		{
			name: "label cannot be blank",
			req:  &dto.SubmitExamReq{ProfileID: 1, UserAiExamID: 1, Answers: []question.StudentAnswer{answer(1, "  ")}},
			want: status.EXAM_INVALID_ANSWERS,
		},
		{
			// Caught here rather than in the scorer so the client gets a
			// field error instead of a grading failure.
			name: "duplicate question number",
			req: &dto.SubmitExamReq{ProfileID: 1, UserAiExamID: 1, Answers: []question.StudentAnswer{
				answer(1, "A"), answer(1, "B"),
			}},
			want: status.EXAM_INVALID_ANSWERS,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := codeOf(t, ValidateSubmitExam(ctx, tc.req)); got != tc.want {
				t.Errorf("code = %d, want %d", got, tc.want)
			}
		})
	}

	t.Run("accepts a partial submission", func(t *testing.T) {
		// Answering three of ten is legitimate in this model — the score is
		// over what was answered, and the blanks are counted separately.
		req := &dto.SubmitExamReq{ProfileID: 1, UserAiExamID: 1, Answers: []question.StudentAnswer{
			answer(1, "A"), answer(4, "B"), answer(9, "C"),
		}}
		if err := ValidateSubmitExam(ctx, req); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
