package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type permissionRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewPermissionRepository(database db.IDatabase) di.IPermissionRepository {
	return &permissionRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanPermission is a reusable helper method to scan permission data from a row
func (r *permissionRepository) scanPermission(scanner Scanner) (*domain.Permission, error) {
	var m models.PermissionModel
	err := scanner.Scan(
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

// List retrieves a paginated list of permissions
func (r *permissionRepository) List(ctx context.Context, params di.ListPermissionsParams) ([]*domain.Permission, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{}
	countArgs := []interface{}{}

	// Base query
	queryBuilder.WriteString(queries.PermissionBaseSelect)
	countBuilder.WriteString(queries.PermissionBaseCount)

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
		permission, err := r.scanPermission(rows)
		if err != nil {
			return nil, nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, paginationObj, nil
}

// FindByID retrieves a permission by ID
func (r *permissionRepository) FindByID(ctx context.Context, id string) (*domain.Permission, error) {
	row := r.db.QueryRow(ctx, nil, queries.PermissionFindByID, id)
	return r.scanPermission(row)
}

// FindByName retrieves a permission by name
func (r *permissionRepository) FindByName(ctx context.Context, name string) (*domain.Permission, error) {
	row := r.db.QueryRow(ctx, nil, queries.PermissionFindByName, name)
	return r.scanPermission(row)
}

// FindByEndpoint retrieves a permission by HTTP method and endpoint path
func (r *permissionRepository) FindByEndpoint(ctx context.Context, httpMethod, endpointPath string) (*domain.Permission, error) {
	row := r.db.QueryRow(ctx, nil, queries.PermissionFindByEndpoint, httpMethod, endpointPath)
	return r.scanPermission(row)
}

// FindByResource retrieves all permissions for a resource
func (r *permissionRepository) FindByResource(ctx context.Context, resource string) ([]*domain.Permission, error) {
	rows, err := r.db.Query(ctx, nil, queries.PermissionFindByResource, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by resource: %v", err)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission, err := r.scanPermission(rows)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// FindByRoleID retrieves all permissions assigned to a role (including inherited)
func (r *permissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	rows, err := r.db.Query(ctx, nil, queries.PermissionFindByRoleID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by role: %v", err)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission, err := r.scanPermission(rows)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// Create inserts a new permission
func (r *permissionRepository) Create(ctx context.Context, tx *sql.Tx, permission *domain.Permission) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.PermissionInsert,
		permission.ID(),
		permission.Name(),
		permission.Description(),
		permission.HTTPMethod(),
		permission.EndpointPath(),
		permission.Resource(),
		permission.Action(),
		enum.StatusActive,
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create permission: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing permission
func (r *permissionRepository) Update(ctx context.Context, tx *sql.Tx, permission *domain.Permission) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.PermissionUpdate,
		PrepareForUpdate(permission.Name()),
		PrepareForUpdate(permission.Description()),
		PrepareForUpdate(permission.HTTPMethod()),
		PrepareForUpdate(permission.EndpointPath()),
		PrepareForUpdate(permission.Resource()),
		PrepareForUpdate(permission.Action()),
		PrepareForUpdate(permission.Status()),
		mathtime.Now(),
		permission.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update permission: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.PermissionSoftDelete, mathtime.Now(), mathtime.Now(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete permission: %v", err)
	}

	return result.RowsAffected()
}

// ForceDelete permanently deletes a permission
func (r *permissionRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.PermissionForceDelete, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete permission: %v", err)
	}

	return result.RowsAffected()
}
