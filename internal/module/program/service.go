package program

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/program"
	query "math-ai.com/math-ai/internal/application/query/program"
	domain "math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

const imageUrlTTL = 1 * time.Hour

type Service struct {
	listProgramsQuery *query.ListProgramsQueryHandler
	storageProvider   *storage.Adapter
}

// NewService wires the program module. storageProvider may be nil; in that
// case responses simply omit image_url.
func NewService(repo domain.IRepository, storageProvider *storage.Adapter) *Service {
	return &Service{
		listProgramsQuery: query.NewListProgramsQueryHandler(repo),
		storageProvider:   storageProvider,
	}
}

func (s *Service) ListPrograms(ctx context.Context, req *dto.ListProgramsReq) (*dto.ListProgramsRes, error) {
	if err := ValidateListPrograms(ctx, req); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	programs, pg, err := s.listProgramsQuery.Handle(ctx, &query.ListProgramsQuery{
		Language: lang,
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
