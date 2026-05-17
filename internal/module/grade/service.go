package grade

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	query "math-ai.com/math-ai/internal/application/query/grade"
	domain "math-ai.com/math-ai/internal/domain/grade"
	"math-ai.com/math-ai/internal/shared/enum"
)

type Service struct {
	listGradesQuery *query.ListGradesQueryHandler
}

func NewService(repo domain.IRepository) *Service {
	return &Service{
		listGradesQuery: query.NewListGradesQueryHandler(repo),
	}
}

func (s *Service) ListGrades(ctx context.Context, req *dto.ListGradesReq) (*dto.ListGradesRes, error) {
	if err := ValidateListGrades(ctx, req); err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	grades, pg, err := s.listGradesQuery.Handle(ctx, &query.ListGradesQuery{
		Language: lang,
		Page:     req.Page,
		Limit:    req.Size,
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListGradesRes{
		Grades:     dto.DomainListToResponse(grades),
		Pagination: pg,
	}, nil
}
