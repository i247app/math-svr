package di

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/role_permission"
)

type IRolePermissionRepository interface {
	// FindByRoleID retrieves all role-permission mappings for a role
	FindByRoleID(ctx context.Context, roleID string) ([]*domain.RolePermission, error)

	// FindByPermissionID retrieves all role-permission mappings for a permission
	FindByPermissionID(ctx context.Context, permissionID string) ([]*domain.RolePermission, error)

	// AssignPermission assigns a permission to a role
	AssignPermission(ctx context.Context, tx *sql.Tx, rolePermission *domain.RolePermission) (int64, error)

	// RevokePermission removes a permission from a role
	RevokePermission(ctx context.Context, tx *sql.Tx, roleID, permissionID string) error

	// RevokeAllPermissions removes all permissions from a role
	RevokeAllPermissions(ctx context.Context, tx *sql.Tx, roleID string) error

	// BulkAssignPermissions assigns multiple permissions to a role
	BulkAssignPermissions(ctx context.Context, tx *sql.Tx, rolePermissions []*domain.RolePermission) error
}
