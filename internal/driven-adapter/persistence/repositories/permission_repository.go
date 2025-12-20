package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type permissionRepository struct {
	db db.IDatabase
}

func NewPermissionRepository(db db.IDatabase) di.IPermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

// List retrieves a paginated list of permissions
func (r *permissionRepository) List(ctx context.Context, params di.ListPermissionsParams) ([]*domain.Permission, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{}
	countArgs := []interface{}{}

	// Base query
	queryBuilder.WriteString(`
		SELECT
			id, name, description, http_method, endpoint_path, resource, action,
			status, create_id, create_dt, modify_id, modify_dt
		FROM permissions
		WHERE deleted_dt IS NULL`)

	countBuilder.WriteString(`
		SELECT COUNT(*)
		FROM permissions
		WHERE deleted_dt IS NULL`)

	// Filter by resource
	if params.Resource != "" {
		condition := ` AND resource = ?`
		queryBuilder.WriteString(condition)
		args = append(args, params.Resource)

		countBuilder.WriteString(condition)
		countArgs = append(countArgs, params.Resource)
	}

	// Add search condition
	if params.Search != "" {
		searchCondition := ` AND (name LIKE ? OR description LIKE ? OR endpoint_path LIKE ?)`
		searchTerm := "%" + params.Search + "%"

		queryBuilder.WriteString(searchCondition)
		args = append(args, searchTerm, searchTerm, searchTerm)

		countBuilder.WriteString(searchCondition)
		countArgs = append(countArgs, searchTerm, searchTerm, searchTerm)
	}

	// Count total records
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countBuilder.String(), countArgs...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count permissions: %v", err)
	}

	// Initialize pagination
	paginationObj := pagination.NewPagination(params.Page, params.Limit, total)
	if params.TakeAll {
		paginationObj.Size = total
		paginationObj.Skip = 0
		paginationObj.Page = 1
		paginationObj.TotalPages = 1
	}

	// Add sorting
	if params.OrderBy != "" {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s", params.OrderBy))
		if params.OrderDesc {
			queryBuilder.WriteString(" DESC")
		} else {
			queryBuilder.WriteString(" ASC")
		}
	} else {
		queryBuilder.WriteString(" ORDER BY resource ASC, name ASC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, paginationObj.Size, paginationObj.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list permissions: %v", err)
	}
	defer rows.Close()

	// Scan results
	var permissions []*domain.Permission
	for rows.Next() {
		var m models.PermissionModel
		if err := rows.Scan(
			&m.ID, &m.Name, &m.Description, &m.HTTPMethod, &m.EndpointPath, &m.Resource, &m.Action,
			&m.Status, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		permissions = append(permissions, domain.BuildPermissionDomainFromModel(&m))
	}

	return permissions, paginationObj, nil
}

// FindByID retrieves a permission by ID
func (r *permissionRepository) FindByID(ctx context.Context, id string) (*domain.Permission, error) {
	query := `
		SELECT id, name, description, http_method, endpoint_path, resource, action,
		       status, create_id, create_dt, modify_id, modify_dt
		FROM permissions
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var m models.PermissionModel
	err := result.Scan(
		&m.ID, &m.Name, &m.Description, &m.HTTPMethod, &m.EndpointPath, &m.Resource, &m.Action,
		&m.Status, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildPermissionDomainFromModel(&m), nil
}

// FindByName retrieves a permission by name
func (r *permissionRepository) FindByName(ctx context.Context, name string) (*domain.Permission, error) {
	query := `
		SELECT id, name, description, http_method, endpoint_path, resource, action,
		       status, create_id, create_dt, modify_id, modify_dt
		FROM permissions
		WHERE name = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, name)

	var m models.PermissionModel
	err := result.Scan(
		&m.ID, &m.Name, &m.Description, &m.HTTPMethod, &m.EndpointPath, &m.Resource, &m.Action,
		&m.Status, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildPermissionDomainFromModel(&m), nil
}

