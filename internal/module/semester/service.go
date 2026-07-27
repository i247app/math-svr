package semester

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/semester"
	dto "math-ai.com/math-ai/internal/application/dto/semester"
	query "math-ai.com/math-ai/internal/application/query/semester"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const imageUrlTTL = 1 * time.Hour

// Service is the semester module's public façade. It composes the
// CQRS handlers behind the validators and owns no I/O of its own —
// every write goes through the UoW-backed commands, every read through
// the repo-bound queries.
type Service struct {
	listSemestersQuery     *query.ListSemestersQueryHandler
	getSemesterQuery       *query.GetSemesterByIdQueryHandler
	createSemesterCmd      *command.CreateSemesterCommandHandler
	updateSemesterCmd      *command.UpdateSemesterCommandHandler
	softDeleteSemesterCmd  *command.SoftDeleteSemesterCommandHandler
	forceDeleteSemesterCmd *command.ForceDeleteSemesterCommandHandler
	storageProvider        *storage.Adapter
}

func NewService(
	semesterRepo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		listSemestersQuery:     query.NewListSemestersQueryHandler(semesterRepo),
		getSemesterQuery:       query.NewGetSemesterByIdQueryHandler(semesterRepo),
		createSemesterCmd:      command.NewCreateSemesterCommandHandler(uow),
		updateSemesterCmd:      command.NewUpdateSemesterCommandHandler(uow),
		softDeleteSemesterCmd:  command.NewSoftDeleteSemesterCommandHandler(uow),
		forceDeleteSemesterCmd: command.NewForceDeleteSemesterCommandHandler(uow),
		storageProvider:        storageProvider,
	}
}

func (s *Service) CreateSemester(ctx context.Context, req *dto.CreateSemesterReq) (*dto.CreateSemesterRes, error) {
	if err := ValidateCreateSemester(ctx, req); err != nil {
		return nil, err
	}

	created, err := s.createSemesterCmd.Handle(ctx, command.CreateSemesterCommand{
		Name:         req.Name,
		Description:  req.Description,
		ImageKey:     req.ImageKey,
		DisplayOrder: req.DisplayOrder,
		Note:         req.Note,
	})
	if err != nil {
		return nil, err
	}
	resp := dto.DomainToResponse(created)
	s.populateImageUrl(ctx, resp)
	return &dto.CreateSemesterRes{Semester: resp}, nil
}

func (s *Service) UpdateSemester(ctx context.Context, req *dto.UpdateSemesterReq) (*dto.UpdateSemesterRes, error) {
	if err := ValidateUpdateSemester(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.updateSemesterCmd.Handle(ctx, command.UpdateSemesterCommand{
		SemesterID:   req.SemesterID,
		Name:         req.Name,
		Description:  req.Description,
		ImageKey:     req.ImageKey,
		DisplayOrder: req.DisplayOrder,
		Note:         req.Note,
	})
	if err != nil {
		return nil, err
	}
	resp := dto.DomainToResponse(updated)
	s.populateImageUrl(ctx, resp)
	return &dto.UpdateSemesterRes{Semester: resp}, nil
}

func (s *Service) SoftDeleteSemester(ctx context.Context, req *dto.DeleteSemesterReq) (*dto.DeleteSemesterRes, error) {
	if err := ValidateDeleteSemester(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteSemesterCmd.Handle(ctx, command.SoftDeleteSemesterCommand{SemesterID: req.SemesterID}); err != nil {
		return nil, err
	}
	return &dto.DeleteSemesterRes{}, nil
}

func (s *Service) ForceDeleteSemester(ctx context.Context, req *dto.DeleteSemesterReq) (*dto.DeleteSemesterRes, error) {
	if err := ValidateDeleteSemester(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteSemesterCmd.Handle(ctx, command.ForceDeleteSemesterCommand{SemesterID: req.SemesterID}); err != nil {
		return nil, err
	}
	return &dto.DeleteSemesterRes{}, nil
}

func (s *Service) GetSemester(ctx context.Context, req *dto.GetSemesterReq) (*dto.GetSemesterRes, error) {
	if err := ValidateGetSemester(ctx, req); err != nil {
		return nil, err
	}
	sm, err := s.getSemesterQuery.Handle(ctx, query.GetSemesterByIdQuery{
		SemesterID: req.SemesterID,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if sm == nil {
		return nil, errs.NewError(ctx, status.SEMESTER_NOT_FOUND, nil, ErrSemesterNotFound)
	}
	resp := dto.DomainToResponse(sm)
	s.populateImageUrl(ctx, resp)
	return &dto.GetSemesterRes{Semester: resp}, nil
}

func (s *Service) ListSemesters(ctx context.Context, req *dto.ListSemestersReq) (*dto.ListSemestersRes, error) {
	if err := ValidateListSemesters(ctx, req); err != nil {
		return nil, err
	}

	semesters, pg, err := s.listSemestersQuery.Handle(ctx, &query.ListSemestersQuery{
		Page:  req.Page,
		Limit: req.Size,
	})
	if err != nil {
		return nil, err
	}

	responses := dto.DomainListToResponse(semesters)
	for _, r := range responses {
		s.populateImageUrl(ctx, r)
	}
	return &dto.ListSemestersRes{
		Semesters:  responses,
		Pagination: pg,
	}, nil
}

func (s *Service) populateImageUrl(ctx context.Context, resp *dto.SemesterResponse) {
	if resp == nil || s.storageProvider == nil || resp.ImageKey == nil || *resp.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.ImageKey,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("semester.image presign failed semester_id=%d err=%v", resp.SemesterID, err)
		return
	}
	resp.ImageUrl = &url
}
