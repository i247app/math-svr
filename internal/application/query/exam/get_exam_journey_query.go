package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// GetExamJourneyQuery reads one journey in full. Ownership is proven on
// both ids, the same as for a single attempt.
type GetExamJourneyQuery struct {
	UserExamID int64
	// ExamType picks the row of the journey — ASSESSMENT or PRACTICE —
	// since both share UserExamID. Normalised by the caller.
	ExamType  string
	UserID    int64
	ProfileID int64
}

// ExamJourneyDetail is the journey review screen in one read: the
// journey's running totals, every sitting that fed them, the question
// sets those sittings were drawn from, and every question answered.
//
// Attempts are recovered from the detail log rather than from a column on
// the attempt row: an attempt joins a journey the moment it is SUBMITTED
// (that is when its details are written against the journey), so the set
// of distinct user_ai_exam_id values in the log IS the set of sittings
// that belong here. An attempt still IN_PROGRESS belongs to no journey
// yet, and correctly does not appear.
type ExamJourneyDetail struct {
	Journey  *exam.UserExam
	Attempts []*exam.UserAiExam
	AiExams  map[int64]*exam.AiExam
	Details  []*exam.UserExamDetail
}

type GetExamJourneyQueryHandler struct {
	journeyRepo exam.IUserExamRepository
	attemptRepo exam.IUserAiExamRepository
	aiExamRepo  exam.IAiExamRepository
	detailRepo  exam.IUserExamDetailRepository
}

func NewGetExamJourneyQueryHandler(
	journeyRepo exam.IUserExamRepository,
	attemptRepo exam.IUserAiExamRepository,
	aiExamRepo exam.IAiExamRepository,
	detailRepo exam.IUserExamDetailRepository,
) *GetExamJourneyQueryHandler {
	return &GetExamJourneyQueryHandler{
		journeyRepo: journeyRepo,
		attemptRepo: attemptRepo,
		aiExamRepo:  aiExamRepo,
		detailRepo:  detailRepo,
	}
}

func (h *GetExamJourneyQueryHandler) Handle(ctx context.Context, q GetExamJourneyQuery) (*ExamJourneyDetail, error) {
	journey, err := h.journeyRepo.FindByUserExamIdAndType(ctx, q.UserExamID, q.ExamType)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if journey == nil {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil, nil)
	}
	if journey.UserId() != q.UserID || journey.ProfileId() != q.ProfileID {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_OWNED, nil, nil)
	}

	details, err := h.detailRepo.ListByUserExamId(ctx, journey.UserExamId(), journey.ReqExamType())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	out := &ExamJourneyDetail{
		Journey: journey,
		Details: details,
		AiExams: map[int64]*exam.AiExam{},
	}
	if len(details) == 0 {
		return out, nil
	}

	attemptIDs := distinctInOrder(details, func(d *exam.UserExamDetail) int64 { return d.UserAiExamId() })
	attempts, err := h.attemptRepo.ListByUserAiExamIds(ctx, attemptIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	out.Attempts = attempts

	aiExamIDs := distinctInOrder(details, func(d *exam.UserExamDetail) int64 { return d.AiExamId() })
	aiExams, err := h.aiExamRepo.ListByAiExamIds(ctx, aiExamIDs)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	for _, e := range aiExams {
		out.AiExams[e.AiExamId()] = e
	}
	return out, nil
}

// distinctInOrder collects one id per distinct value, keeping first-seen
// order — which, over a log sorted by sitting, is chronological.
func distinctInOrder(details []*exam.UserExamDetail, key func(*exam.UserExamDetail) int64) []int64 {
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
