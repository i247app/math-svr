package chapter

import (
	"context"
	"errors"

	command "math-ai.com/math-ai/internal/application/command/chapter"
	dto "math-ai.com/math-ai/internal/application/dto/chapter"
	query "math-ai.com/math-ai/internal/application/query/chapter"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/chapter"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Service is the chapter module's public façade. It composes the
// CQRS handlers behind the validators and owns no I/O of its own —
// every write goes through the UoW-backed commands, every read through
// the repo-bound queries.
type Service struct {
	getChapterQuery       *query.GetChapterByIdQueryHandler
	listChaptersQuery     *query.ListChaptersQueryHandler
	createChapterCmd      *command.CreateChapterCommandHandler
	updateChapterCmd      *command.UpdateChapterCommandHandler
	softDeleteChapterCmd  *command.SoftDeleteChapterCommandHandler
	forceDeleteChapterCmd *command.ForceDeleteChapterCommandHandler
}

func NewService(
	chapterRepo domain.IRepository,
	translationRepo domain.ITranslationRepository,
	uow transaction.UnitOfWork,
) *Service {
	return &Service{
		getChapterQuery:       query.NewGetChapterByIdQueryHandler(chapterRepo, translationRepo),
		listChaptersQuery:     query.NewListChaptersQueryHandler(chapterRepo),
		createChapterCmd:      command.NewCreateChapterCommandHandler(uow),
		updateChapterCmd:      command.NewUpdateChapterCommandHandler(uow),
		softDeleteChapterCmd:  command.NewSoftDeleteChapterCommandHandler(uow),
		forceDeleteChapterCmd: command.NewForceDeleteChapterCommandHandler(uow),
	}
}

func (s *Service) CreateChapter(ctx context.Context, req *dto.CreateChapterReq) (*dto.CreateChapterRes, error) {
	if err := ValidateCreateChapter(ctx, req); err != nil {
		return nil, err
	}

	created, err := s.createChapterCmd.Handle(ctx, command.CreateChapterCommand{
		ProgramID:    req.ProgramID,
		GradeID:      req.GradeID,
		SemesterID:   req.SemesterID,
		Label:        req.Label,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		Note:         req.Note,
		Translations: translationInputsFromDTO(req.Translations),
	})
	if err != nil {
		return nil, err
	}
	return &dto.CreateChapterRes{Chapter: dto.DomainToResponse(created)}, nil
}

func (s *Service) UpdateChapter(ctx context.Context, req *dto.UpdateChapterReq) (*dto.UpdateChapterRes, error) {
	if err := ValidateUpdateChapter(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.updateChapterCmd.Handle(ctx, command.UpdateChapterCommand{
		ChapterID:    req.ChapterID,
		ProgramID:    req.ProgramID,
		GradeID:      req.GradeID,
		SemesterID:   req.SemesterID,
		Label:        req.Label,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		Note:         req.Note,
		Translations: translationInputsFromDTO(req.Translations),
	})
	if err != nil {
		return nil, err
	}
	return &dto.UpdateChapterRes{Chapter: dto.DomainToResponse(updated)}, nil
}

func (s *Service) SoftDeleteChapter(ctx context.Context, req *dto.DeleteChapterReq) (*dto.DeleteChapterRes, error) {
	if err := ValidateDeleteChapter(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteChapterCmd.Handle(ctx, command.SoftDeleteChapterCommand{ChapterID: req.ChapterID}); err != nil {
		return nil, err
	}
	return &dto.DeleteChapterRes{}, nil
}

func (s *Service) ForceDeleteChapter(ctx context.Context, req *dto.DeleteChapterReq) (*dto.DeleteChapterRes, error) {
	if err := ValidateDeleteChapter(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteChapterCmd.Handle(ctx, command.ForceDeleteChapterCommand{ChapterID: req.ChapterID}); err != nil {
		return nil, err
	}
	return &dto.DeleteChapterRes{}, nil
}

func (s *Service) GetChapter(ctx context.Context, req *dto.GetChapterReq) (*dto.GetChapterRes, error) {
	if err := ValidateGetChapter(ctx, req); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	c, err := s.getChapterQuery.Handle(ctx, query.GetChapterByIdQuery{
		ChapterID: req.ChapterID,
		Language:  lang,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if c == nil {
		return nil, errs.NewError(ctx, status.CHAPTER_NOT_FOUND, nil,
			errors.New("chapter not found"))
	}
	return &dto.GetChapterRes{Chapter: dto.DomainToResponse(c)}, nil
}

func (s *Service) ListChapters(ctx context.Context, req *dto.ListChaptersReq) (*dto.ListChaptersRes, error) {
	if err := ValidateListChapters(ctx, req); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = enum.LanguageTypeEnglish
	}

	chapters, pg, err := s.listChaptersQuery.Handle(ctx, query.ListChaptersQuery{
		ProgramID:  req.ProgramID,
		GradeID:    req.GradeID,
		SemesterID: req.SemesterID,
		Language:   lang,
		Page:       req.Page,
		Limit:      req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &dto.ListChaptersRes{
		Chapters:   dto.DomainListToResponse(chapters),
		Pagination: pg,
	}, nil
}

func translationInputsFromDTO(in []*dto.ChapterTranslationDTO) []command.TranslationInput {
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
