package di

import (
	"context"

	domain "math-ai.com/math-ai/internal/core/domain/role"
	permDomain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListRolesServiceParams struct {
	Search        string
	Page          int64
	Limit         int64
	OrderBy       string
	OrderDesc     bool
	TakeAll       bool
	IncludeSystem bool
}

type IRoleService interface {
	// List retrieves a paginated list of roles
	List(ctx context.Context, params ListRolesServiceParams) ([]*domain.Role, *pagination.Pagination, error)

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id string) (*domain.Role, error)

	// GetByCode retrieves a role by code
	GetByCode(ctx context.Context, code string) (*domain.Role, error)

	// GetPermissions retrieves all permissions for a role (including inherited)
	GetPermissions(ctx context.Context, roleID string) ([]*permDomain.Permission, error)

	// Create creates a new role
	Create(ctx context.Context, role *domain.Role) (*domain.Role, error)

	// Update updates an existing role
	Update(ctx context.Context, role *domain.Role) (*domain.Role, error)

	// Delete soft deletes a role
	Delete(ctx context.Context, id string) error

	// ForceDelete permanently deletes a role
	ForceDelete(ctx context.Context, id string) error

	// AssignPermissions assigns multiple permissions to a role
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error

	// RevokePermissions removes multiple permissions from a role
	RevokePermissions(ctx context.Context, roleID string, permissionIDs []string) error
}
