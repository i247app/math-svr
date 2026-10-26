// Package question is the question bank module (ma_questions), admin only
// and independent of the exam module: /questions/*.
package question

import (
	"context"
	"errors"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/question"
	dto "math-ai.com/math-ai/internal/application/dto/question"
	query "math-ai.com/math-ai/internal/application/query/question"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	sctx "math-ai.com/math-ai/internal/shared/context"
)

// imageUrlTTL bounds how long a presigned question / option image URL stays
// valid.
const imageUrlTTL = 1 * time.Hour

// Service is the admin surface over the question bank. storage may be nil;
// image URLs are then null.
type Service struct {
	createCmd *command.CreateQuestionsCommandHandler
	updateCmd *command.UpdateQuestionCommandHandler
	reviewCmd *command.ReviewQuestionCommandHandler
	deleteCmd *command.DeleteQuestionCommandHandler
	listQuery *query.ListQuestionsQueryHandler
	storage   *storage.Adapter
}

func NewService(uow transaction.UnitOfWork, repo domain.IRepository, storageProvider *storage.Adapter) *Service {
	return &Service{
		createCmd: command.NewCreateQuestionsCommandHandler(uow),
		updateCmd: command.NewUpdateQuestionCommandHandler(uow),
		reviewCmd: command.NewReviewQuestionCommandHandler(uow),
		deleteCmd: command.NewDeleteQuestionCommandHandler(uow),
		listQuery: query.NewListQuestionsQueryHandler(repo),
		storage:   storageProvider,
	}
}

var ErrAdminGateMissing = errors.New("question: route did not pass the admin gate")

// requireAdminGate fails closed when a route was registered without the
// admin middleware: every bank response carries the answer key.
func requireAdminGate(ctx context.Context) error {
	if sctx.GetAdminVia(ctx) == "" {
		return errs.NewError(ctx, status.FORBIDDEN, nil, ErrAdminGateMissing)
	}
	return nil
}

// adminActor is the admin's uid for create_id, or nil when the request came
// in with the X-Api-Admin-Key header and has no account behind it.
func adminActor(ctx context.Context) *int64 {
	if uid := sctx.UID(ctx); uid > 0 {
		return &uid
	}
	return nil
}

// CreateQuestions stores every question of the request, or none.
func (s *Service) CreateQuestions(ctx context.Context, req *dto.CreateQuestionsReq) (*dto.CreateQuestionsRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	questions, err := ValidateCreateQuestions(ctx, req)
	if err != nil {
		return nil, err
	}
	actor := adminActor(ctx)
	saved, err := s.createCmd.Handle(ctx, command.CreateQuestionsCommand{Questions: questions, ActorUID: actor})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question.created admin_uid=%d via=%s count=%d", sctx.UID(ctx), sctx.GetAdminVia(ctx), len(saved))
	return &dto.CreateQuestionsRes{Questions: s.toResponses(ctx, saved)}, nil
}

