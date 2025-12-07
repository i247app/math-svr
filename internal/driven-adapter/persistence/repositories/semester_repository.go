package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/semester"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type semesterRepository struct {
	db db.IDatabase
}

func NewSemesterRepository(db db.IDatabase) di.ISemesterRepository {
	return &semesterRepository{
		db: db,
	}
}

// List retrieves a paginated list of semesters with optional search and sorting.
func (r *semesterRepository) List(ctx context.Context, params di.ListSemestersParams) ([]*domain.Semester, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	// Base query
	queryBuilder.WriteString(`
		SELECT id, name, description, iamge_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM semesters WHERE deleted_dt IS NULL
	`)

	// Add search condition
	if params.Search != "" {
		queryBuilder.WriteString(` AND (name LIKE ? OR description LIKE ?)`)
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM semesters WHERE deleted_dt IS NULL"
	if params.Search != "" {
		countQuery += ` AND (name LIKE ? OR description LIKE ?)`
	}
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countQuery, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count semesters: %v", err)
	}

	// Initialize pagination
	pagination := pagination.NewPagination(params.Page, params.Limit, total)
	if params.TakeAll {
		pagination.Size = total
		pagination.Skip = 0
		pagination.Page = 1
		pagination.TotalPages = 1
	}

	// Add sorting
	if params.OrderBy != "" {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s", params.OrderBy))
		if params.OrderDesc {
			queryBuilder.WriteString(" DESC")
		}
	} else {
		queryBuilder.WriteString(" ORDER BY display_order ASC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, pagination.Size, pagination.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list semesters: %v", err)
	}
	defer rows.Close()

	// Scan results
	var semesters []*domain.Semester
	for rows.Next() {
		var s models.SemesterModel
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.ImageKey, &s.Status, &s.DisplayOrder,
			&s.CreateID, &s.CreateDT, &s.ModifyID, &s.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		semesters = append(semesters, domain.BuildSemesterDomainFromModel(&s))
	}

	return semesters, pagination, nil
}

// FindByID retrieves a semester by ID.
func (r *semesterRepository) FindByID(ctx context.Context, id string) (*domain.Semester, error) {
	query := `
		SELECT id, name, description, iamge_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM semesters
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var s models.SemesterModel
	err := result.Scan(
		&s.ID, &s.Name, &s.Description, &s.ImageKey, &s.Status, &s.DisplayOrder,
		&s.CreateID, &s.CreateDT, &s.ModifyID, &s.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	semester := domain.BuildSemesterDomainFromModel(&s)

	return semester, nil
}

// FindByName retrieves a semester by name.
func (r *semesterRepository) FindByName(ctx context.Context, name string) (*domain.Semester, error) {
	query := `
		SELECT id, name, description, iamge_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM semesters
		WHERE name = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, name)

	var s models.SemesterModel
	err := result.Scan(
		&s.ID, &s.Name, &s.Description, &s.ImageKey, &s.Status, &s.DisplayOrder,
		&s.CreateID, &s.CreateDT, &s.ModifyID, &s.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	semester := domain.BuildSemesterDomainFromModel(&s)

	return semester, nil
}

// Create inserts a new semester into the database.
func (r *semesterRepository) Create(ctx context.Context, tx *sql.Tx, semester *domain.Semester) (int64, error) {
	query := `
		INSERT INTO semesters (id, name, description, iamge_key, status, display_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		semester.ID(),
		semester.Name(),
		semester.Description(),
		semester.ImageKey(),
		enum.StatusActive,
		semester.DisplayOrder(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create semester: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing semester in the database.
func (r *semesterRepository) Update(ctx context.Context, semester *domain.Semester) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE semesters SET ")
	updates := []string{}

	if semester.Name() != "" {
		updates = append(updates, "name = ?")
		args = append(args, semester.Name())
	}

	if semester.Description() != nil {
		updates = append(updates, "description = ?")
		args = append(args, semester.Description())
	}

	// ImageKey can be nil, so we check if it's explicitly set
	if semester.ImageKey() != nil {
		updates = append(updates, "iamge_key = ?")
		args = append(args, semester.ImageKey())
	}

	if semester.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, semester.Status())
	}

	if semester.DisplayOrder() != 0 {
		updates = append(updates, "display_order = ?")
		args = append(args, semester.DisplayOrder())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, semester.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update semester: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a semester by setting deleted_dt.
func (r *semesterRepository) Delete(ctx context.Context, id string) error {
	query := `
			UPDATE semesters
			SET deleted_dt = ?,
				modify_dt = ?
			WHERE id = ? AND deleted_dt IS NULL`
	_, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete semester: %v", err)
	}

	return nil
}

// ForceDelete permanently deletes a semester from the database.
func (r *semesterRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	query := `DELETE FROM semesters WHERE id = ?`
	_, err := r.db.Exec(ctx, tx, query, id)
	if err != nil {
		return fmt.Errorf("failed to force delete semester: %v", err)
	}

	return nil
}
