package repositories

import (
	"context"
	"database/sql"

	domain "math-ai.com/math-ai/internal/core/domain/user"
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
	CreateUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx, uid int64) (*domain.User, error)
	DeleteUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx) error
	ForceDeleteUserWithAssociations(ctx context.Context, handler db.HanderlerWithTx) error

	GetUserByLoginName(ctx context.Context, loginName string) (*domain.User, error)

	// users
	List(ctx context.Context, params ListUsersParams) ([]*domain.User, *pagination.Pagination, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, tx *sql.Tx, user *domain.User) (int64, error)
	Update(ctx context.Context, user *domain.User) (int64, error)
	Delete(ctx context.Context, uid int64) error
	ForceDelete(ctx context.Context, tx *sql.Tx, uid int64) error

	// aliases
	StoreUserAlias(ctx context.Context, tx *sql.Tx, alias *domain.Alias) error
	DeleteUserAlias(ctx context.Context, uid int64) error
	ForceDeleteUserAlias(ctx context.Context, tx *sql.Tx, uid int64) error

	// logins
	StoreLogin(ctx context.Context, tx *sql.Tx, login *domain.Login) error
	DeleteLogin(ctx context.Context, uid int64) error
	ForceDeleteLogin(ctx context.Context, tx *sql.Tx, uid int64) error
}
