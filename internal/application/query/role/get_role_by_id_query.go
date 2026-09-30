package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/role"
)

type GetRoleByIdQuery struct {
	RoleID int64
}

type GetRoleByIdQueryHandler struct {
	roleRepo role.IRepository
}

func NewGetRoleByIdQueryHandler(roleRepo role.IRepository) *GetRoleByIdQueryHandler {
	return &GetRoleByIdQueryHandler{roleRepo: roleRepo}
}

func (h *GetRoleByIdQueryHandler) Handle(ctx context.Context, q GetRoleByIdQuery) (*role.Role, error) {
	return h.roleRepo.FindByRoleId(ctx, q.RoleID)
}
