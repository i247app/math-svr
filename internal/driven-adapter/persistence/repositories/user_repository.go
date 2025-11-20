package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/core/di/repositories"
	"math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type userRepository struct {
	db db.IDatabase
}

func NewUserRepository(db db.IDatabase) repositories.IUserRepository {
	return &userRepository{
		db: db,
	}
}

// CreateWithAliases creates a user and their aliases in a single transaction.
func (r *userRepository) CreateUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx, uid string) (*user.User, error) {
	err := r.db.WithTransaction(handler)

	if err != nil {
		return nil, err
	}

	result, err := r.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteUserWithAssociations deletes a user and their associated records in a single transaction.
func (r *userRepository) DeleteUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx) error {
	err := r.db.WithTransaction(handler)
	if err != nil {
		return err
	}

	return nil
}

// ForceDeleteUserWithAssociations permanently deletes a user and their associated records in a single transaction.
func (r *userRepository) ForceDeleteUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx) error {
	err := r.db.WithTransaction(handler)
	if err != nil {
		return err
	}

	return nil
}

// GetUserByLoginName retrieves a user by their login name (email or phone).
func (r *userRepository) GetUserByLoginName(ctx context.Context, loginName string) (*user.User, error) {
	query := `
		SELECT u.id, u.name, u.phone, u.email, u.avatar_url, 
		u.role, u.status, l.hash_pass, u.create_id, u.create_dt, u.modify_id, u.modify_dt
		FROM users u
		JOIN aliases a ON u.id = a.uid
		JOIN logins l ON u.id = l.uid
		WHERE a.aka = ? AND u.deleted_dt IS NULL AND a.deleted_dt IS NULL AND l.deleted_dt IS NULL
	`
	result := r.db.QueryRow(ctx, nil, query, loginName)

	var u models.UserModel
	err := result.Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarUrl,
		&u.Role, &u.Status, &u.HashPassword, &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	user := user.UserDomainFromModel(&u)

	return user, nil
}

// List retrieves a paginated list of users with optional search and sorting.
func (r *userRepository) List(ctx context.Context, params repositories.ListUsersParams) ([]*user.User, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	// Base query
	queryBuilder.WriteString(`
		SELECT id, name, phone, email, avatar_url, 
		role, status, create_id, create_dt, modify_id, modify_dt
	`)

	// Add search condition
	if params.Search != "" {
		queryBuilder.WriteString(` AND name LIKE ? OR email LIKE ?`)
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM users"
	if params.Search != "" {
		countQuery += ` WHERE name LIKE ? OR email LIKE ? AND deleted_dt IS NULL`
	} else {
		countQuery += ` WHERE deleted_dt IS NULL`
	}
	var total int64
	countRow := r.db.QueryRow(ctx, nil, countQuery, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count users: %v", err)
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
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, pagination.Size, pagination.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list users: %v", err)
	}
	defer rows.Close()

	// Scan results
	var users []*user.User
	for rows.Next() {
		var u models.UserModel
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarUrl,
			&u.Role, &u.Status, &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		users = append(users, user.UserDomainFromModel(&u))
	}

	return users, pagination, nil
}

// FindByID retrieves a user by ID.
func (r *userRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT id, name, phone, email, avatar_url,
		role, status, create_id, create_dt, modify_id, modify_dt
		FROM users
		WHERE email = ? AND deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, id)

	var u models.UserModel
	err := result.Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarUrl,
		&u.Role, &u.Status, &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	user := user.UserDomainFromModel(&u)

	return user, nil
}

// FindByEmail retrieves a user by email.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, name, phone, email, avatar_url,
		role, status, create_id, create_dt, modify_id, modify_dt
		FROM users
		WHERE email = ? AND deleted_dt IS NULL
	`
	result := r.db.QueryRow(ctx, nil, query, email)

	var u models.UserModel
	err := result.Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarUrl,
		&u.Role, &u.Status, &u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	user := user.UserDomainFromModel(&u)

	return user, nil
}

// Create inserts a new user into the database.
func (r *userRepository) Create(ctx context.Context, tx *sql.Tx, user *user.User) (int64, error) {
	query := `
		INSERT INTO users (name, phone, email, avatar_url, role, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		user.Name(),
		user.Phone(),
		user.Email(),
		user.AvatarURL(),
		user.Role(),
		enum.StatusActive,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %v", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %v", err)
	}

	return insertedID, nil
}

// Update updates an existing user.
func (r *userRepository) Update(ctx context.Context, user *user.User) (int64, error) {
	query := `
		UPDATE users
		SET name = COALESCE(?, name),
			phone = COALESCE(?, phone),
			email = COALESCE(?, email),
			avatar_url = COALESCE(?, avatar_url),
			role = COALESCE(?, role),
			status = COALESCE(?, status)
		WHERE id = ? AND deleted_dt IS NULL
	`
	result, err := r.db.Exec(ctx, nil, query,
		user.Name(),
		user.Phone(),
		user.Email(),
		user.Role(),
		user.AvatarURL(),
		user.Status(),
		user.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update user: %v", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve affected rows: %v", err)
	}

	return affectedRows, nil
}

// Delete removes a user by ID.
func (r *userRepository) Delete(ctx context.Context, uid string) error {
	query := `
		UPDATE users
		SET deleted_dt = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(ctx, nil, query, time.Now().UTC(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// ForceDelete removes a user by ID permanently.
func (r *userRepository) ForceDelete(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`
	_, err := r.db.Exec(ctx, tx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user: %v", err)
	}
	return nil
}
