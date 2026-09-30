package enum

// ExamType names what a round of AI-generated questions is FOR. It is the
// only discriminator the exam model keeps — the old GENERAL /
// REINFORCEMENT axis was dropped along with the reinforcement flow, and
// the EXAM checkpoint round was dropped when PRACTICE moved inside the
// journeys (ASSESSMENT and GRADE).
type ExamType string

const (
	// ExamTypeAssessment measures where the child actually is. Its
	// generated set carries probe questions one grade above the requested
	// one (see the Q3/Q6 rule), so a child who is ready to move up shows it.
	// An ASSESSMENT journey is the unit of lifecycle: it opens, accumulates
	// sittings, and is ended by the parent.
	ExamTypeAssessment ExamType = "ASSESSMENT"
	// ExamTypePractice is a round drawn AFTER a journey — ASSESSMENT or
	// GRADE — has been COMPLETED, from the child's latest submitted sitting in it —
	// re-drilling what went wrong, or pushing further when nothing did.
	// It is not a journey of its own: it shares the journey's
	// esess_id and never moves the journey's grade. A journey that is
	// still open, or was cancelled, cannot be practised in; reopening a
	// completed one closes practice again until it is completed anew. It
	// carries the same probe questions an ASSESSMENT does (see HasProbes).
	ExamTypePractice ExamType = "PRACTICE"
	// ExamTypeGrade 1-5 is for review specific grade 1-5
	ExamTypeGrade ExamType = "GRADE"
)

func (t ExamType) String() string { return string(t) }

// HasProbes reports whether a round of this type carries the harder
// probe questions (Q3/Q6 at grade + 1). Every type does today: ASSESSMENT
// needs them to measure, and PRACTICE and GRADE keep them so any round
// stretches the child the same way. The method stays so the rule has one
// home if a flat type ever comes back.
func (t ExamType) HasProbes() bool {
	return t.IsValid()
}

func (t ExamType) IsValid() bool {
	switch t {
	case ExamTypeAssessment, ExamTypePractice, ExamTypeGrade:
		return true
	default:
		return false
	}
}

// PracticeMode is how a PRACTICE round is aimed, decided by the server
// from the child's latest submitted sitting in the journey:
//
//   - RETRY_WEAK when that sitting had wrong answers — the round stays at
//     the same grade and drills the topics that went wrong.
//   - ADVANCE when it had none — still the same grade (placement is the
//     ASSESSMENT's job), but harder within it and/or on topics the child
//     has not been tested on yet.
type PracticeMode string

const (
	PracticeModeRetryWeak PracticeMode = "RETRY_WEAK"
	PracticeModeAdvance   PracticeMode = "ADVANCE"
)

func (m PracticeMode) String() string { return string(m) }

// ElinkStatusType is the lifecycle of ONE attempt at an exam.
//
// An attempt is created IN_PROGRESS the moment the exam is handed out and
// only ever leaves that state by being submitted. An attempt the child
// abandons simply stays IN_PROGRESS — that is what lets the app show
// "you still have this one open" and render it as 0/N.
type ElinkStatusType string

const (
	ElinkStatusInProgress ElinkStatusType = "IN_PROGRESS"
	ElinkStatusSubmitted  ElinkStatusType = "SUBMITTED"
	ElinkStatusDeleted    ElinkStatusType = "DELETED"
)

func (s ElinkStatusType) String() string { return string(s) }

func (s ElinkStatusType) IsValid() bool {
	switch s {
	case ElinkStatusInProgress, ElinkStatusSubmitted, ElinkStatusDeleted:
		return true
	default:
		return false
	}
}

// ExamStatusType is the business lifecycle of a generated question set.
// A DELETED row is excluded from cache lookups as well as from reads.
type ExamStatusType string

const (
	ExamStatusActive  ExamStatusType = "ACTIVE"
	ExamStatusDeleted ExamStatusType = "DELETED"
)

func (s ExamStatusType) String() string { return string(s) }

// EsessStatusType is the lifecycle of one JOURNEY — a stretch of
// ASSESSMENT sittings (and the PRACTICE rounds drawn inside it) that a
// child works through and then closes.
//
// ACTIVE is the open journey; every submission folds into it — its own
// sittings into its owning row (ASSESSMENT or GRADE), PRACTICE sittings
// into the PRACTICE row that shares its id. Ending the journey ends both rows.
//
// An ended journey CAN be reopened — marked ACTIVE again — as long as no
// other journey holds its slot open (same type, and same grade for GRADE):
// a child may change their mind and pick a run back up. Reopening clears ended_dt, and from then on the
// journey behaves as if it had never ended.
//
// The PRACTICE row that shares a journey's id is not part of this
// lifecycle. It exists only once the journey is COMPLETE and carries that
// status for as long as it lives; ending and reopening touch the owning
// row alone.
// COMPLETE and CANCEL both end it: COMPLETE is a run the child finished
// (and the only state a PRACTICE round may be drawn on), CANCEL a run
// they abandoned. Neither hands anything to the next journey — a new one
// starts where the client, or the profile, says.
type EsessStatusType string

const (
	EsessStatusActive   EsessStatusType = "ACTIVE"
	EsessStatusComplete EsessStatusType = "COMPLETE"
	EsessStatusCancel   EsessStatusType = "CANCEL"
	EsessStatusDeleted  EsessStatusType = "DELETED"
)

// IsValid accepts every lifecycle value, including DELETED, for callers
// that filter on status.
func (s EsessStatusType) IsValid() bool {
	switch s {
	case EsessStatusActive, EsessStatusComplete, EsessStatusCancel, EsessStatusDeleted:
		return true
	default:
		return false
	}
}

// IsEnding reports whether a status finishes a journey.
func (s EsessStatusType) IsEnding() bool {
	return s == EsessStatusComplete || s == EsessStatusCancel
}

// IsActive reports whether a status finishes a journey.
func (s EsessStatusType) IsActive() bool {
	return s == EsessStatusActive
}

// IsMarkable reports whether a status is one a client may MARK a journey
// with: an ending, or ACTIVE to reopen an ended journey. DELETED is
// deliberately excluded: it is a soft-delete, not a lifecycle move, and
// reaches the row through a different path.
func (s EsessStatusType) IsMarkable() bool {
	return s.IsEnding() || s.IsActive()
}

func (s EsessStatusType) String() string { return string(s) }

// EsessLnStatusType is the business lifecycle of one answered
// question. Rows are written once at submit and never edited, so SUBMITTED
// is the only state they normally hold.
type EsessLnStatusType string

const (
	EsessLnStatusActive  EsessLnStatusType = "ACTIVE"
	EsessLnStatusDeleted EsessLnStatusType = "DELETED"
)

func (s EsessLnStatusType) String() string { return string(s) }

// Bounds for the grade axis.
//
// Grade is the CONTENT band and maps 1:1 onto the bot's grade profiles:
// 0 is kindergarten (mẫu giáo), 5 is the last elementary year.
//
// Level is a 1..10 scale the CLIENT states on a hand-out and the server
// records (ma_exam_sessions.current_level, ma_exam_links.req_level). No
// server rule reads it yet — the prompt does not see it and nothing
// derives it — so it is a recorded fact, not a behaviour.
//
// ExamQuestionGradeMax is one above ExamGradeMax on purpose: an
// ASSESSMENT at grade 5 still has to probe upward, so its probe questions
// carry grade 6 even though no child can be PLACED at 6.
const (
	ExamGradeMin         = 0
	ExamGradeMax         = 5
	ExamQuestionGradeMax = ExamGradeMax + 1

	ExamLevelMin = 1
	ExamLevelMax = 10
)
