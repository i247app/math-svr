package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListExamAttemptsQuery reads one child's history. Status narrows it to
// the unfinished ones, which is how the "you left this open" surface is
// built — the same list, filtered, rather than a second endpoint.
type ListExamAttemptsQuery struct {
	ProfileID int64
	ExamType  *string
	EsessID   *int64
	Status    *string
	Page      int64
	Limit     int64
}

// ListExamAttemptsResult pairs the attempts with the question sets they
// were drawn from, keyed by exam_id. A card needs the title, which
// lives on the shared row, not on the attempt.
type ListExamAttemptsResult struct {
	Attempts   []*exam.ExamLink
	ExamPools  map[int64]*exam.ExamPool
	Pagination *pagination.Pagination
}

type ListExamAttemptsQueryHandler struct {
	attemptRepo  exam.IExamLinkRepository
	examPoolRepo exam.IExamPoolRepository
}

func NewListExamAttemptsQueryHandler(
	attemptRepo exam.IExamLinkRepository,
	examPoolRepo exam.IExamPoolRepository,
) *ListExamAttemptsQueryHandler {
	return &ListExamAttemptsQueryHandler{attemptRepo: attemptRepo, examPoolRepo: examPoolRepo}
}

func (h *ListExamAttemptsQueryHandler) Handle(ctx context.Context, q ListExamAttemptsQuery) (*ListExamAttemptsResult, error) {
	attempts, pg, err := h.attemptRepo.ListAttempts(ctx, exam.ListAttemptsFilter{
		ProfileID: q.ProfileID,
		ExamType:  q.ExamType,
		EsessID:   q.EsessID,
		Status:    q.Status,
	}, q.Page, q.Limit)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	byID, err := hydrateExamPools(ctx, h.examPoolRepo, attempts)
	if err != nil {
		return nil, err
	}

	return &ListExamAttemptsResult{Attempts: attempts, ExamPools: byID, Pagination: pg}, nil
}

// hydrateExamPools fetches the distinct question sets behind a page of
// attempts in one query. Distinct matters: a cached exam served to the
// same child twice would otherwise be fetched twice.
func hydrateExamPools(ctx context.Context, examPoolRepo exam.IExamPoolRepository, attempts []*exam.ExamLink) (map[int64]*exam.ExamPool, error) {
	if len(attempts) == 0 {
		return map[int64]*exam.ExamPool{}, nil
	}

	seen := make(map[int64]struct{}, len(attempts))
	ids := make([]int64, 0, len(attempts))
	for _, a := range attempts {
		if _, ok := seen[a.ExamId()]; ok {
			continue
		}
		seen[a.ExamId()] = struct{}{}
		ids = append(ids, a.ExamId())
	}

	rows, err := examPoolRepo.ListByExamIds(ctx, ids)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	byID := make(map[int64]*exam.ExamPool, len(rows))
	for _, r := range rows {
		byID[r.ExamId()] = r
	}
	return byID, nil
}
