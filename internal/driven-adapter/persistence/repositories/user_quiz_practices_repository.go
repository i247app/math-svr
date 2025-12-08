package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user_quiz_practices"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/db"
)

type userQuizPracticesRepository struct {
	db db.IDatabase
}

func NewUserQuizPracticesRepository(db db.IDatabase) di.IUserQuizPracticesRepository {
	return &userQuizPracticesRepository{
		db: db,
	}
}

// FindByID retrieves a user latest quiz by ID.
func (r *userQuizPracticesRepository) FindByID(ctx context.Context, id string) (*domain.UserQuizPractices, error) {
	query := `
		SELECT id, uid, questions, answers, ai_review, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_quiz_practices
		WHERE id = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var q models.UserQuizPracticesModel
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

	quiz := domain.BuildUserQuizPracticesDomainFromModel(&q)

	return quiz, nil
}

// FindByUID retrieves the latest quiz for a user by UID.
func (r *userQuizPracticesRepository) FindByUID(ctx context.Context, uid string) (*domain.UserQuizPractices, error) {
	query := `
		SELECT id, uid, questions, answers, ai_review, status,
		create_id, create_dt, modify_id, modify_dt
		FROM user_quiz_practices
		WHERE uid = ? AND deleted_dt IS NULL
		ORDER BY create_dt DESC
		LIMIT 1
	`

	result := r.db.QueryRow(ctx, nil, query, uid)

	var q models.UserQuizPracticesModel
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

	quiz := domain.BuildUserQuizPracticesDomainFromModel(&q)

	return quiz, nil
}

// Create inserts a new user latest quiz into the database.
func (r *userQuizPracticesRepository) Create(ctx context.Context, tx *sql.Tx, quiz *domain.UserQuizPractices) (int64, error) {
	query := `
		INSERT INTO user_quiz_practices (id, uid, questions, answers, ai_review, status)
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
func (r *userQuizPracticesRepository) Update(ctx context.Context, quiz *domain.UserQuizPractices) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE user_quiz_practices SET ")
	updates := []string{}
	if quiz.Questions() != "" {
		updates = append(updates, "questions = ?")
		args = append(args, quiz.Questions())
	}

	if quiz.Answers() != "" {
		updates = append(updates, "answers = ?")
		args = append(args, quiz.Answers())
	}

	if quiz.AIReview() != "" {
		updates = append(updates, "ai_review = ?")
		args = append(args, quiz.AIReview())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update for user latest quiz")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? OR uid = ? AND deleted_dt IS NULL")
	args = append(args, quiz.ID(), quiz.UID())

	query := queryBuilder.String()

	result, err := r.db.Exec(ctx, nil, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// Delete performs a soft delete on a user latest quiz.
func (r *userQuizPracticesRepository) Delete(ctx context.Context, id string) (int64, error) {
	query := `
		UPDATE user_quiz_practices
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ? AND deleted_dt IS NULL
	`

	result, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), time.Now().UTC(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user latest quiz: %v", err)
	}

	return result.RowsAffected()
}

// ForceDelete permanently removes a user latest quiz from the database.
func (r *userQuizPracticesRepository) ForceDelete(ctx context.Context, id string) (int64, error) {
	query := `
		DELETE FROM user_quiz_practices
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, nil, query, id)
	if err != nil {
		return 0, fmt.Errorf("failed to force delete user latest quiz: %v", err)
	}

	return result.RowsAffected()
}
