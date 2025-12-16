package di

import (
	"context"
	"database/sql"

	"math-ai.com/math-ai/internal/applications/dto"
	domain "math-ai.com/math-ai/internal/core/domain/contact"
)

type IContactRepository interface {
	CreateContact(ctx context.Context, tx *sql.Tx, contact *domain.Contact) (int64, error)
	GetContacts(ctx context.Context, params *dto.ListContactsParams) ([]*domain.Contact, error)
	CountContacts(ctx context.Context) (int64, error)
}
