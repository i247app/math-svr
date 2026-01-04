package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/role_permission"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/db"
)

type rolePermissionRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewRolePermissionRepository(database db.IDatabase) di.IRolePermissionRepository {
	return &rolePermissionRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanRolePermission is a reusable helper method to scan role_permission data from a row
func (r *rolePermissionRepository) scanRolePermission(scanner Scanner) (*domain.RolePermission, error) {
	var m models.RolePermissionModel
	err := scanner.Scan(
		&m.ID, &m.RoleID, &m.PermissionID, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildRolePermissionDomainFromModel(&m), nil
}

// FindByRoleID retrieves all role-permission mappings for a role
func (r *rolePermissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]*domain.RolePermission, error) {
	rows, err := r.db.Query(ctx, nil, queries.RolePermissionFindByRoleID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find role permissions: %v", err)
	}
	defer rows.Close()

	var rolePermissions []*domain.RolePermission
	for rows.Next() {
		rp, err := r.scanRolePermission(rows)
		if err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rp)
	}

	return rolePermissions, nil
}

// FindByPermissionID retrieves all role-permission mappings for a permission
func (r *rolePermissionRepository) FindByPermissionID(ctx context.Context, permissionID string) ([]*domain.RolePermission, error) {
	rows, err := r.db.Query(ctx, nil, queries.RolePermissionFindByPermissionID, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find role permissions: %v", err)
	}
	defer rows.Close()

	var rolePermissions []*domain.RolePermission
	for rows.Next() {
		rp, err := r.scanRolePermission(rows)
		if err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rp)
	}

	return rolePermissions, nil
}

// AssignPermission assigns a permission to a role
func (r *rolePermissionRepository) AssignPermission(ctx context.Context, tx *sql.Tx, rolePermission *domain.RolePermission) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.RolePermissionInsert,
		rolePermission.ID(),
		rolePermission.RoleID(),
		rolePermission.PermissionID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to assign permission: %v", err)
	}

	return result.RowsAffected()
}

// RevokePermission removes a permission from a role
func (r *rolePermissionRepository) RevokePermission(ctx context.Context, tx *sql.Tx, roleID, permissionID string) error {
	_, err := r.db.Exec(ctx, tx, queries.RolePermissionRevoke, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %v", err)
	}

	return nil
}

// RevokeAllPermissions removes all permissions from a role
func (r *rolePermissionRepository) RevokeAllPermissions(ctx context.Context, tx *sql.Tx, roleID string) error {
	_, err := r.db.Exec(ctx, tx, queries.RolePermissionRevokeAll, roleID)
	if err != nil {
		return fmt.Errorf("failed to revoke all permissions: %v", err)
	}

	return nil
}

// BulkAssignPermissions assigns multiple permissions to a role
func (r *rolePermissionRepository) BulkAssignPermissions(ctx context.Context, tx *sql.Tx, rolePermissions []*domain.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}

	// Build bulk insert query
	query := `INSERT INTO role_permissions (id, role_id, permission_id) VALUES `
	values := []interface{}{}
	placeholders := []string{}

	for _, rp := range rolePermissions {
		placeholders = append(placeholders, "(?, ?, ?)")
		values = append(values, rp.ID(), rp.RoleID(), rp.PermissionID())
	}

	query += fmt.Sprintf("%s", joinPlaceholders(placeholders, ", "))

	_, err := r.db.Exec(ctx, tx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to bulk assign permissions: %v", err)
	}

	return nil
}

// Helper function to join placeholders
func joinPlaceholders(placeholders []string, sep string) string {
	result := ""
	for i, p := range placeholders {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
