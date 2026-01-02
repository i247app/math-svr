package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/grade"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/metadata"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

const (
	gradeTableName = "ma_grades"
)

type gradeRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewGradeRepository(database db.IDatabase) di.IGradeRepository {
	return &gradeRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanGrade is a reusable helper method to scan grade data from a row
func (r *gradeRepository) scanGrade(scanner Scanner) (*domain.Grade, error) {
	var g models.GradeModel
	err := scanner.Scan(
		&g.ID, &g.Label, &g.Description, &g.ImageKey, &g.Status, &g.DisplayOrder,
		&g.CreateID, &g.CreateDT, &g.ModifyID, &g.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildGradeDomainFromModel(&g), nil
}

// DoTransaction executes a series of operations within a transaction
func (r *gradeRepository) DoTransaction(ctx context.Context, handler db.HanderlerWithTx) error {
	err := r.ExecuteInTransaction(handler)
	if err != nil {
		return err
	}

	return nil
}

// List retrieves a paginated list of grades with optional search and sorting.
// Now uses BaseRepository.PaginatedList to eliminate duplication
func (r *gradeRepository) List(ctx context.Context, params di.ListGradesParams) ([]*domain.Grade, *pagination.Pagination, error) {
	// Get language from context for translations
	language := metadata.GetLanguage(ctx)

	// Get base queries with language parameter
	baseQuery := queries.GradeListQuery
	countQuery := queries.GradeListCountQuery
	args := []interface{}{language}

	// Add search condition if provided
	if params.Search != "" {
		baseQuery, countQuery, args = queries.GradeQueries{}.BuildListQueryWithSearch(language, params.Search)
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
		paginationParams.OrderBy = "g.display_order"
		paginationParams.OrderDesc = false
	} else {
		// Prefix with table alias
		paginationParams.OrderBy = "g." + paginationParams.OrderBy
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
		return nil, nil, fmt.Errorf("failed to list grades: %v", err)
	}
	defer rows.Close()

	// Scan results
	var grades []*domain.Grade
	for rows.Next() {
		grade, err := r.scanGrade(rows)
		if err != nil {
			return nil, nil, err
		}
		grades = append(grades, grade)
	}

	return grades, paginationObj, nil
}

// FindByID retrieves a grade by ID.
// Now uses query constants from queries package
func (r *gradeRepository) FindByID(ctx context.Context, id string) (*domain.Grade, error) {
	language := metadata.GetLanguage(ctx)

	row := r.db.QueryRow(ctx, nil, queries.GradeFindByIDWithTranslation, language, id)
	return r.scanGrade(row)
}

// FindByLabel retrieves a grade by label.
// Now uses query constants from queries package
func (r *gradeRepository) FindByLabel(ctx context.Context, label string) (*domain.Grade, error) {
	language := metadata.GetLanguage(ctx)

	row := r.db.QueryRow(ctx, nil, queries.GradeFindByLabelWithTranslation, language, label, label)
	return r.scanGrade(row)
}

// Create inserts a new grade into the database.
// Now uses query constants from queries package
func (r *gradeRepository) Create(ctx context.Context, tx *sql.Tx, grade *domain.Grade) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.GradeInsert,
		grade.ID(),
		grade.Label(),
		grade.Description(),
		grade.ImageKey(),
		enum.StatusActive,
		grade.DisplayOrder(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create grade: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing grade in the database.
// LEGACY METHOD - kept for backward compatibility
// New code should use UpdateFields instead
func (r *gradeRepository) Update(ctx context.Context, tx *sql.Tx, grade *domain.Grade) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.GradeUpdate,
		PrepareForUpdate(grade.Label()),
		PrepareForUpdate(grade.Description()),
		PrepareForUpdate(grade.ImageKey()),
		PrepareForUpdate(grade.GradeStatus()),
		PrepareForUpdate(grade.Note()),
		PrepareForUpdate(grade.DisplayOrder()),
		mathtime.Now(),
		grade.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create grade: %v", err)
	}

	return result.RowsAffected()
}

// Delete soft deletes a grade by setting deleted_dt.
func (r *gradeRepository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := r.SoftDelete(ctx, tx, gradeTableName, id)
	if err != nil {
		return fmt.Errorf("failed to delete grade: %v", err)
	}
	return nil
}

// ForceDelete permanently deletes a grade from the database.
// Now uses BaseRepository.HardDelete
func (r *gradeRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := r.HardDelete(ctx, tx, gradeTableName, id)
	if err != nil {
		return fmt.Errorf("failed to force delete grade: %v", err)
	}
	return nil
}

// CreateGradeTranslation inserts a new grade translation into the database.
func (r *gradeRepository) CreateGradeTranslation(ctx context.Context, tx *sql.Tx, gradeTranslation *domain.GradeTranslation) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.GradeTranslationInsert,
		gradeTranslation.ID(),
		gradeTranslation.GradeID(),
		gradeTranslation.Language(),
		gradeTranslation.Label(),
		gradeTranslation.Description(),
		gradeTranslation.Note(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create grade translation: %v", err)
	}

	return result.RowsAffected()
}

// UpdateGradeTranslation modifies an existing grade translation in the database.
func (r *gradeRepository) UpdateGradeTranslation(ctx context.Context, tx *sql.Tx, gradeTranslation *domain.GradeTranslation) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.GradeTranslationUpdate,
		PrepareForUpdate(gradeTranslation.Label()),
		PrepareForUpdate(gradeTranslation.Description()),
		PrepareForUpdate(gradeTranslation.Note()),
		mathtime.Now(),
		gradeTranslation.GradeID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update grade translation: %v", err)
	}

	return result.RowsAffected()
}

// DeleteGradeTranslationsByGradeID soft deletes grade translations by grade ID.
func (r *gradeRepository) DeleteGradeTranslationsByGradeID(ctx context.Context, tx *sql.Tx, gradeID string) error {
	_, err := r.db.Exec(ctx, tx, queries.GradeTranslationDeleteByGradeID, mathtime.Now(), mathtime.Now(), gradeID)
	if err != nil {
		return fmt.Errorf("failed to delete grade translations: %v", err)
	}
	return nil
}

// ForceDeleteGradeTranslationsByGradeID permanently deletes grade translations by grade ID.
func (r *gradeRepository) ForceDeleteGradeTranslationsByGradeID(ctx context.Context, tx *sql.Tx, gradeID string) error {
	_, err := r.db.Exec(ctx, tx, queries.GradeTranslationForceDeleteByGradeID, gradeID)
	if err != nil {
		return fmt.Errorf("failed to force delete grade translations: %v", err)
	}
	return nil
}
