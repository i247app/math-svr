package command_test

import (
	"context"
	"testing"

	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

type genAiExamRepo struct {
	exam.IAiExamRepository
	created []*exam.AiExam
}

func (r *genAiExamRepo) Create(ctx context.Context, e *exam.AiExam) (*exam.AiExam, error) {
	r.created = append(r.created, e)
	return e, nil
}

type genAttemptRepo struct {
	exam.IUserAiExamRepository
	created []*exam.UserAiExam
}

func (r *genAttemptRepo) Create(ctx context.Context, a *exam.UserAiExam) (*exam.UserAiExam, error) {
	r.created = append(r.created, a)
	return a, nil
}

type genHarness struct {
	journeys *memJourneyRepo
	attempts *genAttemptRepo
	seq      *countingSeq
	handler  *command.GenerateExamCommandHandler
}

func newGenHarness(journeys ...*exam.UserExam) *genHarness {
	h := &genHarness{
		journeys: &memJourneyRepo{rows: map[journeyKey]*exam.UserExam{}},
		attempts: &genAttemptRepo{},
		seq:      &countingSeq{next: subNextSeq - 1},
	}
	for _, j := range journeys {
		h.journeys.rows[journeyKey{j.UserExamId(), j.ReqExamType()}] = j
	}
	h.handler = command.NewGenerateExamCommandHandler(fakeUoW{repos: transaction.Repositories{
		Seq:        h.seq,
		AiExam:     &genAiExamRepo{},
		UserAiExam: h.attempts,
		UserExam:   h.journeys,
	}})
	return h
}

func generate(typ enum.ExamType, journey *int64) command.GenerateExamCommand {
	return command.GenerateExamCommand{
		UserID: subUser, ProfileID: subProfile, ExamType: typ, Grade: 1, UserExamID: journey,
		NewContent: &command.NewAiExamContent{NumQues: 2, QuestionsJSON: submitSet},
	}
}

// TestGenerateOpensJourneyAtHandOut: the first exam a child is handed
// opens their journey right there, empty, in the same transaction — so a
// child who walks away still has a journey the sitting shows up under.
func TestGenerateOpensJourneyAtHandOut(t *testing.T) {
	h := newGenHarness()

	res, err := h.handler.Handle(context.Background(), generate(enum.ExamTypeAssessment, nil))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if len(h.journeys.creates) != 1 {
		t.Fatalf("%d journeys opened, want 1", len(h.journeys.creates))
	}
	opened := h.journeys.creates[0]
	if opened.ReqExamType() != string(enum.ExamTypeAssessment) || opened.ResTotalQuestions() != 0 || opened.ResGrade() != nil {
		t.Fatalf("journey opened as (%s, %d answered, grade %v); want an empty ASSESSMENT row", opened.ReqExamType(), opened.ResTotalQuestions(), opened.ResGrade())
	}
	if got := res.Attempt.UserExamId(); got == nil || *got != opened.UserExamId() {
		t.Fatalf("attempt pinned to journey %v, want %d", got, opened.UserExamId())
	}
}

// TestGenerateJoinsTheOpenJourney: with a journey already open, hand-out
// pins the sitting to it and opens nothing.
func TestGenerateJoinsTheOpenJourney(t *testing.T) {
	open := journeyRow(subJourney, string(enum.ExamTypeAssessment), enum.UserExamStatusActive)
	h := newGenHarness(open)

	res, err := h.handler.Handle(context.Background(), generate(enum.ExamTypeAssessment, nil))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(h.journeys.creates) != 0 {
		t.Fatal("a second journey must not be opened while one is active")
	}
	if got := res.Attempt.UserExamId(); got == nil || *got != subJourney {
		t.Fatalf("attempt pinned to journey %v, want %d", got, subJourney)
	}
}

// TestGenerateSurvivesOpenRace: two hand-outs race to open the child's
// first journey; the loser's INSERT collides, it re-reads, and pins to
// the winner's row.
func TestGenerateSurvivesOpenRace(t *testing.T) {
	h := newGenHarness()
	h.journeys.conflictOnce = true

	res, err := h.handler.Handle(context.Background(), generate(enum.ExamTypeAssessment, nil))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if res.Attempt.UserExamId() == nil {
		t.Fatal("attempt must be pinned to the journey that won the race")
	}
	if len(h.journeys.creates) != 0 {
		t.Fatal("the loser must not end up owning a journey")
	}
}

// TestGeneratePracticeNeverOpens: a PRACTICE round names its journey; one
// that arrives without it is a caller bug, not a reason to open a journey.
func TestGeneratePracticeNeverOpens(t *testing.T) {
	h := newGenHarness()
	_, err := h.handler.Handle(context.Background(), generate(enum.ExamTypePractice, nil))
	if got := codeOf(t, err); got != status.EXAM_MISSING_JOURNEY_ID {
		t.Fatalf("code = %d, want EXAM_MISSING_JOURNEY_ID", got)
	}
	if len(h.journeys.creates) != 0 || len(h.attempts.created) != 0 {
		t.Fatal("nothing may be written")
	}
}
