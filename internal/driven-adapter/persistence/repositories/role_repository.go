package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/role"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type roleRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewRoleRepository(database db.IDatabase) di.IRoleRepository {
	return &roleRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanRole is a reusable helper method to scan role data from a row
func (r *roleRepository) scanRole(scanner Scanner) (*domain.Role, error) {
	var m models.RoleModel
	err := scanner.Scan(
		&m.ID, &m.Name, &m.Code, &m.Description, &m.ParentRoleID, &m.IsSystemRole,
		&m.Status, &m.DisplayOrder, &m.CreateID, &m.CreateDT, &m.ModifyID, &m.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildRoleDomainFromModel(&m), nil
}

// List retrieves a paginated list of roles
func (r *roleRepository) List(ctx context.Context, params di.ListRolesParams) ([]*domain.Role, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{}
	countArgs := []interface{}{}

	// Base query
	queryBuilder.WriteString(queries.RoleBaseSelect)
	countBuilder.WriteString(queries.RoleBaseCount)

	// Filter system roles if not included
	if !params.IncludeSystem {
		condition := ` AND is_system_role = FALSE`
		queryBuilder.WriteString(condition)
		countBuilder.WriteString(condition)
	}

	// Add search condition
	if params.Search != "" {
		searchCondition := ` AND (name LIKE ? OR code LIKE ? OR description LIKE ?)`
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
		return nil, nil, fmt.Errorf("failed to count roles: %v", err)
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
		queryBuilder.WriteString(" ORDER BY display_order ASC, name ASC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, paginationObj.Size, paginationObj.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list roles: %v", err)
	}
	defer rows.Close()

	// Scan results
	var roles []*domain.Role
	for rows.Next() {
		role, err := r.scanRole(rows)
		if err != nil {
			return nil, nil, err
		}
		roles = append(roles, role)
	}

	return roles, paginationObj, nil
}

// FindByID retrieves a role by ID
func (r *roleRepository) FindByID(ctx context.Context, id string) (*domain.Role, error) {
	row := r.db.QueryRow(ctx, nil, queries.RoleFindByID, id)
	return r.scanRole(row)
}

// FindByCode retrieves a role by code
func (r *roleRepository) FindByCode(ctx context.Context, code string) (*domain.Role, error) {
	row := r.db.QueryRow(ctx, nil, queries.RoleFindByCode, code)
	return r.scanRole(row)
}

// GetWithPermissions retrieves a role with its permissions (not yet implemented fully)
func (r *roleRepository) GetWithPermissions(ctx context.Context, roleID string) (*domain.Role, error) {
	// This is a simplified version - just get the role
	return r.FindByID(ctx, roleID)
}

// GetHierarchy retrieves all parent roles up the hierarchy
func (r *roleRepository) GetHierarchy(ctx context.Context, roleID string) ([]*domain.Role, error) {
	var roles []*domain.Role

	currentID := roleID
	visited := make(map[string]bool) // Prevent infinite loops

	for currentID != "" {
		if visited[currentID] {
			break // Circular reference detected
		}
		visited[currentID] = true

		role, err := r.FindByID(ctx, currentID)
		if err != nil {
			return nil, err
		}
		if role == nil {
			break
		}

		roles = append(roles, role)

		// Move to parent
		if role.ParentRoleID() != nil {
			currentID = *role.ParentRoleID()
		} else {
			currentID = ""
		}
	}

	return roles, nil
}

// Create inserts a new role
func (r *roleRepository) Create(ctx context.Context, tx *sql.Tx, role *domain.Role) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.RoleInsert,
		role.ID(),
		role.Name(),
		role.Code(),
		role.Description(),
		role.ParentRoleID(),
		role.IsSystemRole(),
		enum.StatusActive,
		role.DisplayOrder(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create role: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing role
func (r *roleRepository) Update(ctx context.Context, tx *sql.Tx, role *domain.Role) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.RoleUpdate,
		PrepareForUpdate(role.Name()),
		PrepareForUpdate(role.Code()),
		PrepareForUpdate(role.Description()),
		PrepareForUpdate(role.ParentRoleID()),
		PrepareForUpdate(role.Status()),
		PrepareForUpdate(role.DisplayOrder()),
		mathtime.Now(),
		role.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update role: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a role
func (r *roleRepository) Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	now := mathtime.Now()
	result, err := r.db.Exec(ctx, tx, queries.RoleSoftDelete, now, now, id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete role: %v", err)
	}

	return result.RowsAffected()
}

// ForceDelete permanently deletes a role
func (r *roleRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.RoleForceDelete, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete role: %v", err)
	}

	return result.RowsAffected()
}
