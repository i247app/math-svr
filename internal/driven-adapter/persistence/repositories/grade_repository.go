package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/grade"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type gradeRepository struct {
	db db.IDatabase
}

func NewGradeRepository(db db.IDatabase) repositories.IGradeRepository {
	return &gradeRepository{
		db: db,
	}
}

// List retrieves a paginated list of grades with optional search and sorting.
func (r *gradeRepository) List(ctx context.Context, params repositories.ListGradesParams) ([]*domain.Grade, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	// Base query - note: using 'discription' to match the actual table column
	queryBuilder.WriteString(`
		SELECT id, label, discription, icon_url, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM grades WHERE deleted_dt IS NULL
	`)

	// Add search condition
	if params.Search != "" {
		queryBuilder.WriteString(` AND (label LIKE ? OR discription LIKE ?)`)
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM grades WHERE deleted_dt IS NULL"
	if params.Search != "" {
		countQuery += ` AND (label LIKE ? OR discription LIKE ?)`
	}
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countQuery, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count grades: %v", err)
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
		return nil, nil, fmt.Errorf("failed to list grades: %v", err)
	}
	defer rows.Close()

	// Scan results
	var grades []*domain.Grade
	for rows.Next() {
		var g models.GradeModel
		if err := rows.Scan(
			&g.ID, &g.Label, &g.Description, &g.IconURL, &g.Status, &g.DisplayOrder,
			&g.CreateID, &g.CreateDT, &g.ModifyID, &g.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		grades = append(grades, domain.BuildGradeDomainFromModel(&g))
	}

	return grades, pagination, nil
}

// FindByID retrieves a grade by ID.
func (r *gradeRepository) FindByID(ctx context.Context, id string) (*domain.Grade, error) {
	query := `
		SELECT id, label, discription, icon_url, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM grades
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var g models.GradeModel
	err := result.Scan(
		&g.ID, &g.Label, &g.Description, &g.IconURL, &g.Status, &g.DisplayOrder,
		&g.CreateID, &g.CreateDT, &g.ModifyID, &g.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	grade := domain.BuildGradeDomainFromModel(&g)

	return grade, nil
}

// FindByLabel retrieves a grade by label.
func (r *gradeRepository) FindByLabel(ctx context.Context, label string) (*domain.Grade, error) {
	query := `
		SELECT id, label, discription, icon_url, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM grades
		WHERE label = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, label)

	var g models.GradeModel
	err := result.Scan(
		&g.ID, &g.Label, &g.Description, &g.IconURL, &g.Status, &g.DisplayOrder,
		&g.CreateID, &g.CreateDT, &g.ModifyID, &g.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	grade := domain.BuildGradeDomainFromModel(&g)

	return grade, nil
}

// Create inserts a new grade into the database.
func (r *gradeRepository) Create(ctx context.Context, tx *sql.Tx, grade *domain.Grade) (int64, error) {
	query := `
		INSERT INTO grades (id, label, discription, icon_url, status, display_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		grade.ID(),
		grade.Label(),
		grade.Description(),
		grade.IconURL(),
		enum.StatusActive,
		grade.DisplayOrder(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create grade: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing grade in the database.
func (r *gradeRepository) Update(ctx context.Context, grade *domain.Grade) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE grades SET ")
	updates := []string{}

	if grade.Label() != "" {
		updates = append(updates, "label = ?")
		args = append(args, grade.Label())
	}

	if grade.Description() != "" {
		updates = append(updates, "discription = ?")
		args = append(args, grade.Description())
	}

	// IconURL can be nil, so we check if it's explicitly set
	if grade.IconURL() != nil {
		updates = append(updates, "icon_url = ?")
		args = append(args, grade.IconURL())
	}

	if grade.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, grade.Status())
	}

	if grade.DisplayOrder() != 0 {
		updates = append(updates, "display_order = ?")
		args = append(args, grade.DisplayOrder())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, grade.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update grade: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a grade by setting deleted_dt.
func (r *gradeRepository) Delete(ctx context.Context, id string) error {
	query := `
			UPDATE grades 
			SET deleted_dt = ?,
				modify_dt = ? 
			WHERE id = ? AND deleted_dt IS NULL`
	_, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete grade: %v", err)
	}

	return nil
}

// ForceDelete permanently deletes a grade from the database.
func (r *gradeRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	query := `DELETE FROM grades WHERE id = ?`
	_, err := r.db.Exec(ctx, tx, query, id)
	if err != nil {
		return fmt.Errorf("failed to force delete grade: %v", err)
	}

	return nil
}
