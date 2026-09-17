package exam

import (
	"encoding/json"
	"testing"
)

// TestExamQuestionGradeDecodes: question_grade arrives as whatever the
// prompt of the day asked for — an int, a Vietnamese band label, an
// English one — and must never fail the round. Stored rows always carry
// the int, so that path must keep round-tripping.
func TestExamQuestionGradeDecodes(t *testing.T) {
	tests := []struct {
		raw  string
		want *int
	}{
		{`2`, ptr(2)},
		{`0`, ptr(0)},
		{`"Mẫu giáo"`, ptr(0)},
		{`"mầm non"`, ptr(0)},
		{`"Kindergarten"`, ptr(0)},
		{`"Lớp 3"`, ptr(3)},
		{`"lớp 6"`, ptr(6)},
		{`"Grade 1"`, ptr(1)},
		{`"khó"`, nil},
		{`null`, nil},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			var q ExamQuestion
			if err := json.Unmarshal([]byte(`{"question_number":1,"question_name":"1+1=?","question_grade":`+tc.raw+`}`), &q); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			switch {
			case tc.want == nil && q.QuestionGrade != nil:
				t.Errorf("question_grade %s decoded to %d, want nil", tc.raw, *q.QuestionGrade)
			case tc.want != nil && (q.QuestionGrade == nil || *q.QuestionGrade != *tc.want):
				t.Errorf("question_grade %s decoded to %v, want %d", tc.raw, q.QuestionGrade, *tc.want)
			}
			if q.QuestionNumber != 1 || q.QuestionName != "1+1=?" {
				t.Error("the other fields must still decode")
			}
		})
	}

	// Absent stays absent.
	var q ExamQuestion
	if err := json.Unmarshal([]byte(`{"question_number":1}`), &q); err != nil || q.QuestionGrade != nil {
		t.Errorf("absent question_grade: err=%v grade=%v", err, q.QuestionGrade)
	}

	// Round trip through the stored shape.
	out, _ := json.Marshal(ExamQuestion{QuestionNumber: 1, QuestionGrade: ptr(4)})
	var back ExamQuestion
	if err := json.Unmarshal(out, &back); err != nil || back.QuestionGrade == nil || *back.QuestionGrade != 4 {
		t.Errorf("round trip: err=%v grade=%v", err, back.QuestionGrade)
	}
}

func ptr(n int) *int { return &n }
