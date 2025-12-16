package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/repositories"
	domain "math-ai.com/math-ai/internal/core/domain/contact"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/db"
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
		INSERT INTO contact_us (id, uid, contact_name, contact_email, contact_phone, contact_message) 
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := cr.db.Exec(ctx, tx, query,
		contact.ID(),
		contact.UID(),
		contact.ContactName(),
		contact.ContactEmail(),
		contact.ContactPhone(),
		contact.ContactMessage(),
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
	query := `SELECT id, uid, contact_name, contact_email, contact_phone, contact_message
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
			&c.ID, &c.UID, &c.ContactName, &c.ContactEmail, &c.ContactPhone, &c.ContactMessage,
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

func (cr *contactRepository) CountContacts(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM contact_us`
	row := cr.db.QueryRow(ctx, nil, query)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count contacts: %v", err)
	}

	return count, nil
}
