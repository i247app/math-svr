package grade

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/grade"
	dto "math-ai.com/math-ai/internal/application/dto/grade"
	query "math-ai.com/math-ai/internal/application/query/grade"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/grade"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
)

const imageUrlTTL = 1 * time.Hour

// Service is the grade module's public façade. It composes the
// CQRS handlers behind the validators and owns no I/O of its own —
// every write goes through the UoW-backed commands, every read through
// the repo-bound queries.
type Service struct {
	listGradesQuery     *query.ListGradesQueryHandler
	getGradeQuery       *query.GetGradeByIdQueryHandler
	createGradeCmd      *command.CreateGradeCommandHandler
	updateGradeCmd      *command.UpdateGradeCommandHandler
	softDeleteGradeCmd  *command.SoftDeleteGradeCommandHandler
	forceDeleteGradeCmd *command.ForceDeleteGradeCommandHandler
	storageProvider     *storage.Adapter
}

func NewService(
	gradeRepo domain.IRepository,
	translationRepo domain.ITranslationRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		listGradesQuery:     query.NewListGradesQueryHandler(gradeRepo),
		getGradeQuery:       query.NewGetGradeByIdQueryHandler(gradeRepo, translationRepo),
		createGradeCmd:      command.NewCreateGradeCommandHandler(uow),
		updateGradeCmd:      command.NewUpdateGradeCommandHandler(uow),
		softDeleteGradeCmd:  command.NewSoftDeleteGradeCommandHandler(uow),
		forceDeleteGradeCmd: command.NewForceDeleteGradeCommandHandler(uow),
		storageProvider:     storageProvider,
	}
}

func (s *Service) CreateGrade(ctx context.Context, req *dto.CreateGradeReq) (*dto.CreateGradeRes, error) {
	if err := ValidateCreateGrade(ctx, req); err != nil {
		return nil, err
	}

	created, err := s.createGradeCmd.Handle(ctx, command.CreateGradeCommand{
		Label:        req.Label,
		Description:  req.Description,
		ImageKey:     req.ImageKey,
		DisplayOrder: req.DisplayOrder,
		Note:         req.Note,
		Translations: translationInputsFromDTO(req.Translations),
	})
	if err != nil {
		return nil, err
	}
	resp := dto.DomainToResponse(created)
	s.populateImageUrl(ctx, resp)
	return &dto.CreateGradeRes{Grade: resp}, nil
}

func (s *Service) UpdateGrade(ctx context.Context, req *dto.UpdateGradeReq) (*dto.UpdateGradeRes, error) {
	if err := ValidateUpdateGrade(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.updateGradeCmd.Handle(ctx, command.UpdateGradeCommand{
		GradeID:      req.GradeID,
		Label:        req.Label,
		Description:  req.Description,
		ImageKey:     req.ImageKey,
		DisplayOrder: req.DisplayOrder,
		Note:         req.Note,
		Translations: translationInputsFromDTO(req.Translations),
	})
	if err != nil {
		return nil, err
	}
	resp := dto.DomainToResponse(updated)
	s.populateImageUrl(ctx, resp)
	return &dto.UpdateGradeRes{Grade: resp}, nil
}

func (s *Service) SoftDeleteGrade(ctx context.Context, req *dto.DeleteGradeReq) (*dto.DeleteGradeRes, error) {
	if err := ValidateDeleteGrade(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteGradeCmd.Handle(ctx, command.SoftDeleteGradeCommand{GradeID: req.GradeID}); err != nil {
		return nil, err
	}
	return &dto.DeleteGradeRes{}, nil
}

func (s *Service) ForceDeleteGrade(ctx context.Context, req *dto.DeleteGradeReq) (*dto.DeleteGradeRes, error) {
	if err := ValidateDeleteGrade(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteGradeCmd.Handle(ctx, command.ForceDeleteGradeCommand{GradeID: req.GradeID}); err != nil {
		return nil, err
	}
	return &dto.DeleteGradeRes{}, nil
}

func (s *Service) GetGrade(ctx context.Context, req *dto.GetGradeReq) (*dto.GetGradeRes, error) {
	if err := ValidateGetGrade(ctx, req); err != nil {
		return nil, err
	}
	g, err := s.getGradeQuery.Handle(ctx, query.GetGradeByIdQuery{
		GradeID:  req.GradeID,
		Language: req.Language,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if g == nil {
		return nil, errs.NewError(ctx, status.GRADE_NOT_FOUND, nil,
			ErrGradeNotFound)
	}
	resp := dto.DomainToResponse(g)
	s.populateImageUrl(ctx, resp)
	return &dto.GetGradeRes{Grade: resp}, nil
}

func (s *Service) ListGrades(ctx context.Context, req *dto.ListGradesReq) (*dto.ListGradesRes, error) {
	if err := ValidateListGrades(ctx, req); err != nil {
		return nil, err
	}

	language := req.Language
	if language == "" {
		language = metadata.GetClientLanguage(ctx).ToEnumLanguage()
	}

	grades, pg, err := s.listGradesQuery.Handle(ctx, &query.ListGradesQuery{
		Language: language,
		GradeIDs: req.GradeIDs,
		Page:     req.Page,
		Limit:    req.Size,
	})
	if err != nil {
		return nil, err
	}

	responses := dto.DomainListToResponse(grades)
	for _, r := range responses {
		s.populateImageUrl(ctx, r)
	}
	return &dto.ListGradesRes{
		Grades:     responses,
		Pagination: pg,
	}, nil
}

func (s *Service) populateImageUrl(ctx context.Context, resp *dto.GradeResponse) {
	if resp == nil || s.storageProvider == nil || resp.ImageKey == nil || *resp.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.ImageKey,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("grade.image presign failed grade_id=%d err=%v", resp.GradeID, err)
		return
	}
	resp.ImageUrl = &url
}

func translationInputsFromDTO(in []*dto.GradeTranslationDTO) []command.TranslationInput {
	if len(in) == 0 {
		return nil
	}
	out := make([]command.TranslationInput, 0, len(in))
	for _, t := range in {
		if t == nil {
			continue
		}
		out = append(out, command.TranslationInput{
			Language:    t.Language,
			Label:       t.Label,
			Description: t.Description,
			Note:        t.Note,
		})
	}
	return out
}
