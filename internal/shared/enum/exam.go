package enum

// ExamType names what a round of AI-generated questions is FOR. It is the
// only discriminator the exam model keeps — the old GENERAL /
// REINFORCEMENT axis was dropped along with the reinforcement flow.
type ExamType string

const (
	// ExamTypeAssessment measures where the child actually is. Its
	// generated set carries probe questions one grade above the requested
	// one (see the Q3/Q6 rule), so a child who is ready to move up shows it.
	ExamTypeAssessment ExamType = "ASSESSMENT"
	// ExamTypePractice is the everyday round at the child's current band.
	ExamTypePractice ExamType = "PRACTICE"
	// ExamTypeExam is the higher-stakes checkpoint round.
	ExamTypeExam ExamType = "EXAM"
)

func (t ExamType) String() string { return string(t) }

func (t ExamType) IsValid() bool {
	switch t {
	case ExamTypeAssessment, ExamTypePractice, ExamTypeExam:
		return true
	default:
		return false
	}
}

// UserAiExamStatusType is the lifecycle of ONE attempt at an exam.
//
// An attempt is created IN_PROGRESS the moment the exam is handed out and
// only ever leaves that state by being submitted. An attempt the child
// abandons simply stays IN_PROGRESS — that is what lets the app show
// "you still have this one open" and render it as 0/N.
type UserAiExamStatusType string

const (
	UserAiExamStatusInProgress UserAiExamStatusType = "IN_PROGRESS"
	UserAiExamStatusSubmitted  UserAiExamStatusType = "SUBMITTED"
	UserAiExamStatusDeleted    UserAiExamStatusType = "DELETED"
)

func (s UserAiExamStatusType) String() string { return string(s) }

func (s UserAiExamStatusType) IsValid() bool {
	switch s {
	case UserAiExamStatusInProgress, UserAiExamStatusSubmitted, UserAiExamStatusDeleted:
		return true
	default:
		return false
	}
}

// AiExamStatusType is the business lifecycle of a generated question set.
// A DELETED row is excluded from cache lookups as well as from reads.
type AiExamStatusType string

const (
	AiExamStatusActive  AiExamStatusType = "ACTIVE"
	AiExamStatusDeleted AiExamStatusType = "DELETED"
)

func (s AiExamStatusType) String() string { return string(s) }

// UserExamStatusType is the business lifecycle of a lifetime statistics
// row. There is no ARCHIVED state: totals are either live or deleted.
type UserExamStatusType string

const (
	UserExamStatusActive  UserExamStatusType = "ACTIVE"
	UserExamStatusDeleted UserExamStatusType = "DELETED"
)

func (s UserExamStatusType) String() string { return string(s) }

// UserExamDetailStatusType is the business lifecycle of one answered
// question. Rows are written once at submit and never edited, so SUBMITTED
// is the only state they normally hold.
type UserExamDetailStatusType string

const (
	UserExamDetailStatusSubmitted UserExamDetailStatusType = "SUBMITTED"
	UserExamDetailStatusDeleted   UserExamDetailStatusType = "DELETED"
)

func (s UserExamDetailStatusType) String() string { return string(s) }

// Bounds for the two difficulty axes.
//
// Grade is the CONTENT band and maps 1:1 onto the bot's grade profiles:
// 0 is kindergarten (mẫu giáo), 5 is the last elementary year. Level is
// the INTENSITY within whatever content the grade allows.
//
// ExamQuestionGradeMax is one above ExamGradeMax on purpose: an
// ASSESSMENT at grade 5 still has to probe upward, so its probe questions
// carry grade 6 even though no child can be PLACED at 6.
const (
	ExamGradeMin         = 0
	ExamGradeMax         = 5
	ExamLevelMin         = 1
	ExamLevelMax         = 10
	ExamQuestionGradeMax = ExamGradeMax + 1
)
