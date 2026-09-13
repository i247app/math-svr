package command_test

import (
	"context"
	"testing"

	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// A two-question set in the exam vocabulary, as ma_ai_exams stores it.
// Key is A on both, so an all-A submission is 2/2 and an all-B one is 0/2.
const submitSet = `[
{"question_number":1,"question_name":"1+1","answers":[{"label":"A","content":"2"},{"label":"B","content":"3"}],"right_answer_label":"A","right_answer_content":"2","question_topic":"addition","question_grade":1},
{"question_number":2,"question_name":"2+2","answers":[{"label":"A","content":"4"},{"label":"B","content":"5"}],"right_answer_label":"A","right_answer_content":"4","question_topic":"addition","question_grade":1}]`

const (
	subUser    = int64(1)
	subProfile = int64(11)
	subAiExam  = int64(500)
	subAttempt = int64(700)
	subJourney = int64(9001)
	subNextSeq = int64(42)
)

// ---- fakes -----------------------------------------------------------------
//
// Each embeds its interface (nil) so only the methods the submit path
// exercises have bodies; anything else panics rather than no-oping.

type memAiExamRepo struct {
	exam.IAiExamRepository
	rows map[int64]*exam.AiExam
}

func (r *memAiExamRepo) FindByAiExamId(ctx context.Context, id int64) (*exam.AiExam, error) {
	return r.rows[id], nil
}

type memAttemptRepo struct {
	exam.IUserAiExamRepository
	rows       map[int64]*exam.UserAiExam
	submitted  []exam.AttemptResult
	loseSubmit bool // another submit already flipped the row
}

func (r *memAttemptRepo) FindByUserAiExamId(ctx context.Context, id int64) (*exam.UserAiExam, error) {
	return r.rows[id], nil
}

func (r *memAttemptRepo) MarkSubmitted(ctx context.Context, id int64, res exam.AttemptResult) error {
	if r.loseSubmit {
		return exam.ErrAttemptNotInProgress
	}
	r.submitted = append(r.submitted, res)
	a := r.rows[id]
	a.SetUserExamId(res.UserExamId)
	st := string(enum.UserAiExamStatusSubmitted)
	a.SetUserAiExamStatus(&st)
	return nil
}

// journeyKey is the row key the real table enforces with uk_journey_type.
type journeyKey struct {
	id  int64
	typ string
}

type memJourneyRepo struct {
	exam.IUserExamRepository
	rows        map[journeyKey]*exam.UserExam
	creates     []*exam.UserExam
	accumulates []journeyKey
	// conflictOnce makes the first Create collide and plant the row as if
	// another transaction had won — the race the command must survive.
	conflictOnce bool
}

func (r *memJourneyRepo) FindByUserExamIdAndType(ctx context.Context, id int64, typ string) (*exam.UserExam, error) {
	return r.rows[journeyKey{id, typ}], nil
}

func (r *memJourneyRepo) FindActiveByUserProfileType(ctx context.Context, userID, profileID int64, typ string) (*exam.UserExam, error) {
	for k, row := range r.rows {
		if k.typ == typ && row.UserId() == userID && row.ProfileId() == profileID && isActive(row) {
			return row, nil
		}
	}
	return nil, nil
}

func (r *memJourneyRepo) Create(ctx context.Context, e *exam.UserExam, delta exam.StatsDelta) error {
	k := journeyKey{e.UserExamId(), e.ReqExamType()}
	if r.conflictOnce {
		r.conflictOnce = false
		winner := journeyRow(e.UserExamId(), e.ReqExamType(), enum.UserExamStatusActive)
		winner.SetResTotalQuestions(5)
		winner.SetResCorrectNumber(5)
		r.rows[k] = winner
		return exam.ErrJourneyConflict
	}
	if _, dup := r.rows[k]; dup {
		return exam.ErrJourneyConflict
	}
	e.SetResTotalQuestions(delta.TotalQuestions)
	e.SetResCorrectNumber(delta.CorrectNumber)
	e.SetResSkippedNumber(delta.SkippedNumber)
	active := string(enum.UserExamStatusActive)
	e.SetUserExamStatus(&active)
	r.rows[k] = e
	r.creates = append(r.creates, e)
	return nil
}

func (r *memJourneyRepo) Accumulate(ctx context.Context, id int64, typ string, e *exam.UserExam, delta exam.StatsDelta) error {
	k := journeyKey{id, typ}
	row, ok := r.rows[k]
	if !ok || !isActive(row) {
		return exam.ErrJourneyNotActive
	}
	row.SetResTotalQuestions(row.ResTotalQuestions() + delta.TotalQuestions)
	row.SetResCorrectNumber(row.ResCorrectNumber() + delta.CorrectNumber)
	row.SetResSkippedNumber(row.ResSkippedNumber() + delta.SkippedNumber)
	row.SetResReview(e.ResReview())
	row.SetResGrade(e.ResGrade())
	r.accumulates = append(r.accumulates, k)
	return nil
}

type memDetailRepo struct {
	exam.IUserExamDetailRepository
	written []*exam.UserExamDetail
}

func (r *memDetailRepo) CreateBatch(ctx context.Context, details []*exam.UserExamDetail) error {
	r.written = append(r.written, details...)
	return nil
}

type countingSeq struct {
	seq.IRepository
	calls int
	next  int64
}

func (s *countingSeq) Next(ctx context.Context, name string) (int64, error) {
	s.calls++
	s.next++
	return s.next, nil
}

// ---- fixtures ---------------------------------------------------------------

func isActive(row *exam.UserExam) bool {
	s := row.UserExamStatus()
	return s != nil && *s == string(enum.UserExamStatusActive)
}

func journeyRow(id int64, typ string, st enum.UserExamStatusType) *exam.UserExam {
	j := exam.NewUserExam()
	j.SetUserExamId(id)
	j.SetUserId(subUser)
	j.SetProfileId(subProfile)
	j.SetReqExamType(typ)
	s := string(st)
	j.SetUserExamStatus(&s)
	return j
}

func openAttempt(typ enum.ExamType, journey *int64) *exam.UserAiExam {
	a := exam.NewUserAiExam()
	a.SetUserAiExamId(subAttempt)
	a.SetUserId(subUser)
	a.SetProfileId(subProfile)
	a.SetAiExamId(subAiExam)
	a.SetUserExamId(journey)
	a.SetReqExamType(string(typ))
	a.SetReqGrade(1)
	a.SetStartedDt(mtime.Now())
	st := string(enum.UserAiExamStatusInProgress)
	a.SetUserAiExamStatus(&st)
	return a
}

type harness struct {
	attempts *memAttemptRepo
	journeys *memJourneyRepo
	details  *memDetailRepo
	seq      *countingSeq
	handler  *command.SubmitExamCommandHandler
}

func newHarness(attempt *exam.UserAiExam, journeys ...*exam.UserExam) *harness {
	set := exam.NewAiExam()
	set.SetAiExamId(subAiExam)
	set.SetAiQuestionsJson(submitSet)

	h := &harness{
		attempts: &memAttemptRepo{rows: map[int64]*exam.UserAiExam{subAttempt: attempt}},
		journeys: &memJourneyRepo{rows: map[journeyKey]*exam.UserExam{}},
		details:  &memDetailRepo{},
		seq:      &countingSeq{next: subNextSeq - 1},
	}
	for _, j := range journeys {
		h.journeys.rows[journeyKey{j.UserExamId(), j.ReqExamType()}] = j
	}
	h.handler = command.NewSubmitExamCommandHandler(fakeUoW{repos: transaction.Repositories{
		Seq:            h.seq,
		AiExam:         &memAiExamRepo{rows: map[int64]*exam.AiExam{subAiExam: set}},
		UserAiExam:     h.attempts,
		UserExam:       h.journeys,
		UserExamDetail: h.details,
	}})
	return h
}

func submitAll(label string) command.SubmitExamCommand {
	return command.SubmitExamCommand{
		UserAiExamID: subAttempt, UserID: subUser, ProfileID: subProfile,
		Language: enum.LanguageTypeVietnamese,
		Answers: []question.StudentAnswer{
			{QuestionNumber: 1, Label: label},
			{QuestionNumber: 2, Label: label},
		},
	}
}

func ptr(v int64) *int64 { return &v }

// ---- tests -----------------------------------------------------------------

// TestSubmitAssessmentOpensJourney: the first ASSESSMENT sitting mints a
// journey id, opens the row, pins the attempt to it, and logs every
// answer under the ASSESSMENT type.
func TestSubmitAssessmentOpensJourney(t *testing.T) {
	h := newHarness(openAttempt(enum.ExamTypeAssessment, nil))

	res, err := h.handler.Handle(context.Background(), submitAll("A"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	if h.seq.calls == 0 {
		t.Fatal("an ASSESSMENT journey must mint its id from ma_seqs")
	}
	if res.Stats.UserExamId() != subNextSeq || res.Stats.ReqExamType() != string(enum.ExamTypeAssessment) {
		t.Fatalf("stats row = (%d, %s), want (%d, ASSESSMENT)", res.Stats.UserExamId(), res.Stats.ReqExamType(), subNextSeq)
	}
	if res.Stats.ResGrade() == nil {
		t.Fatal("an ASSESSMENT row must carry a measured grade")
	}
	if got := h.attempts.submitted[0].UserExamId; got == nil || *got != subNextSeq {
		t.Fatalf("attempt pinned to journey %v, want %d", got, subNextSeq)
	}
	for _, d := range h.details.written {
		if d.UserExamId() != subNextSeq || d.ReqExamType() != string(enum.ExamTypeAssessment) {
			t.Fatalf("detail row logged under (%d, %s)", d.UserExamId(), d.ReqExamType())
		}
	}
}

// TestSubmitPracticeOpensRowUnderJourneyId: the first PRACTICE sitting of
// a journey opens the PRACTICE row under the journey's OWN id — no
// sequence drawn — with no grade, and leaves the ASSESSMENT row's totals
// untouched.
func TestSubmitPracticeOpensRowUnderJourneyId(t *testing.T) {
	owner := journeyRow(subJourney, string(enum.ExamTypeAssessment), enum.UserExamStatusActive)
	owner.SetResTotalQuestions(10)
	owner.SetResCorrectNumber(7)
	h := newHarness(openAttempt(enum.ExamTypePractice, ptr(subJourney)), owner)

	res, err := h.handler.Handle(context.Background(), submitAll("B"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	if h.seq.calls != 2 { // one per detail row; none for the journey
		t.Fatalf("seq drawn %d times; a PRACTICE row reuses the journey id and only the 2 details mint", h.seq.calls)
	}
	if res.Stats.UserExamId() != subJourney || res.Stats.ReqExamType() != string(enum.ExamTypePractice) {
		t.Fatalf("stats row = (%d, %s), want (%d, PRACTICE)", res.Stats.UserExamId(), res.Stats.ReqExamType(), subJourney)
	}
	if res.Stats.ResGrade() != nil {
		t.Fatalf("PRACTICE row must never carry a grade, got %d", *res.Stats.ResGrade())
	}
	if res.Stats.ResTotalQuestions() != 2 || res.Stats.ResCorrectNumber() != 0 {
		t.Fatalf("PRACTICE totals = %d/%d, want 0/2", res.Stats.ResCorrectNumber(), res.Stats.ResTotalQuestions())
	}
	if owner.ResTotalQuestions() != 10 || owner.ResCorrectNumber() != 7 {
		t.Fatalf("ASSESSMENT row moved to %d/%d; practice must not touch it", owner.ResCorrectNumber(), owner.ResTotalQuestions())
	}
	if got := h.attempts.submitted[0].UserExamId; got == nil || *got != subJourney {
		t.Fatalf("attempt pinned to journey %v, want %d", got, subJourney)
	}
	for _, d := range h.details.written {
		if d.UserExamId() != subJourney || d.ReqExamType() != string(enum.ExamTypePractice) {
			t.Fatalf("detail row logged under (%d, %s)", d.UserExamId(), d.ReqExamType())
		}
	}
}

// TestSubmitPracticeAccumulates: a second PRACTICE sitting folds into the
// existing PRACTICE row.
func TestSubmitPracticeAccumulates(t *testing.T) {
	owner := journeyRow(subJourney, string(enum.ExamTypeAssessment), enum.UserExamStatusActive)
	practice := journeyRow(subJourney, string(enum.ExamTypePractice), enum.UserExamStatusActive)
	practice.SetResTotalQuestions(4)
	practice.SetResCorrectNumber(1)
	h := newHarness(openAttempt(enum.ExamTypePractice, ptr(subJourney)), owner, practice)

	res, err := h.handler.Handle(context.Background(), submitAll("A"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if len(h.journeys.creates) != 0 {
		t.Fatal("must accumulate into the existing PRACTICE row, not open another")
	}
	if res.Stats.ResTotalQuestions() != 6 || res.Stats.ResCorrectNumber() != 3 {
		t.Fatalf("PRACTICE totals = %d/%d, want 3/6", res.Stats.ResCorrectNumber(), res.Stats.ResTotalQuestions())
	}
	if res.Stats.ResGrade() != nil {
		t.Fatal("PRACTICE row must stay ungraded after accumulating")
	}
}

// TestSubmitPracticeSurvivesOpenRace: two first-ever practice submits race
// to open the row; the loser's INSERT collides, it re-reads, and folds
// into the winner's row instead of failing.
func TestSubmitPracticeSurvivesOpenRace(t *testing.T) {
	owner := journeyRow(subJourney, string(enum.ExamTypeAssessment), enum.UserExamStatusActive)
	h := newHarness(openAttempt(enum.ExamTypePractice, ptr(subJourney)), owner)
	h.journeys.conflictOnce = true

	res, err := h.handler.Handle(context.Background(), submitAll("A"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	want := journeyKey{subJourney, string(enum.ExamTypePractice)}
	if len(h.journeys.accumulates) != 1 || h.journeys.accumulates[0] != want {
		t.Fatalf("accumulates = %v, want exactly %v", h.journeys.accumulates, want)
	}
	if res.Stats.ResTotalQuestions() != 7 { // winner's 5 + ours 2
		t.Fatalf("totals after the race = %d, want 7", res.Stats.ResTotalQuestions())
	}
}

// TestSubmitPracticeRefusals: the cases a PRACTICE sitting must be turned
// away from, each without touching the log.
func TestSubmitPracticeRefusals(t *testing.T) {
	ended := journeyRow(subJourney, string(enum.ExamTypeAssessment), enum.UserExamStatusComplete)
	other := journeyRow(subJourney, string(enum.ExamTypeAssessment), enum.UserExamStatusActive)
	other.SetProfileId(subProfile + 1)

	tests := []struct {
		name    string
		attempt *exam.UserAiExam
		seed    []*exam.UserExam
		want    status.StatusCode
	}{
		{"journey ended after hand-out", openAttempt(enum.ExamTypePractice, ptr(subJourney)), []*exam.UserExam{ended}, status.EXAM_JOURNEY_ALREADY_ENDED},
		{"journey missing", openAttempt(enum.ExamTypePractice, ptr(subJourney)), nil, status.EXAM_JOURNEY_NOT_FOUND},
		{"journey of another profile", openAttempt(enum.ExamTypePractice, ptr(subJourney)), []*exam.UserExam{other}, status.EXAM_JOURNEY_NOT_OWNED},
		{"attempt drawn without a journey", openAttempt(enum.ExamTypePractice, nil), nil, status.EXAM_JOURNEY_NOT_FOUND},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(tc.attempt, tc.seed...)
			_, err := h.handler.Handle(context.Background(), submitAll("A"))
			if got := codeOf(t, err); got != tc.want {
				t.Fatalf("code = %d, want %d", got, tc.want)
			}
			if len(h.details.written) != 0 || len(h.attempts.submitted) != 0 {
				t.Fatal("a refused submit must write nothing")
			}
		})
	}
}

// TestSubmitLosesAttemptRace: the fold happens first, then the attempt's
// state guard fires for the loser of a double submit — the whole thing
// rolls back, so the response is ALREADY_SUBMITTED and nothing is logged.
func TestSubmitLosesAttemptRace(t *testing.T) {
	h := newHarness(openAttempt(enum.ExamTypeAssessment, nil))
	h.attempts.loseSubmit = true

	_, err := h.handler.Handle(context.Background(), submitAll("A"))
	if got := codeOf(t, err); got != status.EXAM_ALREADY_SUBMITTED {
		t.Fatalf("code = %d, want EXAM_ALREADY_SUBMITTED", got)
	}
	if len(h.details.written) != 0 {
		t.Fatal("details must not be written after the attempt guard fires")
	}
}
