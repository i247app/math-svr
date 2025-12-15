package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/term"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type termRepository struct {
	db db.IDatabase
}

func NewTermRepository(db db.IDatabase) di.ITermRepository {
	return &termRepository{
		db: db,
	}
}

// List retrieves a paginated list of terms with optional search and sorting.
func (r *termRepository) List(ctx context.Context, params di.ListTermsParams) ([]*domain.Term, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	args := []interface{}{}
	countArgs := []interface{}{}
	language := appctx.GetLocale(ctx)

	// Base query with LEFT JOIN for translations
	queryBuilder.WriteString(`
		SELECT
			s.id,
			COALESCE(st.name, s.name) AS name,
			COALESCE(st.description, s.description) AS description,
			s.image_key,
			s.status,
			s.display_order,
			s.create_id,
			s.create_dt,
			s.modify_id,
			s.modify_dt
		FROM terms s
		LEFT JOIN term_translations st ON s.id = st.term_id AND st.language = ?
		WHERE s.deleted_dt IS NULL`)
	args = append(args, language)

	// Count query base with same JOIN
	countBuilder.WriteString(`
		SELECT COUNT(*)
		FROM terms s
		LEFT JOIN term_translations st ON s.id = st.term_id AND st.language = ?
		WHERE s.deleted_dt IS NULL`)
	countArgs = append(countArgs, language)

	// Add search condition to both queries
	if params.Search != "" {
		searchCondition := ` AND (COALESCE(st.name, s.name) LIKE ? OR COALESCE(st.description, s.description) LIKE ?)`
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
		return nil, nil, fmt.Errorf("failed to count terms: %v", err)
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
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY s.%s", params.OrderBy))
		if params.OrderDesc {
			queryBuilder.WriteString(" DESC")
		} else {
			queryBuilder.WriteString(" ASC")
		}
	} else {
		queryBuilder.WriteString(" ORDER BY s.display_order ASC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, paginationObj.Size, paginationObj.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list terms: %v", err)
	}
	defer rows.Close()

	// Scan results
	var terms []*domain.Term
	for rows.Next() {
		var s models.TermModel
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Description, &s.ImageKey, &s.Status, &s.DisplayOrder,
			&s.CreateID, &s.CreateDT, &s.ModifyID, &s.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		terms = append(terms, domain.BuildTermDomainFromModel(&s))
	}

	return terms, paginationObj, nil
}

// FindByID retrieves a term by ID.
func (r *termRepository) FindByID(ctx context.Context, id string) (*domain.Term, error) {
	query := `
		SELECT id, name, description, image_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM terms
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var s models.TermModel
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

	term := domain.BuildTermDomainFromModel(&s)

	return term, nil
}

// FindByName retrieves a term by name.
func (r *termRepository) FindByName(ctx context.Context, name string) (*domain.Term, error) {
	query := `
		SELECT id, name, description, image_key, status, display_order,
		create_id, create_dt, modify_id, modify_dt
		FROM terms
		WHERE name = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, name)

	var s models.TermModel
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

	term := domain.BuildTermDomainFromModel(&s)

	return term, nil
}

// Create inserts a new term into the database.
func (r *termRepository) Create(ctx context.Context, tx *sql.Tx, term *domain.Term) (int64, error) {
	query := `
		INSERT INTO terms (id, name, description, image_key, status, display_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		term.ID(),
		term.Name(),
		term.Description(),
		term.ImageKey(),
		enum.StatusActive,
		term.DisplayOrder(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create term: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing term in the database.
func (r *termRepository) Update(ctx context.Context, term *domain.Term) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE terms SET ")
	updates := []string{}

	if term.Name() != "" {
		updates = append(updates, "name = ?")
		args = append(args, term.Name())
	}

	if term.Description() != nil {
		updates = append(updates, "description = ?")
		args = append(args, term.Description())
	}

	// ImageKey can be nil, so we check if it's explicitly set
	if term.ImageKey() != nil {
		updates = append(updates, "image_key = ?")
		args = append(args, term.ImageKey())
	}

	if term.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, term.Status())
	}

	if term.DisplayOrder() != 0 {
		updates = append(updates, "display_order = ?")
		args = append(args, term.DisplayOrder())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, term.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update term: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a term by setting deleted_dt.
func (r *termRepository) Delete(ctx context.Context, id string) error {
	query := `
			UPDATE terms
			SET deleted_dt = ?,
				modify_dt = ?
			WHERE id = ? AND deleted_dt IS NULL`
	_, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete term: %v", err)
	}

	return nil
}

// ForceDelete permanently deletes a term from the database.
func (r *termRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	query := `DELETE FROM terms WHERE id = ?`
	_, err := r.db.Exec(ctx, tx, query, id)
	if err != nil {
		return fmt.Errorf("failed to force delete term: %v", err)
	}

	return nil
}
