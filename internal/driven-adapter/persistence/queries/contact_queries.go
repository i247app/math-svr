package queries

// ContactQueries contains all SQL queries for contact repository
type ContactQueries struct{}

// Column lists
const (
	ContactSelectColumns = `id, uid, contact_name, contact_email, contact_phone, contact_message, is_read`
)

// Base queries
const (
	ContactBaseSelect = `SELECT ` + ContactSelectColumns + `
		FROM ma_contact_us`

	ContactCountBase = `SELECT COUNT(*)
		FROM ma_contact_us`
)

// Find queries
const (
	ContactFindByID = `SELECT ` + ContactSelectColumns + `
		FROM ma_contact_us
		WHERE id = ?`
)

// List queries
const (
	ContactListQuery = ContactBaseSelect

	ContactListCountQuery = ContactCountBase
)

// Mutation queries
const (
	ContactInsert = `INSERT INTO ma_contact_us (id, uid, contact_name, contact_email, contact_phone, contact_message, is_read, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	ContactMarkAsRead = `UPDATE ma_contact_us
		SET is_read = ?
		WHERE id = ?`
)

// BuildListQueryWithSearch adds search condition to list query
// Supports searching by contact name, email, phone, and message
func (ContactQueries) BuildListQueryWithSearch(searchTerm string) (string, string, []interface{}) {
	searchCondition := ` WHERE (contact_name LIKE ? OR contact_email LIKE ? OR contact_phone LIKE ? OR contact_message LIKE ?)`
	searchPattern := "%" + searchTerm + "%"

	query := ContactListQuery + searchCondition
	countQuery := ContactListCountQuery + searchCondition

	// Args: search pattern x4 for LIKE
	args := []interface{}{searchPattern, searchPattern, searchPattern, searchPattern}

	return query, countQuery, args
}
