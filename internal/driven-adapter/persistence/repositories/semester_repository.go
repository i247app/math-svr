package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/semester"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/metadata"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

const (
	semesterTableName = "ma_semesters"
)

type semesterRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewSemesterRepository(database db.IDatabase) di.ISemesterRepository {
	return &semesterRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanSemester is a reusable helper method to scan semester data from a row
func (r *semesterRepository) scanSemester(scanner Scanner) (*domain.Semester, error) {
	var s models.SemesterModel
	err := scanner.Scan(
		&s.ID, &s.Name, &s.Description, &s.ImageKey, &s.Status, &s.DisplayOrder,
		&s.CreateID, &s.CreateDT, &s.ModifyID, &s.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildSemesterDomainFromModel(&s), nil
}

// DoTransaction wraps a function in a database transaction
func (r *semesterRepository) DoTransaction(ctx context.Context, handler db.HanderlerWithTx) error {
	return r.ExecuteInTransaction(handler)
}

// List retrieves a paginated list of semesters with optional search and sorting.
// Now uses BaseRepository.PaginatedList to eliminate duplication
func (r *semesterRepository) List(ctx context.Context, params di.ListSemestersParams) ([]*domain.Semester, *pagination.Pagination, error) {
	// Get language from context for translations
	language := metadata.GetLanguage(ctx)

	// Get base queries with language parameter
	baseQuery := queries.SemesterListQuery
	countQuery := queries.SemesterListCountQuery
	args := []interface{}{language}

	// Add search condition if provided
	if params.Search != "" {
		baseQuery, countQuery, args = queries.SemesterQueries{}.BuildListQueryWithSearch(language, params.Search)
	}

	// Build pagination params
	paginationParams := pagination.Params{
		Page:      params.Page,
		Limit:     params.Limit,
		OrderBy:   params.OrderBy,
		OrderDesc: params.OrderDesc,
		TakeAll:   params.TakeAll,
	}

	// Default ordering if not specified
	if paginationParams.OrderBy == "" {
		paginationParams.OrderBy = "s.display_order"
		paginationParams.OrderDesc = false
	} else {
		// Prefix with table alias
		paginationParams.OrderBy = "s." + paginationParams.OrderBy
	}

	// Use BaseRepository.PaginatedList for automatic count and pagination
	query, queryArgs, paginationObj, err := r.PaginatedList(
		ctx,
		baseQuery,
		args,
		countQuery,
		args,
		paginationParams,
	)
	if err != nil {
		return nil, nil, err
	}

	// Execute final query
	rows, err := r.db.Query(ctx, nil, query, queryArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list semesters: %v", err)
	}
	defer rows.Close()

	// Scan results
	var semesters []*domain.Semester
	for rows.Next() {
		semester, err := r.scanSemester(rows)
		if err != nil {
			return nil, nil, err
		}
		semesters = append(semesters, semester)
	}

	return semesters, paginationObj, nil
}

// FindByID retrieves a semester by ID.
func (r *semesterRepository) FindByID(ctx context.Context, id string) (*domain.Semester, error) {
	language := metadata.GetLanguage(ctx)

	row := r.db.QueryRow(ctx, nil, queries.SemesterFindByIDWithTranslation, language, id)
	return r.scanSemester(row)
}

// FindByName retrieves a semester by name.
// Now uses query constants from queries package
func (r *semesterRepository) FindByName(ctx context.Context, name string) (*domain.Semester, error) {
	language := metadata.GetLanguage(ctx)

	row := r.db.QueryRow(ctx, nil, queries.SemesterFindByNameWithTranslation, language, name)
	return r.scanSemester(row)
}

// Create inserts a new semester into the database.
// Now uses query constants from queries package
func (r *semesterRepository) Create(ctx context.Context, tx *sql.Tx, semester *domain.Semester) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.SemesterInsert,
		semester.ID(),
		semester.Name(),
		semester.Description(),
		semester.ImageKey(),
		enum.StatusActive,
		semester.DisplayOrder(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create semester: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing semester in the database.
// LEGACY METHOD - kept for backward compatibility
// New code should use UpdateFields instead
func (r *semesterRepository) Update(ctx context.Context, tx *sql.Tx, semester *domain.Semester) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.SemesterUpdate,
		PrepareForUpdate(semester.Name()),
		PrepareForUpdate(semester.Description()),
		PrepareForUpdate(semester.ImageKey()),
		PrepareForUpdate(semester.SemesterStatus()),
		PrepareForUpdate(semester.Note()),
		PrepareForUpdate(semester.DisplayOrder()),
		mathtime.Now(),
		semester.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create grade: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a semester by setting deleted_dt.
// Now uses BaseRepository.SoftDelete
func (r *semesterRepository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := r.SoftDelete(ctx, tx, semesterTableName, id)
	if err != nil {
		return fmt.Errorf("failed to delete semester: %v", err)
	}
	return nil
}

// ForceDelete permanently deletes a semester from the database.
// Now uses BaseRepository.HardDelete
func (r *semesterRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := r.HardDelete(ctx, tx, semesterTableName, id)
	if err != nil {
		return fmt.Errorf("failed to force delete semester: %v", err)
	}
	return nil
}

// CreateSemesterTranslation inserts a new semester translation into the database.
func (r *semesterRepository) CreateSemesterTranslation(ctx context.Context, tx *sql.Tx, semesterTranslation *domain.SemesterTranslation) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.SemesterTranslationInsert,
		semesterTranslation.ID(),
		semesterTranslation.SemesterID(),
		semesterTranslation.Language(),
		semesterTranslation.Name(),
		semesterTranslation.Description(),
		semesterTranslation.Note(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create semester translation: %v", err)
	}

	return result.RowsAffected()
}

// UpdateSemesterTranslation modifies an existing semester translation in the database.
func (r *semesterRepository) UpdateSemesterTranslation(ctx context.Context, tx *sql.Tx, semesterTranslation *domain.SemesterTranslation) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.SemesterTranslationUpdate,
		PrepareForUpdate(semesterTranslation.Name()),
		PrepareForUpdate(semesterTranslation.Description()),
		PrepareForUpdate(semesterTranslation.Note()),
		PrepareForUpdate(semesterTranslation.STStatus()),
		PrepareForUpdate(semesterTranslation.Status()),
		mathtime.Now(),
		semesterTranslation.SemesterID(),
		semesterTranslation.Language(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update semester translation: %v", err)
	}

	return result.RowsAffected()
}

// DeleteSemesterTranslationsBySemesterID soft deletes semester translations by semester ID.
func (r *semesterRepository) DeleteSemesterTranslationsBySemesterID(ctx context.Context, tx *sql.Tx, semesterID string) error {
	_, err := r.db.Exec(ctx, tx, queries.SemesterTranslationDeleteBySemesterID,
		mathtime.Now(),
		mathtime.Now(),
		semesterID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete semester translations: %v", err)
	}
	return nil
}

// ForceDeleteSemesterTranslationsBySemesterID permanently deletes semester translations by semester ID.
func (r *semesterRepository) ForceDeleteSemesterTranslationsBySemesterID(ctx context.Context, tx *sql.Tx, semesterID string) error {
	_, err := r.db.Exec(ctx, tx, queries.SemesterTranslationForceDeleteBySemesterID, semesterID)
	if err != nil {
		return fmt.Errorf("failed to force delete semester translations: %v", err)
	}
	return nil
}
