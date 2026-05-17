package semester

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/semester"
	query "math-ai.com/math-ai/internal/application/query/semester"
	domain "math-ai.com/math-ai/internal/domain/semester"
	"math-ai.com/math-ai/internal/shared/enum"
)

type Service struct {
	listSemestersQuery *query.ListSemestersQueryHandler
}

func NewService(repo domain.IRepository) *Service {
	return &Service{
		listSemestersQuery: query.NewListSemestersQueryHandler(repo),
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

	return &dto.ListSemestersRes{
		Semesters:  dto.DomainListToResponse(semesters),
		Pagination: pg,
	}, nil
}
