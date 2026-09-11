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
	Status    *string
	Page      int64
	Limit     int64
}

// ListExamAttemptsResult pairs the attempts with the question sets they
// were drawn from, keyed by ai_exam_id. A card needs the title, which
// lives on the shared row, not on the attempt.
type ListExamAttemptsResult struct {
	Attempts   []*exam.UserAiExam
	AiExams    map[int64]*exam.AiExam
	Pagination *pagination.Pagination
}

type ListExamAttemptsQueryHandler struct {
	attemptRepo exam.IUserAiExamRepository
	aiExamRepo  exam.IAiExamRepository
}

func NewListExamAttemptsQueryHandler(
	attemptRepo exam.IUserAiExamRepository,
	aiExamRepo exam.IAiExamRepository,
) *ListExamAttemptsQueryHandler {
	return &ListExamAttemptsQueryHandler{attemptRepo: attemptRepo, aiExamRepo: aiExamRepo}
}

func (h *ListExamAttemptsQueryHandler) Handle(ctx context.Context, q ListExamAttemptsQuery) (*ListExamAttemptsResult, error) {
	attempts, pg, err := h.attemptRepo.ListAttempts(ctx, exam.ListAttemptsFilter{
		ProfileID: q.ProfileID,
		ExamType:  q.ExamType,
		Status:    q.Status,
	}, q.Page, q.Limit)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	byID, err := h.hydrateAiExams(ctx, attempts)
	if err != nil {
		return nil, err
	}

	return &ListExamAttemptsResult{Attempts: attempts, AiExams: byID, Pagination: pg}, nil
}

// hydrateAiExams fetches the distinct question sets behind a page of
// attempts in one query. Distinct matters: a cached exam served to the
// same child twice would otherwise be fetched twice.
func (h *ListExamAttemptsQueryHandler) hydrateAiExams(ctx context.Context, attempts []*exam.UserAiExam) (map[int64]*exam.AiExam, error) {
	if len(attempts) == 0 {
		return map[int64]*exam.AiExam{}, nil
	}

	seen := make(map[int64]struct{}, len(attempts))
	ids := make([]int64, 0, len(attempts))
	for _, a := range attempts {
		if _, ok := seen[a.AiExamId()]; ok {
			continue
		}
		seen[a.AiExamId()] = struct{}{}
		ids = append(ids, a.AiExamId())
	}

	rows, err := h.aiExamRepo.ListByAiExamIds(ctx, ids)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	byID := make(map[int64]*exam.AiExam, len(rows))
	for _, r := range rows {
		byID[r.AiExamId()] = r
	}
	return byID, nil
}
