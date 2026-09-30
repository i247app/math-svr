package role

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ErrDuplicateRoleCode is returned by IRepository.Create when another live
// row already holds the role_code (uk_live_role_code). The command layer
// checks first, so this only surfaces when two creates race.
var ErrDuplicateRoleCode = errors.New("role: role_code already exists")

// ListRolesParams narrows the listing query. Search matches role_code and
// role_name case-insensitively; RoleStatus narrows by exact match; RoleIds
// restricts the result to the supplied ids (nil or empty = no filter).
type ListRolesParams struct {
	Search     *string
	RoleStatus *string
	RoleIds    []int64
	Page       int64
	Limit      int64
}

// IRepository owns ma_roles persistence.
type IRepository interface {
	FindByRoleId(ctx context.Context, roleId int64) (*Role, error)
	FindByRoleCode(ctx context.Context, roleCode string) (*Role, error)
	ListRoles(ctx context.Context, params *ListRolesParams) ([]*Role, *pagination.Pagination, error)
	Create(ctx context.Context, r *Role) (*Role, error)
	Update(ctx context.Context, r *Role) error
	// ClearRoleImageKey sets role_image_key to NULL — the one column Update
	// cannot clear, since a nil field there means "leave unchanged".
	ClearRoleImageKey(ctx context.Context, roleId int64) error
	SoftDeleteByRoleId(ctx context.Context, roleId int64) error
	ForceDeleteByRoleId(ctx context.Context, roleId int64) error
}
