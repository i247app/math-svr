package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// GetExamAttemptQuery reads one sitting. ProfileID is required, not
// optional: an attempt id is a plain integer, so the read has to prove
// whose it is rather than trusting whoever asked.
type GetExamAttemptQuery struct {
	UserAiExamID int64
	UserID       int64
	ProfileID    int64
}

// ExamAttemptDetail is the review screen in one read: the sitting, the
// question set it was drawn from, and the per-question log.
//
// Details is empty while the attempt is still IN_PROGRESS — nothing is
// logged until submit — so a client rendering an unfinished exam falls
// back to AiExam's questions, which is exactly what it was handed at
// generation time.
type ExamAttemptDetail struct {
	Attempt *exam.UserAiExam
	AiExam  *exam.AiExam
	Details []*exam.UserExamDetail
}

type GetExamAttemptQueryHandler struct {
	attemptRepo exam.IUserAiExamRepository
	aiExamRepo  exam.IAiExamRepository
	detailRepo  exam.IUserExamDetailRepository
}

func NewGetExamAttemptQueryHandler(
	attemptRepo exam.IUserAiExamRepository,
	aiExamRepo exam.IAiExamRepository,
	detailRepo exam.IUserExamDetailRepository,
) *GetExamAttemptQueryHandler {
	return &GetExamAttemptQueryHandler{
		attemptRepo: attemptRepo,
		aiExamRepo:  aiExamRepo,
		detailRepo:  detailRepo,
	}
}

func (h *GetExamAttemptQueryHandler) Handle(ctx context.Context, q GetExamAttemptQuery) (*ExamAttemptDetail, error) {
	attempt, err := h.attemptRepo.FindByUserAiExamId(ctx, q.UserAiExamID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if attempt == nil {
		return nil, errs.NewError(ctx, status.EXAM_ATTEMPT_NOT_FOUND, nil, nil)
	}
	if attempt.UserId() != q.UserID || attempt.ProfileId() != q.ProfileID {
		return nil, errs.NewError(ctx, status.EXAM_ATTEMPT_NOT_OWNED, nil, nil)
	}

	aiExam, err := h.aiExamRepo.FindByAiExamId(ctx, attempt.AiExamId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if aiExam == nil {
		return nil, errs.NewError(ctx, status.EXAM_NOT_FOUND, nil, nil)
	}

	details, err := h.detailRepo.ListByUserAiExamId(ctx, attempt.UserAiExamId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &ExamAttemptDetail{Attempt: attempt, AiExam: aiExam, Details: details}, nil
}
