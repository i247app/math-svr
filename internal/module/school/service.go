package school

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/school"
	dto "math-ai.com/math-ai/internal/application/dto/school"
	query "math-ai.com/math-ai/internal/application/query/school"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/school"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// imageUrlTTL bounds how long a generated school image URL is valid.
// Mirrors the program module so cross-aggregate UX stays consistent.
const imageUrlTTL = 1 * time.Hour

// Service is the school module's public façade. It composes the
// CQRS handlers behind the validators and owns no I/O of its own —
// every write goes through the UoW-backed commands, every read through
// the repo-bound queries. storageProvider may be nil; in that case
// responses simply omit image_url.
type Service struct {
	getSchoolQuery       *query.GetSchoolByIdQueryHandler
	listSchoolsQuery     *query.ListSchoolsQueryHandler
	createSchoolCmd      *command.CreateSchoolCommandHandler
	updateSchoolCmd      *command.UpdateSchoolCommandHandler
	softDeleteSchoolCmd  *command.SoftDeleteSchoolCommandHandler
	forceDeleteSchoolCmd *command.ForceDeleteSchoolCommandHandler
	storageProvider      *storage.Adapter
}

func NewService(
	schoolRepo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getSchoolQuery:       query.NewGetSchoolByIdQueryHandler(schoolRepo),
		listSchoolsQuery:     query.NewListSchoolsQueryHandler(schoolRepo),
		createSchoolCmd:      command.NewCreateSchoolCommandHandler(uow),
		updateSchoolCmd:      command.NewUpdateSchoolCommandHandler(uow),
		softDeleteSchoolCmd:  command.NewSoftDeleteSchoolCommandHandler(uow),
		forceDeleteSchoolCmd: command.NewForceDeleteSchoolCommandHandler(uow),
		storageProvider:      storageProvider,
	}
}

func (s *Service) CreateSchool(ctx context.Context, req *dto.CreateSchoolReq) (*dto.CreateSchoolRes, error) {
	if err := ValidateCreateSchool(ctx, req); err != nil {
		return nil, err
	}

	created, err := s.createSchoolCmd.Handle(ctx, command.CreateSchoolCommand{
		Name:        req.Name,
		Description: req.Description,
		ImageKey:    req.ImageKey,
		District:    req.District,
		Province:    req.Province,
		Note:        req.Note,
	})
	if err != nil {
		return nil, err
	}

	resp := dto.DomainToResponse(created)
	s.populateImageUrl(ctx, resp)
	return &dto.CreateSchoolRes{School: resp}, nil
}

func (s *Service) UpdateSchool(ctx context.Context, req *dto.UpdateSchoolReq) (*dto.UpdateSchoolRes, error) {
	if err := ValidateUpdateSchool(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.updateSchoolCmd.Handle(ctx, command.UpdateSchoolCommand{
		SchoolID:    req.SchoolID,
		Name:        req.Name,
		Description: req.Description,
		ImageKey:    req.ImageKey,
		District:    req.District,
		Province:    req.Province,
		Note:        req.Note,
	})
	if err != nil {
		return nil, err
	}

	resp := dto.DomainToResponse(updated)
	s.populateImageUrl(ctx, resp)
	return &dto.UpdateSchoolRes{School: resp}, nil
}

func (s *Service) SoftDeleteSchool(ctx context.Context, req *dto.DeleteSchoolReq) (*dto.DeleteSchoolRes, error) {
	if err := ValidateDeleteSchool(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteSchoolCmd.Handle(ctx, command.SoftDeleteSchoolCommand{SchoolID: req.SchoolID}); err != nil {
		return nil, err
	}
	return &dto.DeleteSchoolRes{}, nil
}

func (s *Service) ForceDeleteSchool(ctx context.Context, req *dto.DeleteSchoolReq) (*dto.DeleteSchoolRes, error) {
	if err := ValidateDeleteSchool(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteSchoolCmd.Handle(ctx, command.ForceDeleteSchoolCommand{SchoolID: req.SchoolID}); err != nil {
		return nil, err
	}
	return &dto.DeleteSchoolRes{}, nil
}

func (s *Service) GetSchool(ctx context.Context, req *dto.GetSchoolReq) (*dto.GetSchoolRes, error) {
	if err := ValidateGetSchool(ctx, req); err != nil {
		return nil, err
	}

	found, err := s.getSchoolQuery.Handle(ctx, query.GetSchoolByIdQuery{SchoolID: req.SchoolID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.SCHOOL_NOT_FOUND, nil,
			ErrSchoolNotFound)
	}

	resp := dto.DomainToResponse(found)
	s.populateImageUrl(ctx, resp)
	return &dto.GetSchoolRes{School: resp}, nil
}

func (s *Service) ListSchools(ctx context.Context, req *dto.ListSchoolsReq) (*dto.ListSchoolsRes, error) {
	if err := ValidateListSchools(ctx, req); err != nil {
		return nil, err
	}

	schools, pg, err := s.listSchoolsQuery.Handle(ctx, query.ListSchoolsQuery{
		Search:    req.Search,
		District:  req.District,
		Province:  req.Province,
		SchoolIDs: req.SchoolIDs,
		Page:      req.Page,
		Limit:     req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	responses := dto.DomainListToResponse(schools)
	for _, r := range responses {
		s.populateImageUrl(ctx, r)
	}
	return &dto.ListSchoolsRes{
		Schools:    responses,
		Pagination: pg,
	}, nil
}

// populateImageUrl mutates resp in place to add a short-lived presigned
// URL when the school carries an image_key. No-op if storage is disabled
// or the school has no image.
func (s *Service) populateImageUrl(ctx context.Context, resp *dto.SchoolResponse) {
	if resp == nil || s.storageProvider == nil || resp.ImageKey == nil || *resp.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.ImageKey,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("school.image presign failed school_id=%s err=%v", resp.SchoolID, err)
		return
	}
	resp.ImageUrl = &url
}
