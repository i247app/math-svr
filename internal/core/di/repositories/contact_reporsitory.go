package di

import (
	"context"
	"database/sql"

	"math-ai.com/math-ai/internal/applications/dto"
	domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)
type ListContactsParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}

type IContactRepository interface {
	List(ctx context.Context, params ListContactsParams) ([]*domain.Contact, *pagination.Pagination, error)
	CreateContact(ctx context.Context, tx *sql.Tx, contact *domain.Contact) (int64, error)
	GetContacts(ctx context.Context, params *dto.ListContactsParams) ([]*domain.Contact, error)
	FindByID(ctx context.Context, id string) (*domain.Contact, error)
	UpdateContactIsRead(ctx context.Context, id string, isRead bool) (int64, error)
}
