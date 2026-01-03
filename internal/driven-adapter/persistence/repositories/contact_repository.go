package repositories

import (
	"context"
	"database/sql"
	"fmt"

	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/queries"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type contactRepository struct {
	BaseRepository // Embed BaseRepository for common operations
}

func NewContactRepository(database db.IDatabase) di.IContactRepository {
	return &contactRepository{
		BaseRepository: NewBaseRepository(database),
	}
}

// scanContact is a reusable helper method to scan contact data from a row
func (r *contactRepository) scanContact(scanner Scanner) (*domain.Contact, error) {
	var c models.ContactModel
	err := scanner.Scan(
		&c.ID, &c.UID, &c.ContactName, &c.ContactEmail,
		&c.ContactPhone, &c.ContactMessage, &c.IsRead,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan error: %v", err)
	}

	return domain.BuildContactDomainFromModel(&c), nil
}

// List retrieves a paginated list of contacts with optional search and sorting.
func (r *contactRepository) List(ctx context.Context, params di.ListContactsParams) ([]*domain.Contact, *pagination.Pagination, error) {
	// Get base queries
	baseQuery := queries.ContactListQuery
	countQuery := queries.ContactListCountQuery
	args := []interface{}{}

	// Add search condition if provided
	if params.Search != "" {
		baseQuery, countQuery, args = queries.ContactQueries{}.BuildListQueryWithSearch(params.Search)
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
		paginationParams.OrderBy = "create_dt"
		paginationParams.OrderDesc = true // Most recent first
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
		return nil, nil, fmt.Errorf("failed to list contacts: %v", err)
	}
	defer rows.Close()

	// Scan results
	var contacts []*domain.Contact
	for rows.Next() {
		contact, err := r.scanContact(rows)
		if err != nil {
			return nil, nil, err
		}
		contacts = append(contacts, contact)
	}

	return contacts, paginationObj, nil
}

// Create inserts a new contact into the database.
func (r *contactRepository) Create(ctx context.Context, tx *sql.Tx, contact *domain.Contact) (int64, error) {
	result, err := r.db.Exec(ctx, tx, queries.ContactInsert,
		contact.ID(),
		contact.UID(),
		contact.ContactName(),
		contact.ContactEmail(),
		contact.ContactPhone(),
		contact.ContactMessage(),
		false,
		mathtime.Now(),
		mathtime.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create contact: %v", err)
	}

	return result.LastInsertId()
}

// FindByID retrieves a contact by ID.
func (r *contactRepository) FindByID(ctx context.Context, id string) (*domain.Contact, error) {
	row := r.db.QueryRow(ctx, nil, queries.ContactFindByID, id)
	return r.scanContact(row)
}

// MarkAsRead updates the is_read status of a contact.
func (r *contactRepository) MarkAsRead(ctx context.Context, id string, isRead bool) (int64, error) {
	result, err := r.db.Exec(ctx, nil, queries.ContactMarkAsRead, isRead, id)
	if err != nil {
		return 0, fmt.Errorf("failed to update contact is_read: %v", err)
	}

	return result.RowsAffected()
}
