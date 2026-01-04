package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_practices"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/db"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type userQuizPracticesRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewUserQuizPracticesRepository(database db.IDatabase) di.IUserQuizPracticesRepository {
	return &userQuizPracticesRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanUserQuizPractices is a reusable helper method to scan user quiz practices data from a row
func (r *userQuizPracticesRepository) scanUserQuizPractices(scanner Scanner) (*domain.UserQuizPractices, error) {
	var q models.UserQuizPracticesModel
	err := scanner.Scan(
		&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.Status,
		&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildUserQuizPracticesDomainFromModel(&q), nil
}

// FindByID retrieves a user latest quiz by ID.
func (r *userQuizPracticesRepository) FindByID(ctx context.Context, id string) (*domain.UserQuizPractices, error) {
	row := r.db.QueryRow(ctx, nil, queries.UserQuizPracticesFindByID, id)
	return r.scanUserQuizPractices(row)
}

// FindByUID retrieves the latest quiz for a user by UID.
func (r *userQuizPracticesRepository) FindByUID(ctx context.Context, uid string) (*domain.UserQuizPractices, error) {
	row := r.db.QueryRow(ctx, nil, queries.UserQuizPracticesFindByUID, uid)
	return r.scanUserQuizPractices(row)
}

// Create inserts a new user latest quiz into the database.
func (r *userQuizPracticesRepository) Create(ctx context.Context, tx *sql.Tx, quiz *domain.UserQuizPractices) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizPracticesInsert,
		quiz.ID(),
		quiz.UID(),
		quiz.Questions(),
		quiz.Answers(),
		quiz.AIReview(),
		quiz.Status(),
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing user latest quiz in the database.
func (r *userQuizPracticesRepository) Update(ctx context.Context, tx *sql.Tx, quiz *domain.UserQuizPractices) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizPracticesUpdate,
		PrepareForUpdate(quiz.Questions()),
		PrepareForUpdate(quiz.Answers()),
		PrepareForUpdate(quiz.AIReview()),
		mathtime.Now(),
		quiz.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// Delete performs a soft delete on a user latest quiz.
func (r *userQuizPracticesRepository) Delete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizPracticesSoftDelete, mathtime.Now(), mathtime.Now(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// DeleteByUID performs a soft delete of user quiz practices by UID.
func (r *userQuizPracticesRepository) DeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.UserQuizPracticesSoftDeleteByUID, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user quiz practices by UID: %v", err)
	}

	return nil
}

// ForceDelete permanently removes a user latest quiz from the database.
func (r *userQuizPracticesRepository) ForceDelete(ctx context.Context, tx *sql.Tx, id string) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserQuizPracticesForceDelete, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// ForceDeleteByUID permanently removes user quiz practices by UID from the database.
func (r *userQuizPracticesRepository) ForceDeleteByUID(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.UserQuizPracticesForceDeleteByUID, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user quiz practices by UID: %v", err)
	}

	return nil
}
