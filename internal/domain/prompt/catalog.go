package prompt

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Code names one kind of LLM call. The catalog lives here, in code; the
// text lives in ma_prompts. A row only means something when code reads it,
// so admins rewrite the text of existing rows and never create a kind —
// adding one is a code change (a new reader) plus a seed row.
type Code string

const (
	// CodeExamAssessment / CodeExamGrade: generation of an ASSESSMENT /
	// GRADE exam, one prompt per grade.
	CodeExamAssessment Code = "EXAM_ASSESSMENT"
	CodeExamGrade      Code = "EXAM_GRADE"
	// CodeExamPractice: generation of a PRACTICE round — one prompt per
	// grade, like ASSESSMENT and GRADE; the user message adds the child's
	// mistakes.
	CodeExamPractice Code = "EXAM_PRACTICE"
	// CodeExamReview: the AI review of a finished journey.
	CodeExamReview Code = "EXAM_REVIEW"
)

// Spec is what code expects of a Code's rows: which dimensions pick a row,
// and which JSON keys the reply parser reads. RequiredKeys are quoted, the
// way the prompt's output structure names them; an update that drops one
// is refused, because every later call of that kind would be unreadable.
type Spec struct {
	ByGrade      bool
	ByLanguage   bool
	RequiredKeys []string
}

// examGenerationKeys are the keys aiquestion.ParseGenerationOf reads over
// dto.ExamQuestion / aiquestion.AnswerChoice.
var examGenerationKeys = []string{
	`"short_text"`, `"questions"`, `"question_number"`, `"question_kind"`, `"render_type"`,
	`"question_text"`, `"options"`, `"id"`, `"text"`,
	`"correct_answer"`, `"option_ids"`, `"right_answer_text"`, `"question_topic"`, `"question_grade"`,
}

var specs = map[Code]Spec{
	CodeExamAssessment: {ByGrade: true, RequiredKeys: examGenerationKeys},
	CodeExamGrade:      {ByGrade: true, RequiredKeys: examGenerationKeys},
	CodeExamPractice:   {ByGrade: true, RequiredKeys: examGenerationKeys},
	CodeExamReview:     {RequiredKeys: []string{`"short"`, `"long"`}},
}

// SpecOf returns the catalog entry of a code, false for a code no reader
// knows.
func SpecOf(c Code) (Spec, bool) {
	s, ok := specs[c]
	return s, ok
}

// Scope identifies one row: a code plus the dimensions its Spec uses. Build
// it with the helpers below so a dimension the code does not use is never
// set (the lookup is an exact match — there is no fallback row).
type Scope struct {
	Code     Code
	Grade    *int
	Language *string
}

// ExamGenerationScope is the row that writes an exam of this type at this
// grade.
func ExamGenerationScope(examType enum.ExamType, grade int) (Scope, error) {
	var code Code
	switch examType {
	case enum.ExamTypeAssessment:
		code = CodeExamAssessment
	case enum.ExamTypeGrade:
		code = CodeExamGrade
	case enum.ExamTypePractice:
		code = CodeExamPractice
	default:
		return Scope{}, fmt.Errorf("prompt: no generation prompt for exam type %q", examType)
	}
	s := Scope{Code: code}
	if specs[code].ByGrade {
		s.Grade = &grade
	}
	return s, nil
}

// JourneyReviewScope is the row that reviews a journey.
func JourneyReviewScope() Scope { return Scope{Code: CodeExamReview} }
