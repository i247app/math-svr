package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type contactRepository struct {
	db db.IDatabase
}

func NewContactRepository(db db.IDatabase) di.IContactRepository {
	return &contactRepository{
		db: db,
	}
}

func (cr *contactRepository) CreateContact(ctx context.Context, tx *sql.Tx, contact *domain.Contact) (int64, error) {
	query := `
		INSERT INTO contact_us (id, uid, contact_name, contact_email, contact_phone, contact_message, is_read)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := cr.db.Exec(ctx, tx, query,
		contact.ID(),
		contact.UID(),
		contact.ContactName(),
		contact.ContactEmail(),
		contact.ContactPhone(),
		contact.ContactMessage(),
		false,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create contact: %v", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %v", err)
	}

	return insertedID, nil
}

func (cr *contactRepository) GetContacts(ctx context.Context, params *dto.ListContactsParams) ([]*domain.Contact, error) {
	query := `SELECT id, uid, contact_name, contact_email, contact_phone, contact_message, is_read
	          FROM contact_us
	          LIMIT ? OFFSET ?`

	rows, err := cr.db.Query(ctx, nil, query, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %v", err)
	}
	defer rows.Close()

	var contacts []*domain.Contact
	for rows.Next() {
		var c models.ContactModel
		err := rows.Scan(
			&c.ID, &c.UID, &c.ContactName, &c.ContactEmail, &c.ContactPhone, &c.ContactMessage, &c.IsRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %v", err)
		}
		contacts = append(contacts, domain.BuildContactDomainFromModel(&c))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contacts: %v", err)
	}

	return contacts, nil
}

func (cr *contactRepository) List(ctx context.Context, params di.ListContactsParams) ([] *domain.Contact, *pagination.Pagination, error) {
	var queryBuilder strings.Builder
	
	args := []interface{}{}

	queryBuilder.WriteString(`
		SELECT id, uid, contact_name, contact_email, contact_phone, contact_message, is_read
		FROM contact_us
	`)

	// Count total records for pagination
	countQuery := "SELECT COUNT(*) FROM users"
	if params.Search != "" {
		countQuery += ` WHERE name LIKE ? OR email LIKE ? AND deleted_dt IS NULL`
	} else {
		countQuery += ` WHERE deleted_dt IS NULL`
	}
	var total int64
	countRow := cr.db.QueryRow(ctx, nil, countQuery, args...)	
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

	if !params.TakeAll {
		queryBuilder.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, pagination.Size, pagination.Skip)
	}

	
	// Execute query
	rows, err := cr.db.Query(ctx, nil, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list contacts: %v", err)
	}
	defer rows.Close()

	// Scan results
	var contacts []*domain.Contact
	for rows.Next() {
		var c models.ContactModel
		if err := rows.Scan(
			&c.ID, &c.UID, &c.ContactName, &c.ContactEmail,
			&c.ContactPhone, &c.ContactMessage, &c.IsRead,
		); err != nil {
			return nil, nil, fmt.Errorf("scan error: %v", err)
		}
		contacts = append(contacts, domain.BuildContactDomainFromModel(&c))
	}

	return contacts, pagination, nil
}

// FindByID retrieves a contact by ID.
func (cr *contactRepository) FindByID(ctx context.Context, id string) (*domain.Contact, error) {
	query := `
		SELECT id, uid, contact_name, contact_email, contact_phone, contact_message, is_read
		FROM contact_us
		WHERE id = ?
	`

	result := cr.db.QueryRow(ctx, nil, query, id)

	var c models.ContactModel
	err := result.Scan(
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

// UpdateContactIsRead updates the is_read status of a contact.
func (cr *contactRepository) UpdateContactIsRead(ctx context.Context, id string, isRead bool) (int64, error) {
	query := `
		UPDATE contact_us
		SET is_read = ?
		WHERE id = ?
	`

	result, err := cr.db.Exec(ctx, nil, query, isRead, id)
	if err != nil {
		return 0, fmt.Errorf("failed to update contact is_read: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}

	return rowsAffected, nil
}
