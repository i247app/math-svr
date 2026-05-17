package program

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/program"
	query "math-ai.com/math-ai/internal/application/query/program"
	domain "math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/shared/enum"
)

type Service struct {
	listProgramsQuery *query.ListProgramsQueryHandler
}

func NewService(repo domain.IRepository) *Service {
	return &Service{
		listProgramsQuery: query.NewListProgramsQueryHandler(repo),
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

	return &dto.ListProgramsRes{
		Programs:   dto.DomainListToResponse(programs),
		Pagination: pg,
	}, nil
}