// FindByEndpoint retrieves a permission by HTTP method and endpoint path
func (r *permissionRepository) FindByEndpoint(ctx context.Context, httpMethod, endpointPath string) (*domain.Permission, error) {
	query := `
		SELECT id, name, description, http_method, endpoint_path, resource, action,
		       status, create_id, create_dt, modify_id, modify_dt
		FROM permissions
		WHERE http_method = ? AND endpoint_path = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, httpMethod, endpointPath)

	var m models.PermissionModel
	err := result.Scan(
		&m.ID, &m.Name, &m.Description, &m.HTTPMethod, &m.EndpointPath, &m.Resource, &m.Action,
		&m.Status, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildPermissionDomainFromModel(&m), nil
}

// FindByResource retrieves all permissions for a resource
func (r *permissionRepository) FindByResource(ctx context.Context, resource string) ([]*domain.Permission, error) {
	query := `
		SELECT id, name, description, http_method, endpoint_path, resource, action,
		       status, create_id, create_dt, modify_id, modify_dt
		FROM permissions
		WHERE resource = ? AND deleted_dt IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, nil, query, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by resource: %v", err)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		var m models.PermissionModel
		if err := rows.Scan(
			&m.ID, &m.Name, &m.Description, &m.HTTPMethod, &m.EndpointPath, &m.Resource, &m.Action,
			&m.Status, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
		); err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}

		permissions = append(permissions, domain.BuildPermissionDomainFromModel(&m))
	}

	return permissions, nil
}

// FindByRoleID retrieves all permissions assigned to a role (including inherited)
func (r *permissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.http_method, p.endpoint_path, p.resource, p.action,
		       p.status, p.create_id, p.create_dt, p.modify_id, p.modify_dt
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ? AND p.deleted_dt IS NULL AND rp.deleted_dt IS NULL
		ORDER BY p.resource ASC, p.name ASC
	`

	rows, err := r.db.Query(ctx, nil, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by role: %v", err)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		var m models.PermissionModel
		if err := rows.Scan(
			&m.ID, &m.Name, &m.Description, &m.HTTPMethod, &m.EndpointPath, &m.Resource, &m.Action,
			&m.Status, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
		); err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}

		permissions = append(permissions, domain.BuildPermissionDomainFromModel(&m))
	}

	return permissions, nil
}

// Create inserts a new permission
func (r *permissionRepository) Create(ctx context.Context, tx *sql.Tx, permission *domain.Permission) (int64, error) {
	query := `
		INSERT INTO permissions (id, name, description, http_method, endpoint_path, resource, action, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		permission.ID(),
		permission.Name(),
		permission.Description(),
		permission.HTTPMethod(),
		permission.EndpointPath(),
		permission.Resource(),
		permission.Action(),
		enum.StatusActive,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create permission: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing permission
func (r *permissionRepository) Update(ctx context.Context, permission *domain.Permission) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE permissions SET ")
	updates := []string{}

	if permission.Name() != "" {
		updates = append(updates, "name = ?")
		args = append(args, permission.Name())
	}

	if permission.Description() != nil {
		updates = append(updates, "description = ?")
		args = append(args, permission.Description())
	}

	if permission.HTTPMethod() != "" {
		updates = append(updates, "http_method = ?")
		args = append(args, permission.HTTPMethod())
	}

	if permission.EndpointPath() != "" {
		updates = append(updates, "endpoint_path = ?")
		args = append(args, permission.EndpointPath())
	}

	if permission.Resource() != nil {
		updates = append(updates, "resource = ?")
		args = append(args, permission.Resource())
	}

	if permission.Action() != nil {
		updates = append(updates, "action = ?")
		args = append(args, permission.Action())
	}

	if permission.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, permission.Status())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, mathtime.Now())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, permission.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update permission: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE permissions
		SET deleted_dt = ?,
		    modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL`
	now := mathtime.Now()
	_, err := r.db.Exec(ctx, nil, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %v", err)
	}

	return nil
}

// ForceDelete permanently deletes a permission
func (r *permissionRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	query := `DELETE FROM permissions WHERE id = ?`
	_, err := r.db.Exec(ctx, tx, query, id)
	if err != nil {
		return fmt.Errorf("failed to force delete permission: %v", err)
	}

	return nil
}
