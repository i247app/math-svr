package llmroute

import (
	"slices"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Code names one kind of LLM call whose provider can be routed. The catalog
// is code, the target is a ma_llm_routes row: a row means something only
// when code resolves it, so the API edits seeded rows and never creates one.
// A new kind = a constant here + its caller + a seed row.
type Code string

const (
	// CodeDefault is the fallback of every route with no provider set; with
	// none set either, env LLM_PROVIDER decides.
	CodeDefault Code = "DEFAULT"

	CodeExamAssessment Code = "EXAM_ASSESSMENT"
	CodeExamGrade      Code = "EXAM_GRADE"
	CodeExamPractice   Code = "EXAM_PRACTICE"
	CodeExamReview     Code = "EXAM_REVIEW"

	CodeExerciseGenerate Code = "EXERCISE_GENERATE"
	CodeExerciseGrade    Code = "EXERCISE_GRADE"
)

func (c Code) String() string {
	return string(c)
}

// codes is the catalog in display order; it matches the seed file.
var codes = []Code{
	CodeDefault,
	CodeExamAssessment, CodeExamGrade, CodeExamPractice, CodeExamReview,
	CodeExerciseGenerate, CodeExerciseGrade,
}

// Codes returns the catalog in display order (a copy).
func Codes() []Code { return append([]Code(nil), codes...) }

// IsKnown reports whether c is in the catalog.
func IsKnown(c Code) bool { return slices.Contains(codes, c) }

// ExamCode is the route of an exam generation of the given type, false for
// a type that has none.
func ExamCode(t enum.ExamType) (Code, bool) {
	switch t {
	case enum.ExamTypeAssessment:
		return CodeExamAssessment, true
	case enum.ExamTypeGrade:
		return CodeExamGrade, true
	case enum.ExamTypePractice:
		return CodeExamPractice, true
	}
	return "", false
}
