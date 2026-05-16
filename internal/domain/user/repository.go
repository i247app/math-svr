package user

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/shared/pagination"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindById(ctx context.Context, id int64) (*User, error)
	FindByUserId(ctx context.Context, userId uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	ListUsers(ctx context.Context, params *ListUsersParams) ([]*User, *pagination.Pagination, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) error
	DeleteById(ctx context.Context, id int64) error
	DeleteByUserId(ctx context.Context, userId uuid.UUID) error
	MarkStatusByUserId(ctx context.Context, userId uuid.UUID, status enum.UserStatusType) error
}

type IAliasRepository interface {
	Create(ctx context.Context, alias *Alias) (*Alias, error)
	FindByAliasId(ctx context.Context, aliasId uuid.UUID) (*Alias, error)
	FindByAka(ctx context.Context, alias string) (*Alias, error)
	FindByUserId(ctx context.Context, userId uuid.UUID) ([]*Alias, error)
	UpdateByAliasId(ctx context.Context, alias *Alias) error
	DeleteByUserId(ctx context.Context, userId uuid.UUID) error
	MarkStatusByUserId(ctx context.Context, userId uuid.UUID, status enum.UserAliasStatusType) error
}

type ListUsersParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}
