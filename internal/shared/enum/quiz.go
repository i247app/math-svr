package enum

// QuizPurpose names what the quiz is FOR: a diagnostic round, a daily
// practice, or a higher-stakes mock test. Persisted in
// `ma_quizzes.purpose` (renamed from the legacy `type` column).
type QuizPurpose string

const (
	QuizPurposeAssessment QuizPurpose = "ASSESSMENT"
	QuizPurposePractice   QuizPurpose = "PRACTICE"
	QuizPurposeExam       QuizPurpose = "EXAM"
)

func (p QuizPurpose) String() string {
	return string(p)
}

// IsValid accepts the three currently-modeled purposes. EXAM is listed
// but template support is still partial — callers that build prompts
// should branch when EXAM ships.
func (p QuizPurpose) IsValid() bool {
	switch p {
	case QuizPurposeAssessment, QuizPurposePractice, QuizPurposeExam:
		return true
	default:
		return false
	}
}

type AcademicPerformance string

const (
	AcademicPerformanceExcellentEN AcademicPerformance = "Excellent"
	AcademicPerformanceGoodEN      AcademicPerformance = "Good"
	AcademicPerformanceAverageEN   AcademicPerformance = "Average"
	AcademicPerformanceWeakEN      AcademicPerformance = "Weak"
	AcademicPerformanceVeryWeakEN  AcademicPerformance = "Very Weak"

	AcademicPerformanceExcellentVI AcademicPerformance = "Xuất sắc"
	AcademicPerformanceGoodVI      AcademicPerformance = "Giỏi"
	AcademicPerformanceAverageVI   AcademicPerformance = "Khá"
	AcademicPerformanceWeakVI      AcademicPerformance = "Trung bình"
	AcademicPerformanceVeryWeakVI  AcademicPerformance = "Yếu"
)

func DetectAcademicPerformanceByPercentage(percentage float64, language LanguageType) AcademicPerformance {
	switch language {
	case LanguageTypeEnglish:
		if percentage >= 90 {
			return AcademicPerformanceExcellentEN
		} else if percentage >= 70 {
			return AcademicPerformanceGoodEN
		} else if percentage >= 50 {
			return AcademicPerformanceAverageEN
		} else if percentage >= 30 {
			return AcademicPerformanceWeakEN
		} else {
			return AcademicPerformanceVeryWeakEN
		}
	case LanguageTypeVietnamese:
		if percentage >= 90 {
			return AcademicPerformanceExcellentVI
		} else if percentage >= 70 {
			return AcademicPerformanceGoodVI
		} else if percentage >= 50 {
			return AcademicPerformanceAverageVI
		} else if percentage >= 30 {
			return AcademicPerformanceWeakVI
		} else {
			return AcademicPerformanceVeryWeakVI
		}
	default:
		return AcademicPerformanceAverageEN
	}
}

// QuizTypeOfQuiz names the learning intent of the quiz: a GENERAL round
// introduces new knowledge from the curriculum, a REINFORCEMENT round
// consolidates / reviews material the student got wrong on a previous
// round. Persisted in `ma_quizzes.type_of_quiz`.
type QuizTypeOfQuiz string

const (
	// QuizTypeOfQuizGeneral is the default — a fresh round generated from
	// the curriculum context without a previous quiz to consolidate.
	QuizTypeOfQuizGeneral QuizTypeOfQuiz = "GENERAL"

	// QuizTypeOfQuizReinforcement consolidates previously-incorrect
	// questions. Generation always reads from a previous_quiz_id row.
	QuizTypeOfQuizReinforcement QuizTypeOfQuiz = "REINFORCEMENT"
)

func (t QuizTypeOfQuiz) String() string {
	return string(t)
}

func (t QuizTypeOfQuiz) IsValid() bool {
	switch t {
	case QuizTypeOfQuizGeneral, QuizTypeOfQuizReinforcement:
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
	// the bot has graded. review / assessment_grade are populated.
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
