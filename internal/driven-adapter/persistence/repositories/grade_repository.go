package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/grade"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type gradeRepository struct {
	db db.IDatabase
}

func NewGradeRepository(db db.IDatabase) di.IGradeRepository {
	return &gradeRepository{
		db: db,
	}
}

// List retrieves a paginated list of grades with optional search and sorting.
func (r *gradeRepository) List(ctx context.Context, params di.ListGradesParams) ([]*domain.Grade, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{}
	countArgs := []interface{}{}
	language := appctx.GetLocale(ctx)

	// Base query with LEFT JOIN for translations
	queryBuilder.WriteString(`
		SELECT
			g.id,
			COALESCE(gt.label, g.label) AS label,
			COALESCE(gt.description, g.discription) AS description,
			g.image_key,
			g.status,
			g.display_order,
			g.create_id,
			g.create_dt,
			g.modify_id,
			g.modify_dt
		FROM grades g
		LEFT JOIN grade_translations gt ON g.id = gt.grade_id AND gt.language = ?
		WHERE g.deleted_dt IS NULL`)
	args = append(args, language)

	// Count query base with same JOIN
	countBuilder.WriteString(`
		SELECT COUNT(*)
		FROM grades g
		LEFT JOIN grade_translations gt ON g.id = gt.grade_id AND gt.language = ?
		WHERE g.deleted_dt IS NULL`)
	countArgs = append(countArgs, language)

	// Add search condition to both queries
	if params.Search != "" {
		searchCondition := ` AND (COALESCE(gt.label, g.label) LIKE ? OR COALESCE(gt.description, g.discription) LIKE ?)`
		searchTerm := "%" + params.Search + "%"

		queryBuilder.WriteString(searchCondition)
		args = append(args, searchTerm, searchTerm)

		countBuilder.WriteString(searchCondition)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	// Count total records for pagination
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countBuilder.String(), countArgs...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count grades: %v", err)
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
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY g.%s", params.OrderBy))
		if params.OrderDesc {
			queryBuilder.WriteString(" DESC")
		} else {
			queryBuilder.WriteString(" ASC")
		}
	} else {
		queryBuilder.WriteString(" ORDER BY g.display_order ASC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, paginationObj.Size, paginationObj.Skip)
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
			&g.ID, &g.Label, &g.Description, &g.ImageKey, &g.Status, &g.DisplayOrder,
			&g.CreateID, &g.CreateDT, &g.ModifyID, &g.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		grades = append(grades, domain.BuildGradeDomainFromModel(&g))
	}

	return grades, paginationObj, nil
}

// FindByID retrieves a grade by ID.
func (r *gradeRepository) FindByID(ctx context.Context, id string) (*domain.Grade, error) {
	query := `
		SELECT id, label, discription, image_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM grades
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var g models.GradeModel
	err := result.Scan(
		&g.ID, &g.Label, &g.Description, &g.ImageKey, &g.Status, &g.DisplayOrder,
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
		SELECT id, label, discription, image_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM grades
		WHERE label = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, label)

	var g models.GradeModel
	err := result.Scan(
		&g.ID, &g.Label, &g.Description, &g.ImageKey, &g.Status, &g.DisplayOrder,
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
		INSERT INTO grades (id, label, discription, image_key, status, display_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		grade.ID(),
		grade.Label(),
		grade.Description(),
		grade.ImageKey(),
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

	if grade.Description() != nil {
		updates = append(updates, "discription = ?")
		args = append(args, grade.Description())
	}

	// ImageKey can be nil, so we check if it's explicitly set
	if grade.ImageKey() != nil {
		updates = append(updates, "image_key = ?")
		args = append(args, grade.ImageKey())
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
