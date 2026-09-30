package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/role"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type ListRolesQuery struct {
	Search     *string
	RoleStatus *string
	RoleIDs    []int64
	Page       int64
	Limit      int64
}

type ListRolesQueryHandler struct {
	roleRepo role.IRepository
}

func NewListRolesQueryHandler(roleRepo role.IRepository) *ListRolesQueryHandler {
	return &ListRolesQueryHandler{roleRepo: roleRepo}
}

func (h *ListRolesQueryHandler) Handle(ctx context.Context, q ListRolesQuery) ([]*role.Role, *pagination.Pagination, error) {
	return h.roleRepo.ListRoles(ctx, &role.ListRolesParams{
		Search:     q.Search,
		RoleStatus: q.RoleStatus,
		RoleIds:    q.RoleIDs,
		Page:       q.Page,
		Limit:      q.Limit,
	})
}
