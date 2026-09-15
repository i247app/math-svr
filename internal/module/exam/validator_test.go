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
			req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "ASSESSMENT", Grade: &g}
			if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_INVALID_GRADE {
				t.Errorf("grade %d: code = %d, want EXAM_INVALID_GRADE", grade, got)
			}
		}
	})

	t.Run("level must sit on the 1..10 scale", func(t *testing.T) {
		for _, level := range []int{0, 11} {
			l := level
			req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "ASSESSMENT", Level: &l}
			if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_INVALID_LEVEL {
				t.Errorf("level %d: code = %d, want EXAM_INVALID_LEVEL", level, got)
			}
		}
		for _, level := range []int{1, 10} {
			l := level
			if _, err := ValidateGenerateExam(ctx, &dto.GenerateExamReq{ProfileID: 1, ExamType: "ASSESSMENT", Level: &l}); err != nil {
				t.Errorf("level %d must pass, got %v", level, err)
			}
		}
	})

	t.Run("a PRACTICE round must name its journey", func(t *testing.T) {
		zero := int64(0)
		for _, id := range []*int64{nil, &zero} {
			req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "PRACTICE", UserExamID: id}
			if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_MISSING_JOURNEY_ID {
				t.Errorf("user_exam_id %v: code = %d, want EXAM_MISSING_JOURNEY_ID", id, got)
			}
		}
		one := int64(1)
		if _, err := ValidateGenerateExam(ctx, &dto.GenerateExamReq{ProfileID: 1, ExamType: "practice", UserExamID: &one}); err != nil {
			t.Errorf("a PRACTICE round with a journey must pass, got %v", err)
		}
	})

	t.Run("EXAM is no longer a type", func(t *testing.T) {
		req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "EXAM"}
		if got := codeOf(t, mustFail(t, ctx, req)); got != status.EXAM_INVALID_EXAM_TYPE {
			t.Errorf("code = %d, want EXAM_INVALID_EXAM_TYPE", got)
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
		req := &dto.GenerateExamReq{ProfileID: 1, ExamType: "ASSESSMENT"}
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

func TestValidateMarkExamJourney(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		req      *dto.MarkExamJourneyReq
		wantCode status.StatusCode
		wantSt   enum.UserExamStatusType
	}{
		{"missing profile", &dto.MarkExamJourneyReq{UserExamID: 1, Status: "COMPLETE"}, status.EXAM_MISSING_PROFILE_ID, ""},
		{"missing journey id", &dto.MarkExamJourneyReq{ProfileID: 1, Status: "COMPLETE"}, status.EXAM_MISSING_JOURNEY_ID, ""},
		{"empty status", &dto.MarkExamJourneyReq{ProfileID: 1, UserExamID: 1}, status.EXAM_INVALID_JOURNEY_STATUS, ""},
		{"active reopens, normalised", &dto.MarkExamJourneyReq{ProfileID: 1, UserExamID: 1, Status: " active "}, 0, enum.UserExamStatusActive},
		{"DELETED is not an ending", &dto.MarkExamJourneyReq{ProfileID: 1, UserExamID: 1, Status: "DELETED"}, status.EXAM_INVALID_JOURNEY_STATUS, ""},
		{"unknown word", &dto.MarkExamJourneyReq{ProfileID: 1, UserExamID: 1, Status: "DONE"}, status.EXAM_INVALID_JOURNEY_STATUS, ""},
		{"complete, normalised", &dto.MarkExamJourneyReq{ProfileID: 1, UserExamID: 1, Status: " complete "}, 0, enum.UserExamStatusComplete},
		{"cancel, normalised", &dto.MarkExamJourneyReq{ProfileID: 1, UserExamID: 1, Status: "cancel"}, 0, enum.UserExamStatusCancel},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateMarkExamJourney(ctx, tc.req)
			if tc.wantCode != 0 {
				if code := codeOf(t, err); code != tc.wantCode {
					t.Errorf("code = %d, want %d", code, tc.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Status != tc.wantSt {
				t.Errorf("Status = %q, want %q", got.Status, tc.wantSt)
			}
			if tc.req.Status != string(tc.wantSt) {
				t.Errorf("request status not normalised in place: %q", tc.req.Status)
			}
		})
	}
}

func TestValidateGetExamStatsStatusFilter(t *testing.T) {
	ctx := context.Background()

	t.Run("blank status means no filter", func(t *testing.T) {
		blank := "  "
		req := &dto.GetExamStatsReq{ProfileID: 1, Status: &blank}
		if err := ValidateGetExamStats(ctx, req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Status != nil {
			t.Errorf("blank status should be dropped, got %q", *req.Status)
		}
	})

	t.Run("known status is normalised", func(t *testing.T) {
		raw := "active"
		req := &dto.GetExamStatsReq{ProfileID: 1, Status: &raw}
		if err := ValidateGetExamStats(ctx, req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Status == nil || *req.Status != "ACTIVE" {
			t.Errorf("Status = %v, want ACTIVE", req.Status)
		}
	})

	t.Run("unknown status is rejected", func(t *testing.T) {
		raw := "OPEN"
		req := &dto.GetExamStatsReq{ProfileID: 1, Status: &raw}
		if code := codeOf(t, ValidateGetExamStats(ctx, req)); code != status.EXAM_INVALID_JOURNEY_STATUS {
			t.Errorf("code = %d, want EXAM_INVALID_JOURNEY_STATUS", code)
		}
	})
}

func TestValidateGetExamAcceptsExactlyOneID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		req      *dto.GetExamReq
		wantCode status.StatusCode
	}{
		{"missing profile", &dto.GetExamReq{UserAiExamID: 1}, status.EXAM_MISSING_PROFILE_ID},
		{"neither id", &dto.GetExamReq{ProfileID: 1}, status.EXAM_MISSING_ATTEMPT_ID},
		// Both ids is accepted; the service reads the journey (user_exam_id
		// takes precedence). Rejecting it was considered and dropped.
		{"both ids", &dto.GetExamReq{ProfileID: 1, UserAiExamID: 1, UserExamID: 2}, 0},
		{"a sitting", &dto.GetExamReq{ProfileID: 1, UserAiExamID: 1}, 0},
		{"a journey", &dto.GetExamReq{ProfileID: 1, UserExamID: 2}, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGetExam(ctx, tc.req)
			if tc.wantCode == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if code := codeOf(t, err); code != tc.wantCode {
				t.Errorf("code = %d, want %d", code, tc.wantCode)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
func i64Ptr(v int64) *int64   { return &v }

// TestValidateGetExamJourneyRow: a journey read settles which row of the
// journey it wants. Absent means ASSESSMENT — the row every client that
// predates PRACTICE has always been reading.
func TestValidateGetExamJourneyRow(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		in       *string
		wantType string
		wantCode status.StatusCode
	}{
		{"absent defaults to ASSESSMENT", nil, "ASSESSMENT", 0},
		{"blank defaults to ASSESSMENT", strPtr("  "), "ASSESSMENT", 0},
		{"practice, normalised", strPtr(" practice "), "PRACTICE", 0},
		{"unknown row", strPtr("HOMEWORK"), "", status.EXAM_INVALID_EXAM_TYPE},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &dto.GetExamReq{ProfileID: 1, UserExamID: 2, ExamType: tc.in}
			err := ValidateGetExam(ctx, req)
			if tc.wantCode != 0 {
				if code := codeOf(t, err); code != tc.wantCode {
					t.Errorf("code = %d, want %d", code, tc.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if req.ExamType == nil || *req.ExamType != tc.wantType {
				t.Errorf("ExamType = %v, want %q", req.ExamType, tc.wantType)
			}
		})
	}

	t.Run("a sitting read leaves req_exam_type alone", func(t *testing.T) {
		req := &dto.GetExamReq{ProfileID: 1, UserAiExamID: 1, ExamType: strPtr("HOMEWORK")}
		if err := ValidateGetExam(ctx, req); err != nil {
			t.Fatalf("req_exam_type is meaningless for a sitting and must not be validated: %v", err)
		}
	})
}

// TestPracticeReadsNeedAJourney: history reads of PRACTICE rounds only
// make sense inside one journey, and stats never lists PRACTICE as a
// journey of its own.
func TestPracticeReadsNeedAJourney(t *testing.T) {
	ctx := context.Background()

	t.Run("list PRACTICE without a journey", func(t *testing.T) {
		req := &dto.ListExamsReq{ProfileID: 1, ExamType: strPtr("practice")}
		if code := codeOf(t, ValidateListExams(ctx, req)); code != status.EXAM_MISSING_JOURNEY_ID {
			t.Errorf("code = %d, want EXAM_MISSING_JOURNEY_ID", code)
		}
	})
	t.Run("list PRACTICE of a journey", func(t *testing.T) {
		req := &dto.ListExamsReq{ProfileID: 1, ExamType: strPtr("practice"), UserExamID: i64Ptr(9)}
		if err := ValidateListExams(ctx, req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *req.ExamType != "PRACTICE" {
			t.Errorf("exam_type = %q, want normalised PRACTICE", *req.ExamType)
		}
	})
	t.Run("list ASSESSMENT needs no journey", func(t *testing.T) {
		if err := ValidateListExams(ctx, &dto.ListExamsReq{ProfileID: 1, ExamType: strPtr("ASSESSMENT")}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("progress PRACTICE without a journey", func(t *testing.T) {
		req := &dto.ExamProgressReq{ProfileID: 1, ExamType: strPtr("PRACTICE")}
		if code := codeOf(t, ValidateExamProgress(ctx, req)); code != status.EXAM_MISSING_JOURNEY_ID {
			t.Errorf("code = %d, want EXAM_MISSING_JOURNEY_ID", code)
		}
	})
	t.Run("stats refuses PRACTICE as a journey type", func(t *testing.T) {
		req := &dto.GetExamStatsReq{ProfileID: 1, ExamType: strPtr("PRACTICE")}
		if code := codeOf(t, ValidateGetExamStats(ctx, req)); code != status.EXAM_INVALID_EXAM_TYPE {
			t.Errorf("code = %d, want EXAM_INVALID_EXAM_TYPE", code)
		}
	})
}

// TestResolveLevel: the level is clamped once, here, for a GRADE review
// only — kindergarten tops out at 4 — and passed through untouched as a
// record for every other type.
func TestResolveLevel(t *testing.T) {
	ctx := context.Background()
	nine := 9

	if got := resolveLevel(ctx, enum.ExamTypeGrade, 0, &nine); got == nil || *got != 4 {
		t.Errorf("GRADE at kindergarten with level 9 → %v, want clamped 4", got)
	}
	if got := resolveLevel(ctx, enum.ExamTypeGrade, 3, &nine); got == nil || *got != 9 {
		t.Errorf("GRADE at grade 3 with level 9 → %v, want 9", got)
	}
	if got := resolveLevel(ctx, enum.ExamTypeAssessment, 0, &nine); got == nil || *got != 9 {
		t.Errorf("ASSESSMENT records the stated level as is, got %v", got)
	}
	if got := resolveLevel(ctx, enum.ExamTypeGrade, 0, nil); got != nil {
		t.Errorf("no level stated → none, got %v", got)
	}
}