// UpdateQuestion edits one question: absent fields keep their value, the
// merged question must pass every create rule, and its new content must not
// be another live question's. Review state is left alone.
func (s *Service) UpdateQuestion(ctx context.Context, req *dto.UpdateQuestionReq) (*dto.QuestionRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if err := ValidateUpdateQuestion(ctx, req); err != nil {
		return nil, err
	}
	saved, err := s.updateCmd.Handle(ctx, command.UpdateQuestionCommand{
		QuestionID: req.QuestionID,
		ActorUID:   adminActor(ctx),
		Apply: func(stored *domain.Question) (*domain.Question, error) {
			return MergeUpdate(ctx, stored, req)
		},
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question.updated admin_uid=%d via=%s question_id=%d", sctx.UID(ctx), sctx.GetAdminVia(ctx), saved.QuestionId())
	return &dto.QuestionRes{Question: s.toResponses(ctx, []*domain.Question{saved})[0]}, nil
}

// MarkQuestionStatus sets one question's review state; content stays.
func (s *Service) MarkQuestionStatus(ctx context.Context, req *dto.MarkQuestionStatusReq) (*dto.QuestionRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if err := ValidateMarkQuestionStatus(ctx, req); err != nil {
		return nil, err
	}
	saved, err := s.reviewCmd.Handle(ctx, command.ReviewQuestionCommand{
		QuestionID: req.QuestionID,
		ActorUID:   adminActor(ctx),
		Status:     &req.QuestionStatus,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question.status_marked admin_uid=%d via=%s question_id=%d status=%s",
		sctx.UID(ctx), sctx.GetAdminVia(ctx), saved.QuestionId(), req.QuestionStatus)
	return &dto.QuestionRes{Question: s.toResponses(ctx, []*domain.Question{saved})[0]}, nil
}

// VerifyQuestion records one more admin check: verified_count + 1, nothing
// else changes.
func (s *Service) VerifyQuestion(ctx context.Context, req *dto.VerifyQuestionReq) (*dto.QuestionRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if req.QuestionID <= 0 {
		return nil, errs.NewError(ctx, status.QUESTION_MISSING_ID, nil, errors.New("question_id is required"))
	}
	saved, err := s.reviewCmd.Handle(ctx, command.ReviewQuestionCommand{
		QuestionID: req.QuestionID, ActorUID: adminActor(ctx), Verify: true,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question.verified admin_uid=%d via=%s question_id=%d verified_count=%d",
		sctx.UID(ctx), sctx.GetAdminVia(ctx), saved.QuestionId(), saved.VerifiedCount())
	return &dto.QuestionRes{Question: s.toResponses(ctx, []*domain.Question{saved})[0]}, nil
}

// SoftDeleteQuestion hides a live question; the row stays and its content
// may be created again.
func (s *Service) SoftDeleteQuestion(ctx context.Context, req *dto.DeleteQuestionReq) (*dto.DeleteQuestionRes, error) {
	return s.deleteQuestion(ctx, req, false)
}

// ForceDeleteQuestion removes a question's row for good, live or already
// soft-deleted.
func (s *Service) ForceDeleteQuestion(ctx context.Context, req *dto.DeleteQuestionReq) (*dto.DeleteQuestionRes, error) {
	return s.deleteQuestion(ctx, req, true)
}

func (s *Service) deleteQuestion(ctx context.Context, req *dto.DeleteQuestionReq, force bool) (*dto.DeleteQuestionRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if req.QuestionID <= 0 {
		return nil, errs.NewError(ctx, status.QUESTION_MISSING_ID, nil, errors.New("question_id is required"))
	}
	if err := s.deleteCmd.Handle(ctx, command.DeleteQuestionCommand{QuestionID: req.QuestionID, ActorUID: adminActor(ctx), Force: force}); err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question.deleted admin_uid=%d via=%s question_id=%d force=%t",
		sctx.UID(ctx), sctx.GetAdminVia(ctx), req.QuestionID, force)
	return &dto.DeleteQuestionRes{}, nil
}

// ListQuestions pages the bank, newest first, key and explanation included.
func (s *Service) ListQuestions(ctx context.Context, req *dto.ListQuestionsReq) (*dto.ListQuestionsRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	filter, err := ValidateListQuestions(ctx, req)
	if err != nil {
		return nil, err
	}
	rows, page, err := s.listQuery.Handle(ctx, query.ListQuestionsQuery{Filter: filter, Page: req.Page, Size: req.Size})
	if err != nil {
		return nil, err
	}
	return &dto.ListQuestionsRes{Questions: s.toResponses(ctx, rows), Pagination: page}, nil
}

func (s *Service) toResponses(ctx context.Context, qs []*domain.Question) []*dto.QuestionResponse {
	out := dto.DomainListToResponse(qs)
	for _, r := range out {
		r.QuestionImageUrl = s.presign(ctx, r.QuestionID, r.QuestionImageKey)
		for i := range r.Options {
			r.Options[i].ImageUrl = s.presign(ctx, r.QuestionID, r.Options[i].ImageKey)
		}
	}
	return out
}

// presign returns a short-lived URL for key, or nil when there is no key,
// no storage, or signing failed (logged; the read does not fail for it).
func (s *Service) presign(ctx context.Context, questionID int64, key *string) *string {
	if s.storage == nil || key == nil || *key == "" {
		return nil
	}
	url, err := s.storage.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{Key: *key, Expiration: imageUrlTTL})
	if err != nil {
		logger.From(ctx).Warnf("question.image presign failed question_id=%d err=%v", questionID, err)
		return nil
	}
	return &url
}
