package semester

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/semester"
	query "math-ai.com/math-ai/internal/application/query/semester"
	domain "math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

const imageUrlTTL = 1 * time.Hour

type Service struct {
	listSemestersQuery *query.ListSemestersQueryHandler
	storageProvider    *storage.Adapter
}

func NewService(repo domain.IRepository, storageProvider *storage.Adapter) *Service {
	return &Service{
		listSemestersQuery: query.NewListSemestersQueryHandler(repo),
		storageProvider:    storageProvider,
	}
}

func (s *Service) ListSemesters(ctx context.Context, req *dto.ListSemestersReq) (*dto.ListSemestersRes, error) {
	if err := ValidateListSemesters(ctx, req); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	semesters, pg, err := s.listSemestersQuery.Handle(ctx, &query.ListSemestersQuery{
		Language: lang,
		Page:     req.Page,
		Limit:    req.Size,
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
		logger.From(ctx).Warnf("semester.image presign failed semester_id=%s err=%v", resp.SemesterID, err)
		return
	}
	resp.ImageUrl = &url
}
