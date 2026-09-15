package query_test

import (
	"context"
	"testing"

	query "math-ai.com/math-ai/internal/application/query/exam"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

type jqJourneyRepo struct {
	exam.IUserExamRepository
	row *exam.UserExam
}

func (r *jqJourneyRepo) FindByUserExamIdAndType(ctx context.Context, id int64, typ string) (*exam.UserExam, error) {
	if r.row != nil && r.row.UserExamId() == id && r.row.ReqExamType() == typ {
		return r.row, nil
	}
	return nil, nil
}

type jqAttemptRepo struct {
	exam.IUserAiExamRepository
	latest *exam.UserAiExam
}

func (r *jqAttemptRepo) FindLatestSubmittedByUserExamId(ctx context.Context, id int64) (*exam.UserAiExam, error) {
	return r.latest, nil
}

func (r *jqAttemptRepo) ListByUserAiExamIds(ctx context.Context, ids []int64) ([]*exam.UserAiExam, error) {
	return nil, nil
}

type jqAiExamRepo struct{ exam.IAiExamRepository }

func (jqAiExamRepo) ListByAiExamIds(ctx context.Context, ids []int64) ([]*exam.AiExam, error) {
	return nil, nil
}

type jqDetailRepo struct {
	exam.IUserExamDetailRepository
	byJourney map[string][]*exam.UserExamDetail // keyed by type
	byAttempt map[int64][]*exam.UserExamDetail
}

func (r *jqDetailRepo) ListByUserExamId(ctx context.Context, id int64, typ string) ([]*exam.UserExamDetail, error) {
	return r.byJourney[typ], nil
}

func (r *jqDetailRepo) ListByUserAiExamId(ctx context.Context, id int64) ([]*exam.UserExamDetail, error) {
	return r.byAttempt[id], nil
}

func answer(attempt int64, topic string, correct bool) *exam.UserExamDetail {
	d := exam.NewUserExamDetail()
	d.SetUserAiExamId(attempt)
	d.SetQuestionTopic(&topic)
	d.SetIsCorrect(correct)
	return d
}

// TestJourneyPreviewReadsTheLatestSitting: the preview is drawn from the
// journey's latest SUBMITTED sitting — here a PRACTICE one, whose answers
// live under the other row and never appear in the ASSESSMENT view — not
// from the log being displayed. That is what keeps the preview equal to
// the round a hand-out would draw.
func TestJourneyPreviewReadsTheLatestSitting(t *testing.T) {
	j := exam.NewUserExam()
	j.SetUserExamId(1)
	j.SetUserId(1)
	j.SetProfileId(11)
	j.SetReqExamType(string(enum.ExamTypeAssessment))

	latest := exam.NewUserAiExam()
	latest.SetUserAiExamId(20)
	latest.SetReqExamType(string(enum.ExamTypePractice))

	h := query.NewGetExamJourneyQueryHandler(
		&jqJourneyRepo{row: j},
		&jqAttemptRepo{latest: latest},
		jqAiExamRepo{},
		&jqDetailRepo{
			byJourney: map[string][]*exam.UserExamDetail{
				"ASSESSMENT": {answer(10, "đếm", false)}, // the displayed log says "đếm" is weak…
			},
			byAttempt: map[int64][]*exam.UserExamDetail{
				20: {answer(20, "đếm", true), answer(20, "phép trừ", false)}, // …but the latest sitting says "phép trừ"
			},
		},
	)

	got, err := h.Handle(context.Background(), query.GetExamJourneyQuery{UserExamID: 1, ExamType: "ASSESSMENT", UserID: 1, ProfileID: 11})
	if err != nil {
		t.Fatal(err)
	}
	if got.PracticeBase == nil || got.PracticeBase.UserAiExamId() != 20 {
		t.Fatalf("base = %v, want sitting 20", got.PracticeBase)
	}
	if b := got.PracticeBrief; b == nil || len(b.WeakTopics) != 1 || b.WeakTopics[0] != "phép trừ" {
		t.Fatalf("brief weak topics = %v, want [phép trừ] from the latest sitting", got.PracticeBrief)
	}
}

// TestJourneyPreviewAbsentWithoutASitting: nothing submitted yet → no
// preview, and the read still succeeds.
func TestJourneyPreviewAbsentWithoutASitting(t *testing.T) {
	j := exam.NewUserExam()
	j.SetUserExamId(1)
	j.SetUserId(1)
	j.SetProfileId(11)
	j.SetReqExamType(string(enum.ExamTypeAssessment))

	h := query.NewGetExamJourneyQueryHandler(&jqJourneyRepo{row: j}, &jqAttemptRepo{}, jqAiExamRepo{}, &jqDetailRepo{})
	got, err := h.Handle(context.Background(), query.GetExamJourneyQuery{UserExamID: 1, ExamType: "ASSESSMENT", UserID: 1, ProfileID: 11})
	if err != nil {
		t.Fatal(err)
	}
	if got.PracticeBase != nil || got.PracticeBrief != nil {
		t.Fatal("no submitted sitting → no preview")
	}
}
