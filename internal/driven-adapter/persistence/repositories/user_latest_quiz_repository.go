package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user_latest_quiz"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/db"
)

type userLatestQuizRepository struct {
	db db.IDatabase
}

func NewUserLatestQuizRepository(db db.IDatabase) repositories.IUserLatestQuizRepository {
	return &userLatestQuizRepository{
		db: db,
	}
}

// FindByID retrieves a user latest quiz by ID.
func (r *userLatestQuizRepository) FindByID(ctx context.Context, id string) (*domain.UserLatestQuiz, error) {
	query := `
		SELECT id, uid, questions, awswers, ai_review, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_latest_quizzes
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var q models.UserLatestQuizModel
	err := result.Scan(
		&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.Status,
		&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	quiz := domain.BuildUserLatestQuizDomainFromModel(&q)

	return quiz, nil
}

// FindByUID retrieves the latest quiz for a user by UID.
func (r *userLatestQuizRepository) FindByUID(ctx context.Context, uid string) (*domain.UserLatestQuiz, error) {
	query := `
		SELECT id, uid, questions, awswers, ai_review, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_latest_quizzes
		WHERE uid = ? AND deleted_dt IS NULL
		ORDER BY create_dt DESC
		LIMIT 1
	`

	result := r.db.QueryRow(ctx, nil, query, uid)

	var q models.UserLatestQuizModel
	err := result.Scan(
		&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.Status,
		&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	quiz := domain.BuildUserLatestQuizDomainFromModel(&q)

	return quiz, nil
}

// List retrieves a list of user latest quizzes with pagination.
func (r *userLatestQuizRepository) List(ctx context.Context, limit, offset int) ([]*domain.UserLatestQuiz, error) {
	query := `
		SELECT id, uid, questions, awswers, ai_review, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_latest_quizzes
		WHERE deleted_dt IS NULL
		ORDER BY create_dt DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, nil, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var quizzes []*domain.UserLatestQuiz
	for rows.Next() {
		var q models.UserLatestQuizModel
		err := rows.Scan(
			&q.ID, &q.UID, &q.Questions, &q.Answers, &q.AIReview, &q.Status,
			&q.CreateID, &q.CreateDT, &q.ModifyID, &q.ModifyDT,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}

		quiz := domain.BuildUserLatestQuizDomainFromModel(&q)
		quizzes = append(quizzes, quiz)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return quizzes, nil
}

// Create inserts a new user latest quiz into the database.
func (r *userLatestQuizRepository) Create(ctx context.Context, tx *sql.Tx, quiz *domain.UserLatestQuiz) (int64, error) {
	query := `
		INSERT INTO user_latest_quizzes (id, uid, questions, awswers, ai_review, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		quiz.ID(),
		quiz.UID(),
		quiz.Questions(),
		quiz.Answers(),
		quiz.AIReview(),
		quiz.Status(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// Update modifies an existing user latest quiz in the database.
func (r *userLatestQuizRepository) Update(ctx context.Context, quiz *domain.UserLatestQuiz) (int64, error) {
	query := `
		UPDATE user_latest_quizzes
		SET questions = ?, awswers = ?, ai_review = ?, status = ?
		WHERE id = ? AND deleted_dt IS NULL
	`

	result, err := r.db.Exec(ctx, nil, query,
		quiz.Questions(),
		quiz.Answers(),
		quiz.AIReview(),
		quiz.Status(),
		quiz.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// Delete performs a soft delete on a user latest quiz.
func (r *userLatestQuizRepository) Delete(ctx context.Context, id string) (int64, error) {
	query := `
		UPDATE user_latest_quizzes
		SET deleted_dt = ?
		WHERE id = ? AND deleted_dt IS NULL
	`

	result, err := r.db.Exec(ctx, nil, query, time.Now(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// ForceDelete permanently removes a user latest quiz from the database.
func (r *userLatestQuizRepository) ForceDelete(ctx context.Context, id string) (int64, error) {
	query := `
		DELETE FROM user_latest_quizzes
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, nil, query, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete user latest quiz: %v", err)
	}

	return result.RowsAffected()
}
