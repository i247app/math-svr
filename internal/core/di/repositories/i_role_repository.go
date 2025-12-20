package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/role"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListRolesParams struct {
	Search        string
	Page          int64
	Limit         int64
	OrderBy       string
	OrderDesc     bool
	TakeAll       bool
	IncludeSystem bool // Include system roles in results
}

type IRoleRepository interface {
	// List retrieves a paginated list of roles
	List(ctx context.Context, params ListRolesParams) ([]*domain.Role, *pagination.Pagination, error)

	// FindByID retrieves a role by ID
	FindByID(ctx context.Context, id string) (*domain.Role, error)

	// FindByCode retrieves a role by code
	FindByCode(ctx context.Context, code string) (*domain.Role, error)

	// GetWithPermissions retrieves a role with its permissions
	GetWithPermissions(ctx context.Context, roleID string) (*domain.Role, error)

	// GetHierarchy retrieves all parent roles up the hierarchy
	GetHierarchy(ctx context.Context, roleID string) ([]*domain.Role, error)

	// Create inserts a new role
	Create(ctx context.Context, tx *sql.Tx, role *domain.Role) (int64, error)

	// Update modifies an existing role
	Update(ctx context.Context, role *domain.Role) (int64, error)

	// Delete soft deletes a role
	Delete(ctx context.Context, id string) error

	// ForceDelete permanently deletes a role
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
}
