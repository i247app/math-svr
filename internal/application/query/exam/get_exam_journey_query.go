package query

import (
	"context"

	"math-ai.com/math-ai/internal/application/command/shared/practice"
	"math-ai.com/math-ai/internal/domain/bot"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// GetExamJourneyQuery reads one journey in full. Ownership is proven on
// both ids, the same as for a single attempt.
type GetExamJourneyQuery struct {
	EsessID int64
	// ExamType picks the row of the journey — ASSESSMENT or PRACTICE —
	// since both share EsessID. Normalised by the caller.
	ExamType  string
	UID       int64
	ProfileID int64
}

// ExamJourneyDetail is the journey review screen in one read: the
// journey's running totals, every sitting that fed them, the question
// sets those sittings were drawn from, and every question answered.
//
// Attempts are recovered from the detail log rather than from a column on
// the attempt row: an attempt joins a journey the moment it is SUBMITTED
// (that is when its details are written against the journey), so the set
// of distinct elink_id values in the log IS the set of sittings
// that belong here. An attempt still IN_PROGRESS belongs to no journey
// yet, and correctly does not appear.
type ExamJourneyDetail struct {
	Journey   *exam.ExamSession
	Attempts  []*exam.ExamLink
	ExamPools map[int64]*exam.ExamPool
	Details   []*exam.ExamSessionLine

	// PracticeBase is the sitting a PRACTICE round would be drawn from
	// right now — the journey's latest submitted one, of any type — and
	// PracticeBrief what that round would be aimed at. Both nil when the
	// journey has nothing submitted yet. They are computed by the same
	// function the hand-out uses, so what the screen previews is what the
	// paper will drill.
	PracticeBase  *exam.ExamLink
	PracticeBrief *bot.PracticeBrief
}

type GetExamJourneyQueryHandler struct {
	journeyRepo  exam.IExamSessionRepository
	attemptRepo  exam.IExamLinkRepository
	examPoolRepo exam.IExamPoolRepository
	detailRepo   exam.IExamSessionLineRepository
}

func NewGetExamJourneyQueryHandler(
	journeyRepo exam.IExamSessionRepository,
	attemptRepo exam.IExamLinkRepository,
	examPoolRepo exam.IExamPoolRepository,
	detailRepo exam.IExamSessionLineRepository,
) *GetExamJourneyQueryHandler {
	return &GetExamJourneyQueryHandler{
		journeyRepo:  journeyRepo,
		attemptRepo:  attemptRepo,
		examPoolRepo: examPoolRepo,
		detailRepo:   detailRepo,
	}
}

func (h *GetExamJourneyQueryHandler) Handle(ctx context.Context, q GetExamJourneyQuery) (*ExamJourneyDetail, error) {
	journey, err := h.journeyRepo.FindByEsessIdAndType(ctx, q.EsessID, q.ExamType)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if journey == nil {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil, nil)
	}
	if journey.Uid() != q.UID || journey.ProfileId() != q.ProfileID {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_OWNED, nil, nil)
	}

	details, err := h.detailRepo.ListByEsessId(ctx, journey.EsessId(), journey.ReqExamType())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	out := &ExamJourneyDetail{
		Journey:   journey,
		Details:   details,
		ExamPools: map[int64]*exam.ExamPool{},
	}
	if err := h.previewPractice(ctx, out); err != nil {
		return nil, err
	}
	if len(details) == 0 {
		return out, nil
	}

	attemptIDs := distinctInOrder(details, func(d *exam.ExamSessionLine) int64 { return d.ElinkId() })
	attempts, err := h.attemptRepo.ListByElinkIds(ctx, attemptIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	out.Attempts = attempts

	examIDs := distinctInOrder(details, func(d *exam.ExamSessionLine) int64 { return d.ExamId() })
	examPools, err := h.examPoolRepo.ListByExamIds(ctx, examIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	for _, e := range examPools {
		out.ExamPools[e.ExamId()] = e
	}
	return out, nil
}

// previewPractice reads the sitting a practice round would be based on
// and derives its brief. The base is looked up on its own rather than
// picked out of the journey's log: the latest submitted sitting may be a
// PRACTICE one, and those live under the journey's other row.
func (h *GetExamJourneyQueryHandler) previewPractice(ctx context.Context, out *ExamJourneyDetail) error {
	base, err := h.attemptRepo.FindLatestSubmittedByEsessId(ctx, out.Journey.EsessId())
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	if base == nil {
		return nil
	}
	answers, err := h.detailRepo.ListByElinkId(ctx, base.ElinkId())
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	brief := practice.BuildBrief(answers)
	out.PracticeBase = base
	out.PracticeBrief = &brief
	return nil
}

// distinctInOrder collects one id per distinct value, keeping first-seen
// order — which, over a log sorted by sitting, is chronological.
func distinctInOrder(details []*exam.ExamSessionLine, key func(*exam.ExamSessionLine) int64) []int64 {
	seen := make(map[int64]struct{}, len(details))
	out := make([]int64, 0, len(details))
	for _, d := range details {
		k := key(d)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}
