package program

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/program"
	dto "math-ai.com/math-ai/internal/application/dto/program"
	query "math-ai.com/math-ai/internal/application/query/program"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/program"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
)

const imageUrlTTL = 1 * time.Hour

// Service is the program module's public façade. It composes the
// CQRS handlers behind the validators and owns no I/O of its own —
// every write goes through the UoW-backed commands, every read through
// the repo-bound queries.
type Service struct {
	listProgramsQuery     *query.ListProgramsQueryHandler
	getProgramQuery       *query.GetProgramByIdQueryHandler
	createProgramCmd      *command.CreateProgramCommandHandler
	updateProgramCmd      *command.UpdateProgramCommandHandler
	softDeleteProgramCmd  *command.SoftDeleteProgramCommandHandler
	forceDeleteProgramCmd *command.ForceDeleteProgramCommandHandler
	storageProvider       *storage.Adapter
}

// NewService wires the program module. storageProvider may be nil; in
// that case responses simply omit image_url.
func NewService(
	programRepo domain.IRepository,
	translationRepo domain.ITranslationRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		listProgramsQuery:     query.NewListProgramsQueryHandler(programRepo),
		getProgramQuery:       query.NewGetProgramByIdQueryHandler(programRepo, translationRepo),
		createProgramCmd:      command.NewCreateProgramCommandHandler(uow),
		updateProgramCmd:      command.NewUpdateProgramCommandHandler(uow),
		softDeleteProgramCmd:  command.NewSoftDeleteProgramCommandHandler(uow),
		forceDeleteProgramCmd: command.NewForceDeleteProgramCommandHandler(uow),
		storageProvider:       storageProvider,
	}
}

func (s *Service) CreateProgram(ctx context.Context, req *dto.CreateProgramReq) (*dto.CreateProgramRes, error) {
	if err := ValidateCreateProgram(ctx, req); err != nil {
		return nil, err
	}

	created, err := s.createProgramCmd.Handle(ctx, command.CreateProgramCommand{
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
	return &dto.CreateProgramRes{Program: resp}, nil
}

func (s *Service) UpdateProgram(ctx context.Context, req *dto.UpdateProgramReq) (*dto.UpdateProgramRes, error) {
	if err := ValidateUpdateProgram(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.updateProgramCmd.Handle(ctx, command.UpdateProgramCommand{
		ProgramID:    req.ProgramID,
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
	return &dto.UpdateProgramRes{Program: resp}, nil
}

func (s *Service) SoftDeleteProgram(ctx context.Context, req *dto.DeleteProgramReq) (*dto.DeleteProgramRes, error) {
	if err := ValidateDeleteProgram(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteProgramCmd.Handle(ctx, command.SoftDeleteProgramCommand{ProgramID: req.ProgramID}); err != nil {
		return nil, err
	}
	return &dto.DeleteProgramRes{}, nil
}

func (s *Service) ForceDeleteProgram(ctx context.Context, req *dto.DeleteProgramReq) (*dto.DeleteProgramRes, error) {
	if err := ValidateDeleteProgram(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteProgramCmd.Handle(ctx, command.ForceDeleteProgramCommand{ProgramID: req.ProgramID}); err != nil {
		return nil, err
	}
	return &dto.DeleteProgramRes{}, nil
}

func (s *Service) GetProgram(ctx context.Context, req *dto.GetProgramReq) (*dto.GetProgramRes, error) {
	if err := ValidateGetProgram(ctx, req); err != nil {
		return nil, err
	}
	p, err := s.getProgramQuery.Handle(ctx, query.GetProgramByIdQuery{
		ProgramID: req.ProgramID,
		Language:  req.Language,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROGRAM_NOT_FOUND, nil,
			ErrProgramNotFound)
	}
	resp := dto.DomainToResponse(p)
	s.populateImageUrl(ctx, resp)
	return &dto.GetProgramRes{Program: resp}, nil
}

func (s *Service) ListPrograms(ctx context.Context, req *dto.ListProgramsReq) (*dto.ListProgramsRes, error) {
	if err := ValidateListPrograms(ctx, req); err != nil {
		return nil, err
	}

	language := req.Language
	if language == "" {
		language = metadata.GetClientLanguage(ctx).ToEnumLanguage()
	}

	programs, pg, err := s.listProgramsQuery.Handle(ctx, &query.ListProgramsQuery{
		Language: language,
		Page:     req.Page,
		Limit:    req.Size,
	})
	if err != nil {
		return nil, err
	}

	responses := dto.DomainListToResponse(programs)
	for _, r := range responses {
		s.populateImageUrl(ctx, r)
	}
	return &dto.ListProgramsRes{
		Programs:   responses,
		Pagination: pg,
	}, nil
}

func (s *Service) populateImageUrl(ctx context.Context, resp *dto.ProgramResponse) {
	if resp == nil || s.storageProvider == nil || resp.ImageKey == nil || *resp.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.ImageKey,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("program.image presign failed program_id=%s err=%v", resp.ProgramID, err)
		return
	}
	resp.ImageUrl = &url
}

func translationInputsFromDTO(in []*dto.ProgramTranslationDTO) []command.TranslationInput {
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
