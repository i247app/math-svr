package exam

import (
	"testing"

	"math-ai.com/math-ai/internal/application/dto/question"
	"math-ai.com/math-ai/internal/shared/enum"
)

func questions(n int) []question.Question {
	out := make([]question.Question, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, question.Question{QuestionNumber: i})
	}
	return out
}

func gradesOf(qs []question.Question) []int {
	out := make([]int, 0, len(qs))
	for _, q := range qs {
		if q.Grade == nil {
			out = append(out, -1)
			continue
		}
		out = append(out, *q.Grade)
	}
	return out
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNormalizeQuestionBands(t *testing.T) {
	tests := []struct {
		name     string
		examType enum.ExamType
		grade    int
		count    int
		want     []int
	}{
		{
			name:     "assessment lifts questions 3 and 6 one band",
			examType: enum.ExamTypeAssessment,
			grade:    1,
			count:    10,
			want:     []int{1, 1, 2, 1, 1, 2, 1, 1, 1, 1},
		},
		{
			name:     "assessment at the top band probes into the ceiling",
			examType: enum.ExamTypeAssessment,
			grade:    5,
			count:    6,
			want:     []int{5, 5, 6, 5, 5, 6},
		},
		{
			name:     "assessment too short to hold the second probe",
			examType: enum.ExamTypeAssessment,
			grade:    2,
			count:    4,
			want:     []int{2, 2, 3, 2},
		},
		{
			name:     "practice pins every question to the requested grade",
			examType: enum.ExamTypePractice,
			grade:    3,
			count:    5,
			want:     []int{3, 3, 3, 3, 3},
		},
		{
			name:     "exam does not probe either",
			examType: enum.ExamTypeExam,
			grade:    0,
			count:    3,
			want:     []int{0, 0, 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := NormalizeQuestionBands(questions(tc.count), tc.examType, tc.grade, nil)
			if !equal(gradesOf(got), tc.want) {
				t.Errorf("grades = %v, want %v", gradesOf(got), tc.want)
			}
		})
	}
}

// TestNormalizeQuestionGradesOverridesModel is the reason this function
// exists: whatever the model claimed, the server's stamp wins, and the
// disagreement is reported rather than swallowed.
func TestNormalizeQuestionBandsOverridesModel(t *testing.T) {
	qs := questions(10)
	wrong := 4
	qs[2].Grade = &wrong // question 3 — a probe slot, model says grade 4

	got, mismatches := NormalizeQuestionBands(qs, enum.ExamTypeAssessment, 1, nil)

	if *got[2].Grade != 2 {
		t.Errorf("question 3 grade = %d, want 2 (server stamp must win)", *got[2].Grade)
	}
	if len(mismatches) != 1 {
		t.Fatalf("got %d mismatches, want 1", len(mismatches))
	}
	if mismatches[0].QuestionNumber != 3 || *mismatches[0].ModelGrade != 4 || mismatches[0].Applied != 2 {
		t.Errorf("mismatch = %+v, want {3, 4, 2}", mismatches[0])
	}
}

// TestNormalizeQuestionGradesFallsBackToPosition covers a payload whose
// numbering the model got wrong. Every question must still be stamped —
// leaving Grade nil would hand the placement rule a hole.
func TestNormalizeQuestionBandsFallsBackToPosition(t *testing.T) {
	qs := []question.Question{
		{QuestionNumber: 0},
		{QuestionNumber: 0},
		{QuestionNumber: 0},
		{QuestionNumber: 99},
	}
	got, _ := NormalizeQuestionBands(qs, enum.ExamTypeAssessment, 1, nil)

	want := []int{1, 1, 2, 1} // position 3 is the probe slot
	if !equal(gradesOf(got), want) {
		t.Errorf("grades = %v, want %v", gradesOf(got), want)
	}
}

func TestNormalizeQuestionBandsEmpty(t *testing.T) {
	got, mismatches := NormalizeQuestionBands(nil, enum.ExamTypeAssessment, 1, nil)
	if got != nil || mismatches != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", got, mismatches)
	}
}
