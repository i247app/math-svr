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

// UserExamStatusType is the lifecycle of one JOURNEY — a stretch of one
// exam type that a child works through and then closes.
//
// ACTIVE is the open journey; every submission of that type folds into it.
// COMPLETE and CANCEL both end it. The difference is what the next journey
// inherits: a COMPLETE journey's measured grade carries forward as the
// starting point, a CANCEL journey is treated as abandoned and the next
// one starts from the profile again. Both are terminal — an ended journey
// is never reopened, a new row is opened instead.
type UserExamStatusType string

const (
	UserExamStatusActive   UserExamStatusType = "ACTIVE"
	UserExamStatusComplete UserExamStatusType = "COMPLETE"
	UserExamStatusCancel   UserExamStatusType = "CANCEL"
	UserExamStatusDeleted  UserExamStatusType = "DELETED"
)

// IsValid accepts every lifecycle value, including DELETED, for callers
// that filter on status.
func (s UserExamStatusType) IsValid() bool {
	switch s {
	case UserExamStatusActive, UserExamStatusComplete, UserExamStatusCancel, UserExamStatusDeleted:
		return true
	default:
		return false
	}
}

// IsEnding reports whether a status is one a client may MARK a journey
// with. DELETED is deliberately excluded: it is a soft-delete, not a way
// to finish a journey, and reaches the row through a different path.
func (s UserExamStatusType) IsEnding() bool {
	return s == UserExamStatusComplete || s == UserExamStatusCancel
}

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

// Bounds for the grade axis.
//
// Grade is the CONTENT band and maps 1:1 onto the bot's grade profiles:
// 0 is kindergarten (mẫu giáo), 5 is the last elementary year.
//
// There is no level axis in code. The req_level / res_level /
// question_level columns exist and stay NULL: the teaching team has not
// defined what a level is or how it moves, so nothing accepts, derives or
// prompts for one. Reintroduce the bounds here when that rule lands.
//
// ExamQuestionGradeMax is one above ExamGradeMax on purpose: an
// ASSESSMENT at grade 5 still has to probe upward, so its probe questions
// carry grade 6 even though no child can be PLACED at 6.
const (
	ExamGradeMin         = 0
	ExamGradeMax         = 5
	ExamQuestionGradeMax = ExamGradeMax + 1
)
