package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type userRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewUserRepository(database db.IDatabase) di.IUserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanUser is a reusable helper method to scan user data from a row
func (r *userRepository) scanUser(scanner Scanner) (*domain.User, error) {
	var u models.UserModel
	err := scanner.Scan(
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

	return domain.BuildUserDomainFromModel(&u), nil
}

// scanUserWithPassword is a reusable helper method to scan user data with password from a row
func (r *userRepository) scanUserWithPassword(scanner Scanner) (*domain.User, error) {
	var u models.UserModel
	err := scanner.Scan(
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

	return domain.BuildUserDomainFromModel(&u), nil
}

// DoTransaction wraps a function in a database transaction
func (r *userRepository) DoTransaction(ctx context.Context, handler db.HanderlerWithTx) error {
	return r.ExecuteInTransaction(handler)
}

// GetUserByLoginName retrieves a user by their login name (email or phone) with role information.
// Now uses query constants from queries package
func (r *userRepository) GetUserByLoginName(ctx context.Context, loginName string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, nil, queries.UserGetByLoginName, loginName)
	return r.scanUserWithPassword(row)
}

// List retrieves a paginated list of users with optional search and sorting.
// Now uses BaseRepository.PaginatedList to eliminate duplication
func (r *userRepository) List(ctx context.Context, params di.ListUsersParams) ([]*domain.User, *pagination.Pagination, error) {
	// Get base queries
	baseQuery := queries.UserListQuery
	countQuery := queries.UserListCountQuery
	args := []interface{}{}

	// Add search condition if provided
	if params.Search != "" {
		baseQuery, countQuery, args = queries.UserQueries{}.BuildListQueryWithSearch(params.Search)
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
		paginationParams.OrderBy = "u.create_dt"
		paginationParams.OrderDesc = true // Most recent first
	} else {
		// Prefix with table alias
		paginationParams.OrderBy = "u." + paginationParams.OrderBy
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
		return nil, nil, fmt.Errorf("failed to list users: %v", err)
	}
	defer rows.Close()

	// Scan results
	var users []*domain.User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, user)
	}

	return users, paginationObj, nil
}

// FindByID retrieves a user by ID with optional role information.
// Now uses query constants from queries package
func (r *userRepository) FindByID(ctx context.Context, uid string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, nil, queries.UserFindByID, uid)
	return r.scanUser(row)
}

// FindByEmail retrieves a user by email with optional role information.
// Now uses query constants from queries package
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, nil, queries.UserFindByEmail, email)
	return r.scanUser(row)
}

// Create inserts a new user into the database.
// Now uses query constants from queries package
func (r *userRepository) Create(ctx context.Context, tx *sql.Tx, user *domain.User) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserInsert,
		user.ID(),
		user.Name(),
		user.Phone(),
		user.Email(),
		user.AvatarKey(),
		user.DOB(),
		user.RoleID(),
		enum.StatusActive,
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %v", err)
	}

	return result.LastInsertId()
}

// Update updates an existing user.
// Now uses query constants from queries package
func (r *userRepository) Update(ctx context.Context, tx *sql.Tx, user *domain.User) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.UserUpdate,
		PrepareForUpdate(user.Name()),
		PrepareForUpdate(user.Phone()),
		PrepareForUpdate(user.Email()),
		PrepareForUpdate(user.DOB()),
		PrepareForUpdate(user.RoleID()),
		PrepareForUpdate(user.AvatarKey()),
		mathtime.Now(),
		user.ID(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update user: %v", err)
	}

	return result.RowsAffected()
}

// Delete removes a user by ID.
// Now uses query constants from queries package
func (r *userRepository) Delete(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.UserSoftDelete, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// ForceDelete removes a user by ID permanently.
// Now uses query constants from queries package
func (r *userRepository) ForceDelete(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.UserForceDelete, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user: %v", err)
	}
	return nil
}

// StoreUserAlias stores a user alias in the database.
// Now uses query constants from queries package
func (r *userRepository) StoreUserAlias(ctx context.Context, tx *sql.Tx, alias *domain.Alias) error {
	_, err := r.db.Exec(ctx, tx, queries.AliasInsert,
		alias.ID(),
		alias.UID(),
		alias.Aka(),
		enum.StatusActive,
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to store user alias: %v", err)
	}
	return nil
}

// DeleteUserAlias deletes user aliases by user ID.
// Now uses query constants from queries package
func (r *userRepository) DeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.AliasSoftDeleteByUID, mathtime.Now(), mathtime.Now(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user aliases: %v", err)
	}
	return nil
}

// ForceDeleteUserAlias permanently deletes user aliases by user ID.
// Now uses query constants from queries package
func (r *userRepository) ForceDeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error {
	_, err := r.db.Exec(ctx, tx, queries.AliasForceDeleteByUID, uid)
	if err != nil {
		return fmt.Errorf("failed to force delete user aliases: %v", err)
	}
	return nil
}
