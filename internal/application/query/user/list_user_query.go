package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListUsersQuery struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
}

type ListUsersQueryHandler struct {
	userRepo user.IRepository
}

func NewListUsersQueryHandler(userRepo user.IRepository) *ListUsersQueryHandler {
	return &ListUsersQueryHandler{
		userRepo: userRepo,
	}
}

func (h *ListUsersQueryHandler) Handle(ctx context.Context, query *ListUsersQuery) ([]*user.User, *pagination.Pagination, error) {
	params := &user.ListUsersParams{
		Page:  query.Page,
		Limit: query.Limit,
	}

	users, pg, err := h.userRepo.ListUsers(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	return users, pg, nil
}
