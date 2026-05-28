package grade

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/grade"
	query "math-ai.com/math-ai/internal/application/query/grade"
	domain "math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
)

const imageUrlTTL = 1 * time.Hour

type Service struct {
	listGradesQuery *query.ListGradesQueryHandler
	storageProvider *storage.Adapter
}

func NewService(repo domain.IRepository, storageProvider *storage.Adapter) *Service {
	return &Service{
		listGradesQuery: query.NewListGradesQueryHandler(repo),
		storageProvider: storageProvider,
	}
}

func (s *Service) ListGrades(ctx context.Context, req *dto.ListGradesReq) (*dto.ListGradesRes, error) {
	if err := ValidateListGrades(ctx, req); err != nil {
		return nil, err
	}

	language := metadata.GetClientLanguage(ctx).ToEnumLanguage()

	grades, pg, err := s.listGradesQuery.Handle(ctx, &query.ListGradesQuery{
		Language: language,
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
		logger.From(ctx).Warnf("grade.image presign failed grade_id=%s err=%v", resp.GradeID, err)
		return
	}
	resp.ImageUrl = &url
}
