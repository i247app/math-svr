// Package questionpool is the question bank module (ma_question_pools), admin only
// and independent of the exam module: /question-pools/*.
package questionpool

import (
	"context"
	"errors"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/questionpool"
	dto "math-ai.com/math-ai/internal/application/dto/questionpool"
	query "math-ai.com/math-ai/internal/application/query/questionpool"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/questionpool"
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
	createCmd *command.CreateQuestionPoolsCommandHandler
	updateCmd *command.UpdateQuestionPoolCommandHandler
	reviewCmd *command.ReviewQuestionPoolCommandHandler
	deleteCmd *command.DeleteQuestionPoolCommandHandler
	importCmd *command.ImportQuestionPoolsCommandHandler
	listQuery *query.ListQuestionPoolsQueryHandler
	repo      domain.IRepository // import's duplicate check, outside any transaction
	storage   *storage.Adapter
}

func NewService(uow transaction.UnitOfWork, repo domain.IRepository, storageProvider *storage.Adapter) *Service {
	return &Service{
		createCmd: command.NewCreateQuestionPoolsCommandHandler(uow),
		updateCmd: command.NewUpdateQuestionPoolCommandHandler(uow),
		reviewCmd: command.NewReviewQuestionPoolCommandHandler(uow),
		deleteCmd: command.NewDeleteQuestionPoolCommandHandler(uow),
		importCmd: command.NewImportQuestionPoolsCommandHandler(uow),
		listQuery: query.NewListQuestionPoolsQueryHandler(repo),
		repo:      repo,
		storage:   storageProvider,
	}
}

var ErrAdminGateMissing = errors.New("question_pool: route did not pass the admin gate")

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
func (s *Service) CreateQuestionPools(ctx context.Context, req *dto.CreateQuestionPoolsReq) (*dto.CreateQuestionPoolsRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	questions, err := ValidateCreateQuestionPools(ctx, req)
	if err != nil {
		return nil, err
	}
	actor := adminActor(ctx)
	saved, err := s.createCmd.Handle(ctx, command.CreateQuestionPoolsCommand{QuestionPools: questions, ActorUID: actor})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question_pool.created admin_uid=%d via=%s count=%d", sctx.UID(ctx), sctx.GetAdminVia(ctx), len(saved))
	res := &dto.CreateQuestionPoolsRes{QuestionPools: s.toResponses(ctx, saved)}
	return res, nil
}

