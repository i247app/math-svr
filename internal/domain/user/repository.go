package user

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	FindById(ctx context.Context, id int64) (*User, error)
	FindByUserId(ctx context.Context, userId string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	ListUsers(ctx context.Context, params *ListUsersParams) ([]*User, *pagination.Pagination, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) error
	DeleteById(ctx context.Context, id int64) error
	DeleteByUserId(ctx context.Context, userId string) error
	MarkStatusByUserId(ctx context.Context, userId string, status enum.UserStatusType) error
	SoftDeleteByUserId(ctx context.Context, userId string) error
}

type IAliasRepository interface {
	Create(ctx context.Context, alias *Alias) (*Alias, error)
	FindByAliasId(ctx context.Context, aliasId string) (*Alias, error)
	FindByAka(ctx context.Context, alias string) (*Alias, error)
	FindByUserId(ctx context.Context, userId string) ([]*Alias, error)
	UpdateByAliasId(ctx context.Context, alias *Alias) error
	DeleteByUserId(ctx context.Context, userId string) error
	MarkStatusByUserId(ctx context.Context, userId string, status enum.UserAliasStatusType) error
	SoftDeleteByUserId(ctx context.Context, userId string) error
}

type ListUsersParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}
