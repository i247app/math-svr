package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListPermissionsParams struct {
	Search    string
	Resource  string // Filter by resource
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type IPermissionRepository interface {
	// List retrieves a paginated list of permissions
	List(ctx context.Context, params ListPermissionsParams) ([]*domain.Permission, *pagination.Pagination, error)

	// FindByID retrieves a permission by ID
	FindByID(ctx context.Context, id string) (*domain.Permission, error)

	// FindByName retrieves a permission by name
	FindByName(ctx context.Context, name string) (*domain.Permission, error)

	// FindByEndpoint retrieves a permission by HTTP method and endpoint path
	FindByEndpoint(ctx context.Context, httpMethod, endpointPath string) (*domain.Permission, error)

	// FindByResource retrieves all permissions for a resource
	FindByResource(ctx context.Context, resource string) ([]*domain.Permission, error)

	// FindByRoleID retrieves all permissions assigned to a role (including inherited)
	FindByRoleID(ctx context.Context, roleID string) ([]*domain.Permission, error)

	// Create inserts a new permission
	Create(ctx context.Context, tx *sql.Tx, permission *domain.Permission) (int64, error)

	// Update modifies an existing permission
	Update(ctx context.Context, permission *domain.Permission) (int64, error)

	// Delete soft deletes a permission
	Delete(ctx context.Context, id string) error

	// ForceDelete permanently deletes a permission
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error
}
