package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_assessment"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type userQuizAssessmentRepository struct {
	db db.IDatabase
}

func NewUserQuizAssessmentRepository(db db.IDatabase) di.IUserQuizAssessmentRepository {
	return &userQuizAssessmentRepository{
		db: db,
	}
}

// FindByID retrieves a quiz assessment by ID.
func (r *userQuizAssessmentRepository) FindByID(ctx context.Context, id string) (*domain.UserQuizAssessment, error) {
	query := `
		SELECT id, uid, questions, answers, ai_review, ai_detect_grade, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_quiz_assessments
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var q models.UserQuizAssessmentModel
	err := result.Scan(
		&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.AIDetectGrade, &q.Status,
		&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	assessment := domain.BuildUserQuizAssessmentDomainFromModel(&q)

	return assessment, nil
}

// ListByUID retrieves paginated quiz assessments for a user.
func (r *userQuizAssessmentRepository) ListByUID(ctx context.Context, params di.ListUserQuizAssessmentsParams) ([]*domain.UserQuizAssessment, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	// Base query
	queryBuilder.WriteString(`
		SELECT id, uid, questions, answers, ai_review, ai_detect_grade, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_quiz_assessments
		WHERE uid = ? AND deleted_dt IS NULL`)
	args = append(args, params.UID)

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM user_quiz_assessments WHERE uid = ? AND deleted_dt IS NULL"
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countQuery, params.UID)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count assessments: %v", err)
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
		queryBuilder.WriteString(" ORDER BY create_dt DESC")
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, paginationObj.Size, paginationObj.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list assessments: %v", err)
	}
	defer rows.Close()

	// Scan results
	var assessments []*domain.UserQuizAssessment
	for rows.Next() {
		var q models.UserQuizAssessmentModel
		if err := rows.Scan(
			&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.AIDetectGrade, &q.Status,
			&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		assessments = append(assessments, domain.BuildUserQuizAssessmentDomainFromModel(&q))
	}

	return assessments, paginationObj, nil
}

// Create inserts a new quiz assessment into the database.
func (r *userQuizAssessmentRepository) Create(ctx context.Context, tx *sql.Tx, assessment *domain.UserQuizAssessment) (int64, error) {
	query := `
		INSERT INTO user_quiz_assessments (id, uid, questions, answers, ai_review, ai_detect_grade, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		assessment.ID(),
		assessment.UID(),
		assessment.Questions(),
		assessment.Answers(),
		assessment.AIReview(),
		assessment.AIDetectGrade(),
		assessment.Status(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create quiz assessment: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing quiz assessment in the database.
func (r *userQuizAssessmentRepository) Update(ctx context.Context, assessment *domain.UserQuizAssessment) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE user_quiz_assessments SET ")
	updates := []string{}

	if assessment.Questions() != "" {
		updates = append(updates, "questions = ?")
		args = append(args, assessment.Questions())
	}

	if assessment.Answers() != "" {
		updates = append(updates, "answers = ?")
		args = append(args, assessment.Answers())
	}

	if assessment.AIReview() != "" {
		updates = append(updates, "ai_review = ?")
		args = append(args, assessment.AIReview())
	}

	if assessment.AIDetectGrade() != "" {
		updates = append(updates, "ai_detect_grade = ?")
		args = append(args, assessment.AIDetectGrade())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update for quiz assessment")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, assessment.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update quiz assessment: %v", err)
	}

	return result.RowsAffected()
}

// Delete performs a soft delete on a quiz assessment.
func (r *userQuizAssessmentRepository) Delete(ctx context.Context, id string) (int64, error) {
	query := `
		UPDATE user_quiz_assessments
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL
	`

	result, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), time.Now().UTC(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete quiz assessment: %v", err)
	}

	return result.RowsAffected()
}

// ForceDelete permanently removes a quiz assessment from the database.
func (r *userQuizAssessmentRepository) ForceDelete(ctx context.Context, id string) (int64, error) {
	query := `
		DELETE FROM user_quiz_assessments
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, nil, query, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete quiz assessment: %v", err)
	}

	return result.RowsAffected()
}