// UpdateQuestion edits one question: absent fields keep their value, the
// merged question must pass every create rule, and its new content must not
// be another live question's. Review state is left alone.
func (s *Service) UpdateQuestionPool(ctx context.Context, req *dto.UpdateQuestionPoolReq) (*dto.QuestionPoolRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if err := ValidateUpdateQuestionPool(ctx, req); err != nil {
		return nil, err
	}
	saved, err := s.updateCmd.Handle(ctx, command.UpdateQuestionPoolCommand{
		QuestionPoolID: req.QuestionPoolID,
		ActorUID:       adminActor(ctx),
		Apply: func(stored *domain.QuestionPool) (*domain.QuestionPool, error) {
			return MergeUpdate(ctx, stored, req)
		},
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question_pool.updated admin_uid=%d via=%s question_pool_id=%d", sctx.UID(ctx), sctx.GetAdminVia(ctx), saved.QuestionPoolId())
	res := &dto.QuestionPoolRes{QuestionPool: s.toResponses(ctx, []*domain.QuestionPool{saved})[0]}
	return res, nil
}

// MarkQuestionStatus sets one question's review state; content stays.
func (s *Service) MarkQuestionPoolStatus(ctx context.Context, req *dto.MarkQuestionPoolStatusReq) (*dto.QuestionPoolRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if err := ValidateMarkQuestionPoolStatus(ctx, req); err != nil {
		return nil, err
	}
	saved, err := s.reviewCmd.Handle(ctx, command.ReviewQuestionPoolCommand{
		QuestionPoolID: req.QuestionPoolID,
		ActorUID:       adminActor(ctx),
		Status:         &req.QuestionStatus,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question_pool.status_marked admin_uid=%d via=%s question_pool_id=%d status=%s",
		sctx.UID(ctx), sctx.GetAdminVia(ctx), saved.QuestionPoolId(), req.QuestionStatus)
	res := &dto.QuestionPoolRes{QuestionPool: s.toResponses(ctx, []*domain.QuestionPool{saved})[0]}
	return res, nil
}

// VerifyQuestion records one more admin check: verified_count + 1, nothing
// else changes.
func (s *Service) VerifyQuestionPool(ctx context.Context, req *dto.VerifyQuestionPoolReq) (*dto.QuestionPoolRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if req.QuestionPoolID <= 0 {
		return nil, errs.NewError(ctx, status.QUESTION_POOL_MISSING_ID, nil, errors.New("question_pool_id is required"))
	}
	saved, err := s.reviewCmd.Handle(ctx, command.ReviewQuestionPoolCommand{
		QuestionPoolID: req.QuestionPoolID, ActorUID: adminActor(ctx), Verify: true,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question_pool.verified admin_uid=%d via=%s question_pool_id=%d verified_count=%d",
		sctx.UID(ctx), sctx.GetAdminVia(ctx), saved.QuestionPoolId(), saved.VerifiedCount())
	res := &dto.QuestionPoolRes{QuestionPool: s.toResponses(ctx, []*domain.QuestionPool{saved})[0]}
	return res, nil
}

// SoftDeleteQuestion hides a live question; the row stays and its content
// may be created again.
func (s *Service) SoftDeleteQuestionPool(ctx context.Context, req *dto.DeleteQuestionPoolReq) (*dto.DeleteQuestionPoolRes, error) {
	return s.deleteQuestionPool(ctx, req, false)
}

// ForceDeleteQuestion removes a question's row for good, live or already
// soft-deleted.
func (s *Service) ForceDeleteQuestionPool(ctx context.Context, req *dto.DeleteQuestionPoolReq) (*dto.DeleteQuestionPoolRes, error) {
	return s.deleteQuestionPool(ctx, req, true)
}

func (s *Service) deleteQuestionPool(ctx context.Context, req *dto.DeleteQuestionPoolReq, force bool) (*dto.DeleteQuestionPoolRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if req.QuestionPoolID <= 0 {
		return nil, errs.NewError(ctx, status.QUESTION_POOL_MISSING_ID, nil, errors.New("question_pool_id is required"))
	}
	if err := s.deleteCmd.Handle(ctx, command.DeleteQuestionPoolCommand{QuestionPoolID: req.QuestionPoolID, ActorUID: adminActor(ctx), Force: force}); err != nil {
		return nil, err
	}
	logger.From(ctx).Infof("question_pool.deleted admin_uid=%d via=%s question_pool_id=%d force=%t",
		sctx.UID(ctx), sctx.GetAdminVia(ctx), req.QuestionPoolID, force)
	res := &dto.DeleteQuestionPoolRes{}
	return res, nil
}

// ListQuestions pages the bank, newest first, key and explanation included.
func (s *Service) ListQuestionPools(ctx context.Context, req *dto.ListQuestionPoolsReq) (*dto.ListQuestionPoolsRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	filter, err := ValidateListQuestionPools(ctx, req)
	if err != nil {
		return nil, err
	}

	filterQry := query.ListQuestionPoolsQuery{
		Filter: filter,
		Page:   req.Page,
		Size:   req.Size,
	}
	rows, page, err := s.listQuery.Handle(ctx, filterQry)
	if err != nil {
		return nil, err
	}
	res := &dto.ListQuestionPoolsRes{QuestionPools: s.toResponses(ctx, rows), Pagination: page}
	return res, nil
}

func (s *Service) toResponses(ctx context.Context, qs []*domain.QuestionPool) []*dto.QuestionPoolResponse {
	out := dto.DomainListToResponse(qs)
	for _, r := range out {
		r.QuestionImageUrl = s.presign(ctx, r.QuestionPoolID, r.QuestionImageKey)
		for i := range r.Options {
			r.Options[i].ImageUrl = s.presign(ctx, r.QuestionPoolID, r.Options[i].ImageKey)
		}
	}
	return out
}

// presign returns a short-lived URL for key, or nil when there is no key,
// no storage, or signing failed (logged; the read does not fail for it).
func (s *Service) presign(ctx context.Context, questionPoolID int64, key *string) *string {
	if s.storage == nil || key == nil || *key == "" {
		return nil
	}
	url, err := s.storage.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{Key: *key, Expiration: imageUrlTTL})
	if err != nil {
		logger.From(ctx).Warnf("question_pool.image presign failed question_pool_id=%d err=%v", questionPoolID, err)
		return nil
	}
	return &url
}
