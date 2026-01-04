package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_assessment"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

const (
	userQuizAssessmentTableName = "ma_user_quiz_assessments"
)

type userQuizAssessmentRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewUserQuizAssessmentRepository(database db.IDatabase) di.IUserQuizAssessmentRepository {
	return &userQuizAssessmentRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanUserQuizAssessment is a reusable helper method to scan user quiz assessment data from a row
func (r *userQuizAssessmentRepository) scanUserQuizAssessment(scanner Scanner) (*domain.UserQuizAssessment, error) {
	var q models.UserQuizAssessmentModel
	err := scanner.Scan(
		&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.AIDetectGrade, &q.Status,
		&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildUserQuizAssessmentDomainFromModel(&q), nil
}

// FindByID retrieves a quiz assessment by ID.
// Now uses query constants from queries package
func (r *userQuizAssessmentRepository) FindByID(ctx context.Context, id string) (*domain.UserQuizAssessment, error) {
	row := r.db.QueryRow(ctx, nil, queries.UserQuizAssessmentFindByID, id)
	return r.scanUserQuizAssessment(row)
}

// ListByUID retrieves paginated quiz assessments for a user.
// Now uses BaseRepository.PaginatedList to eliminate duplication
func (r *userQuizAssessmentRepository) ListByUID(ctx context.Context, params di.ListUserQuizAssessmentsParams) ([]*domain.UserQuizAssessment, *pagination.Pagination, error) {
	// Get base queries with UID parameter
	baseQuery := queries.UserQuizAssessmentBaseSelectByUID
	countQuery := queries.UserQuizAssessmentCountByUID
	args := []interface{}{params.UID}

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
		paginationParams.OrderBy = "create_dt"
		paginationParams.OrderDesc = true // Most recent first
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
		return nil, nil, fmt.Errorf("failed to list assessments: %v", err)
	}
	defer rows.Close()

	// Scan results
	var assessments []*domain.UserQuizAssessment
	for rows.Next() {
		assessment, err := r.scanUserQuizAssessment(rows)
		if err != nil {
			return nil, nil, err
		}
		assessments = append(assessments, assessment)
	}

	return assessments, paginationObj, nil
}

// Create inserts a new quiz assessment into the database.
// Now uses query constants from queries package
func (r *userQuizAssessmentRepository) Create(ctx context.Context, tx *sql.Tx, assessment *domain.UserQuizAssessment) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizAssessmentInsert,
		assessment.ID(),
		assessment.UID(),
		assessment.Questions(),
		assessment.Answers(),
		assessment.AIReview(),
		assessment.AIDetectGrade(),
		assessment.Status(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create quiz assessment: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing quiz assessment in the database.
// Now uses query constants from queries package
func (r *userQuizAssessmentRepository) Update(ctx context.Context, tx *sql.Tx, assessment *domain.UserQuizAssessment) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizAssessmentUpdate,
		PrepareForUpdate(assessment.Questions()),
		PrepareForUpdate(assessment.Answers()),
		PrepareForUpdate(assessment.AIReview()),
		PrepareForUpdate(assessment.AIDetectGrade()),
		mathtime.Now(),
		assessment.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update quiz assessment: %v", err)
	}

	return result.RowsAffected()
}

// Delete performs a soft delete on a quiz assessment.
// Now uses query constants from queries package
func (r *userQuizAssessmentRepository) Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizAssessmentSoftDelete, mathtime.Now(), mathtime.Now(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete quiz assessment: %v", err)
	}

	return result.RowsAffected()
}

// ForceDelete permanently removes a quiz assessment from the database.
// Now uses query constants from queries package
func (r *userQuizAssessmentRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizAssessmentForceDelete, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete quiz assessment: %v", err)
	}

	return result.RowsAffected()
}
