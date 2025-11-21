package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/level"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type levelRepository struct {
	db db.IDatabase
}

func NewLevelRepository(db db.IDatabase) repositories.ILevelRepository {
	return &levelRepository{
		db: db,
	}
}

// List retrieves a paginated list of levels with optional search and sorting.
func (r *levelRepository) List(ctx context.Context, params repositories.ListLevelsParams) ([]*domain.Level, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	// Base query - note: using 'discription' to match the actual table column
	queryBuilder.WriteString(`
		SELECT id, label, discription, status,
		create_id, create_dt, modify_id, modify_dt
		FROM levels WHERE deleted_dt IS NULL
	`)

	// Add search condition
	if params.Search != "" {
		queryBuilder.WriteString(` AND (label LIKE ? OR discription LIKE ?)`)
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM levels WHERE deleted_dt IS NULL"
	if params.Search != "" {
		countQuery += ` AND (label LIKE ? OR discription LIKE ?)`
	}
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countQuery, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count levels: %v", err)
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
		queryBuilder.WriteString(" ORDER BY label ASC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, pagination.Size, pagination.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list levels: %v", err)
	}
	defer rows.Close()

	// Scan results
	var levels []*domain.Level
	for rows.Next() {
		var l models.LevelModel
		if err := rows.Scan(
			&l.ID, &l.Label, &l.Description, &l.Status,
			&l.CreateID, &l.CreateDT, &l.ModifyID, &l.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		levels = append(levels, domain.BuildLevelDomainFromModel(&l))
	}

	return levels, pagination, nil
}

// FindByID retrieves a level by ID.
func (r *levelRepository) FindByID(ctx context.Context, id string) (*domain.Level, error) {
	query := `
		SELECT id, label, discription, status,
		create_id, create_dt, modify_id, modify_dt
		FROM levels
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var l models.LevelModel
	err := result.Scan(
		&l.ID, &l.Label, &l.Description, &l.Status,
		&l.CreateID, &l.CreateDT, &l.ModifyID, &l.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	level := domain.BuildLevelDomainFromModel(&l)

	return level, nil
}

// FindByLabel retrieves a level by label.
func (r *levelRepository) FindByLabel(ctx context.Context, label string) (*domain.Level, error) {
	query := `
		SELECT id, label, discription, status,
		create_id, create_dt, modify_id, modify_dt
		FROM levels
		WHERE label = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, label)

	var l models.LevelModel
	err := result.Scan(
		&l.ID, &l.Label, &l.Description, &l.Status,
		&l.CreateID, &l.CreateDT, &l.ModifyID, &l.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	level := domain.BuildLevelDomainFromModel(&l)

	return level, nil
}

// Create inserts a new level into the database.
func (r *levelRepository) Create(ctx context.Context, tx *sql.Tx, level *domain.Level) (int64, error) {
	query := `
		INSERT INTO levels (id, label, discription, status)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		level.ID(),
		level.Label(),
		level.Description(),
		enum.StatusActive,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create level: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing level in the database.
func (r *levelRepository) Update(ctx context.Context, level *domain.Level) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE levels SET ")
	updates := []string{}

	if level.Label() != "" {
		updates = append(updates, "label = ?")
		args = append(args, level.Label())
	}

	if level.Description() != "" {
		updates = append(updates, "discription = ?")
		args = append(args, level.Description())
	}

	if level.Status() != "" {
		updates = append(updates, "status = ?")
		args = append(args, level.Status())
	}

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, level.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update level: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a level by setting deleted_dt.
func (r *levelRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE levels SET deleted_dt = NOW() WHERE id = ? AND deleted_dt IS NULL`
	_, err := r.db.Exec(ctx, nil, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete level: %v", err)
	}

	return nil
}

// ForceDelete permanently deletes a level from the database.
func (r *levelRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	query := `DELETE FROM levels WHERE id = ?`
	_, err := r.db.Exec(ctx, tx, query, id)
	if err != nil {
		return fmt.Errorf("failed to force delete level: %v", err)
	}

	return nil
}
