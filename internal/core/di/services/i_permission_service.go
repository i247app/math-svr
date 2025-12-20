package di

import (
	"context"

	domain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListPermissionsServiceParams struct {
	Search    string
	Resource  string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type IPermissionService interface {
	// List retrieves a paginated list of permissions
	List(ctx context.Context, params ListPermissionsServiceParams) ([]*domain.Permission, *pagination.Pagination, error)

	// GetByID retrieves a permission by ID
	GetByID(ctx context.Context, id string) (*domain.Permission, error)

	// GetByName retrieves a permission by name
	GetByName(ctx context.Context, name string) (*domain.Permission, error)

	// GetByResource retrieves all permissions for a resource
	GetByResource(ctx context.Context, resource string) ([]*domain.Permission, error)

	// Create creates a new permission
	Create(ctx context.Context, permission *domain.Permission) (*domain.Permission, error)

	// Update updates an existing permission
	Update(ctx context.Context, permission *domain.Permission) (*domain.Permission, error)

	// Delete soft deletes a permission
	Delete(ctx context.Context, id string) error

	// ForceDelete permanently deletes a permission
	ForceDelete(ctx context.Context, id string) error
}
