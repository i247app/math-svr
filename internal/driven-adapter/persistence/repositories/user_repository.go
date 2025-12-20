package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type userRepository struct {
	db db.IDatabase
}

func NewUserRepository(db db.IDatabase) di.IUserRepository {
	return &userRepository{
		db: db,
	}
}

// ForceDeleteUserWithAssociations permanently deletes a user and their associated records in a single transaction.
func (r *userRepository) DoTransaction(ctx context.Context, handler db.HanderlerWithTx) error {
	err := r.db.WithTransaction(handler)
	if err != nil {
		return err
	}

	return nil
}

// GetUserByLoginName retrieves a user by their login name (email or phone) with role information.
func (r *userRepository) GetUserByLoginName(ctx context.Context, loginName string) (*domain.User, error) {
	query := `
		SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		       u.role_id, u.status, l.hash_pass,
		       u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		       r.name as role_name
		FROM users u
		JOIN aliases a ON u.id = a.uid
		JOIN logins l ON u.id = l.uid
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE a.aka = ? AND u.deleted_dt IS NULL AND a.deleted_dt IS NULL AND l.deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, loginName)

	var u models.UserModel
	err := result.Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
		&u.RoleID, &u.Status, &u.HashPassword,
		&u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
		&u.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	user := domain.BuildUserDomainFromModel(&u)

	return user, nil
}

// List retrieves a paginated list of users with optional search and sorting.
// Now supports joining with roles table for enhanced user information.
func (r *userRepository) List(ctx context.Context, params di.ListUsersParams) ([]*domain.User, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	// Base query with LEFT JOIN to roles table
	queryBuilder.WriteString(`
		SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		       u.role_id, u.status,
		       u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		       r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE u.deleted_dt IS NULL
	`)

	// Add search condition
	if params.Search != "" {
		queryBuilder.WriteString(` AND (u.name LIKE ? OR u.email LIKE ?)`)
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM users u WHERE u.deleted_dt IS NULL"
	countArgs := []interface{}{}
	if params.Search != "" {
		countQuery += ` AND (u.name LIKE ? OR u.email LIKE ?)`
		searchTerm := "%" + params.Search + "%"
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	var total int64
	countRow := r.db.QueryRow(ctx, nil, countQuery, countArgs...)
	if err := countRow.Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count users: %v", err)
	}

	// Initialize pagination
	paginationResult := pagination.NewPagination(params.Page, params.Limit, total)
	if params.TakeAll {
		paginationResult.Size = total
		paginationResult.Skip = 0
		paginationResult.Page = 1
		paginationResult.TotalPages = 1
	}

	// Add sorting
	if params.OrderBy != "" {
		queryBuilder.WriteString(fmt.Sprintf(" ORDER BY u.%s", params.OrderBy))
		if params.OrderDesc {
			queryBuilder.WriteString(" DESC")
		}
	}

	// Add pagination
	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, paginationResult.Size, paginationResult.Skip)
	}

	// Execute query
	rows, err := r.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list users: %v", err)
	}
	defer rows.Close()

	// Scan results
	var users []*domain.User
	for rows.Next() {
		var u models.UserModel
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
			&u.RoleID, &u.Status,
			&u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
			&u.Role,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}

		users = append(users, domain.BuildUserDomainFromModel(&u))
	}

	return users, paginationResult, nil
}

// FindByID retrieves a user by ID with optional role information.
func (r *userRepository) FindByID(ctx context.Context, uid string) (*domain.User, error) {
	query := `
		SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		       u.role_id, u.status,
		       u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		       r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE u.id = ? AND u.deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, uid)

	var u models.UserModel
	err := result.Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
		&u.RoleID, &u.Status,
		&u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
		&u.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	user := domain.BuildUserDomainFromModel(&u)

	return user, nil
}

// FindByEmail retrieves a user by email with optional role information.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT u.id, u.name, u.phone, u.email, u.avatar_key, u.dob,
		       u.role_id, u.status,
		       u.create_id, u.create_dt, u.modify_id, u.modify_dt,
		       r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id AND r.deleted_dt IS NULL
		WHERE u.email = ? AND u.deleted_dt IS NULL
	`

	result := r.db.QueryRow(ctx, nil, query, email)

	var u models.UserModel
	err := result.Scan(
		&u.ID, &u.Name, &u.Phone, &u.Email, &u.AvatarKey, &u.Dob,
		&u.RoleID, &u.Status,
		&u.CreateID, &u.CreateDT, &u.ModifyID, &u.ModifyDT,
		&u.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	user := domain.BuildUserDomainFromModel(&u)

	return user, nil
}

// Create inserts a new user into the database.
func (r *userRepository) Create(ctx context.Context, tx *sql.Tx, user *domain.User) (int64, error) {
	query := `
		INSERT INTO users (id, name, phone, email, avatar_key, dob, role_id, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, tx, query,
		user.ID(),
		user.Name(),
		user.Phone(),
		user.Email(),
		user.AvatarKey(),
		user.DOB(),
		user.RoleID(),
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
func (r *userRepository) Update(ctx context.Context, user *domain.User) (int64, error) {
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("UPDATE users SET ")
	updates := []string{}

	if user.Name() != "" {
		updates = append(updates, "name = ?")
		args = append(args, user.Name())
	}

	if user.Phone() != "" {
		updates = append(updates, "phone = ?")
		args = append(args, user.Phone())
	}

	if user.Email() != "" {
		updates = append(updates, "email = ?")
		args = append(args, user.Email())
	}

	if user.DOB() != nil {
		updates = append(updates, "dob = ?")
		args = append(args, user.DOB())
	}

	if user.RoleID() != "" {
		updates = append(updates, "role_id = ?")
		args = append(args, user.RoleID())
	}

	if user.AvatarKey() != nil {
		updates = append(updates, "avatar_key = ?")
		args = append(args, user.AvatarKey())
	}

	updates = append(updates, "modify_dt = ?")
	args = append(args, time.Now().UTC())

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(updates, ", "))
	queryBuilder.WriteString(" WHERE id = ? AND deleted_dt IS NULL")
	args = append(args, user.ID())

	result, err := r.db.Exec(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update user: %v", err)
	}

	return result.RowsAffected()
}

// Delete removes a user by ID.
func (r *userRepository) Delete(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		UPDATE users
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(ctx, tx, query, time.Now().UTC(), time.Now().UTC(), uid)
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

// StoreUserAlias stores a user alias in the database.
func (r *userRepository) StoreUserAlias(ctx context.Context, tx *sql.Tx, alias *domain.Alias) error {
	query := `
		INSERT INTO aliases (id, uid, aka, status)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(ctx, tx, query,
		alias.ID(),
		alias.UID(),
		alias.Aka(),
		enum.StatusActive,
	)
	if err != nil {
		return fmt.Errorf("failed to store user alias: %v", err)
	}
	return nil
}

// DeleteUserAlias deletes user aliases by user ID.
func (r *userRepository) DeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		UPDATE aliases
		SET deleted_dt = ?,
			modify_dt = ?
		WHERE uid = ? AND deleted_dt IS NULL
	`
	_, err := r.db.Exec(ctx, tx, query, time.Now().UTC(), time.Now().UTC(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user aliases: %v", err)
	}
	return nil
}

// ForceDeleteUserAlias permanently deletes user aliases by user ID.
func (r *userRepository) ForceDeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error {
	query := `
		DELETE FROM aliases
		WHERE uid = ?
	`
	_, err := r.db.Exec(ctx, tx, query, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user aliases: %v", err)
	}
	return nil
}
