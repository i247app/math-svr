package repositories

import (
	"context"
	"database/sql"

	"math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type ListUsersParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type IUserRepository interface {
	CreateUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx, uid string) (*user.User, error)
	DeleteUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx) error
	ForceDeleteUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx) error

	GetUserByLoginName(ctx context.Context, loginName string) (*user.User, error)

	// users
	List(ctx context.Context, req ListUsersParams) ([]*user.User, pagination.Pagination, error)
	FindByID(ctx context.Context, id string) ([]*user.User, error)
	FindByEmail(ctx context.Context, email string) ([]*user.User, error)
	Create(ctx context.Context, tx *sql.Tx, user *user.User) (int64, error) // Add tx parameter
	Update(ctx context.Context, user *user.User) (int64, error)
	Delete(ctx context.Context, id string) error
	ForceDelete(ctx context.Context, tx *sql.Tx, id string) error

	// // aliases
	// StoreUserAlias(ctx context.Context, tx *sql.Tx, dto *CreateAliasDTO) error // Add tx parameter
	// DeleteUserAlias(ctx context.Context, uid string) error
	// ForceDeleteUserAlias(ctx context.Context, tx *sql.Tx, uid string) error

	// // logins
	// StoreLogin(ctx context.Context, tx *sql.Tx, dto *CreateLoginDTO) error // Add tx parameter
	// DeleteLogin(ctx context.Context, uid string) error
	// ForceDeleteLogin(ctx context.Context, tx *sql.Tx, uid string) error
}
