package enum

type QuizType string

const (
	QuizTypeAssessment QuizType = "ASSESSMENT"
	QuizTypePractice   QuizType = "PRACTICE"
	QuizTypeExam       QuizType = "EXAM"
)

func (t QuizType) String() string {
	return string(t)
}

// IsValid rejects EXAM until templates exist for it; ValidateExam-Aware
// callers should branch separately when EXAM ships.
func (t QuizType) IsValid() bool {
	switch t {
	case QuizTypeAssessment, QuizTypePractice:
		return true
	default:
		return false
	}
}

type QuizStatusType string

const (
	// QuizStatusTypeGenerated marks a quiz that has questions but no
	// answers yet. Listing endpoints should still expose these so
	// students can resume an in-flight quiz.
	QuizStatusTypeGenerated QuizStatusType = "GENERATED"

	// QuizStatusTypeSubmitted marks a quiz the student has answered and
	// the bot has graded. ai_review / ai_detect_grade are populated.
	QuizStatusTypeSubmitted QuizStatusType = "SUBMITTED"

	// QuizStatusTypeDeleted is the soft-delete marker. Filtered out by
	// reads via the standard active-where clause.
	QuizStatusTypeDeleted QuizStatusType = "DELETED"
)

func (s QuizStatusType) String() string {
	return string(s)
}

func (s QuizStatusType) IsValid() bool {
	switch s {
	case QuizStatusTypeGenerated, QuizStatusTypeSubmitted:
		return true
	default:
		return false
	}
}
