package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/db"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	mathtime "math-ai.com/math-ai/internal/shared/utils/time"
)

type Scanner interface {
	Scan(dest ...interface{}) error
}

// BaseRepository provides common database operations for repositories
type BaseRepository struct {
	db     db.IDatabase
	helper db.QueryHelper
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(database db.IDatabase) BaseRepository {
	return BaseRepository{
		db:     database,
		helper: db.QueryHelper{},
	}
}

// PaginatedList executes a paginated list query with automatic count query execution
// This eliminates duplicate code for building count queries
//
// Parameters:
//   - ctx: Context for the query
//   - baseQuery: The main SELECT query (without LIMIT/OFFSET)
//   - baseArgs: Arguments for the main query
//   - countQuery: The COUNT query (must return COUNT(*))
//   - countArgs: Arguments for the count query
//   - params: Pagination parameters
//
// Returns:
//   - Modified query with pagination applied
//   - Arguments for the paginated query
//   - Pagination object with total count
//   - Error if any
//
// Example:
//
//	baseQuery := `SELECT u.id, u.name FROM ma_users u WHERE u.deleted_dt IS NULL`
//	countQuery := `SELECT COUNT(*) FROM ma_users u WHERE u.deleted_dt IS NULL`
//	query, args, paginationObj, err := r.PaginatedList(ctx, baseQuery, nil, countQuery, nil, params)
func (r *BaseRepository) PaginatedList(
	ctx context.Context,
	baseQuery string,
	baseArgs []interface{},
	countQuery string,
	countArgs []interface{},
	params pagination.Params,
) (string, []interface{}, *pagination.Pagination, error) {

	// Execute count query
	var total int64
	row := r.db.QueryRow(ctx, nil, countQuery, countArgs...)
	if err := row.Scan(&total); err != nil {
		return "", nil, nil, fmt.Errorf("failed to execute count query: %v", err)
	}

	// Build pagination object
	paginationObj := pagination.NewPagination(params.Page, params.Limit, total)

	// Handle TakeAll flag
	if params.TakeAll {
		paginationObj.Size = total
		paginationObj.Skip = 0
		paginationObj.Page = 1
		paginationObj.TotalPages = 1
	}

	// Build final query with sorting and pagination
	var queryBuilder strings.Builder
	queryBuilder.WriteString(baseQuery)

	// Add sorting if specified
	if params.OrderBy != "" {
		orderClause := r.helper.BuildOrderBy(params.OrderBy, params.OrderDesc)
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(orderClause)
	}

	// Add pagination if not taking all
	if !params.TakeAll {
		paginationClause := r.helper.BuildPagination(paginationObj.Size, paginationObj.Skip)
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(paginationClause)
	}

	return queryBuilder.String(), baseArgs, paginationObj, nil
}

// FlexibleUpdate builds and executes an UPDATE query from a map of fields
// This eliminates the need for multiple if-blocks to check each field
//
// Parameters:
//   - ctx: Context for the query
//   - tx: Optional transaction (pass nil for non-transactional updates)
//   - table: Table name to update
//   - id: ID of the record to update
//   - updates: Map of field names to values
//   - allowedFields: Map of field names that are allowed to be updated (whitelist for security)
//
// Returns:
//   - Number of rows affected
//   - Error if any
//
// Example:
//
//	updates := map[string]interface{}{
//	    "name": "John Doe",
//	    "email": "john@example.com",
//	}
//	allowedFields := map[string]bool{"name": true, "email": true, "phone": true}
//	rowsAffected, err := r.FlexibleUpdate(ctx, nil, "ma_users", userID, updates, allowedFields)
func (r *BaseRepository) FlexibleUpdate(
	ctx context.Context,
	tx *sql.Tx,
	table string,
	id string,
	updates map[string]interface{},
	allowedFields map[string]bool,
) (int64, error) {

	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	// Filter updates to only allowed fields
	filteredUpdates := make(map[string]interface{})
	for field, value := range updates {
		// Security check: only allow whitelisted fields
		if allowedFields != nil && !allowedFields[field] {
			return 0, fmt.Errorf("field '%s' is not allowed to be updated", field)
		}
		filteredUpdates[field] = value
	}

	// Always update modify_dt
	filteredUpdates["modify_dt"] = mathtime.Now()

	// Build SET clause
	setClause, setArgs := r.helper.BuildUpdateSet(filteredUpdates)

	// Build full UPDATE query
	query := fmt.Sprintf(
		"UPDATE %s %s WHERE id = ? AND deleted_dt IS NULL",
		table,
		setClause,
	)

	// Append ID to args
	args := append(setArgs, id)

	// Execute update
	result, err := r.db.Exec(ctx, tx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update: %v", err)
	}

	return result.RowsAffected()
}

// BatchInsert inserts multiple records in a single query
// This is much more efficient than inserting records one by one
//
// Parameters:
//   - ctx: Context for the query
//   - tx: Optional transaction (recommended for batch operations)
//   - table: Table name to insert into
//   - columns: Column names to insert
//   - values: Slice of value slices (each inner slice represents one row)
//
// Returns:
//   - Number of rows affected
//   - Error if any
//
// Example:
//
//	columns := []string{"id", "name", "email"}
//	values := [][]interface{}{
//	    {"id1", "John", "john@example.com"},
//	    {"id2", "Jane", "jane@example.com"},
//	}
//	rowsAffected, err := r.BatchInsert(ctx, tx, "ma_users", columns, values)
func (r *BaseRepository) BatchInsert(
	ctx context.Context,
	tx *sql.Tx,
	table string,
	columns []string,
	values [][]interface{},
) (int64, error) {

	if len(values) == 0 {
		return 0, fmt.Errorf("no values to insert")
	}

	if len(columns) == 0 {
		return 0, fmt.Errorf("no columns specified")
	}

	// Build column list
	columnList := strings.Join(columns, ", ")

	// Build value placeholders
	var valuePlaceholders []string
	var args []interface{}

	for _, row := range values {
		if len(row) != len(columns) {
			return 0, fmt.Errorf("row has %d values but %d columns specified", len(row), len(columns))
		}

		// Create placeholders for this row (?, ?, ?)
		placeholders := make([]string, len(row))
		for i := range row {
			placeholders[i] = "?"
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(placeholders, ", ")+")")

		// Add row values to args
		args = append(args, row...)
	}

	// Build final INSERT query
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		table,
		columnList,
		strings.Join(valuePlaceholders, ", "),
	)

	// Execute insert
	result, err := r.db.Exec(ctx, tx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute batch insert: %v", err)
	}

	return result.RowsAffected()
}

// SoftDelete performs a soft delete by setting deleted_dt timestamp
//
// Parameters:
//   - ctx: Context for the query
//   - tx: Optional transaction
//   - table: Table name
//   - id: ID of the record to soft delete
//
// Returns:
//   - Number of rows affected
//   - Error if any
//
// Example:
//
//	rowsAffected, err := r.SoftDelete(ctx, tx, "ma_users", userID)
func (r *BaseRepository) SoftDelete(
	ctx context.Context,
	tx *sql.Tx,
	table string,
	id string,
) (int64, error) {

	now := mathtime.Now()
	query := fmt.Sprintf(
		"UPDATE %s SET deleted_dt = ?, modify_dt = ? WHERE id = ? AND deleted_dt IS NULL",
		table,
	)

	result, err := r.db.Exec(ctx, tx, query, []interface{}{now, now, id}...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute soft delete: %v", err)
	}

	return result.RowsAffected()
}

// HardDelete performs a hard delete by actually removing the record from database
//
// Parameters:
//   - ctx: Context for the query
//   - tx: Transaction (recommended for hard deletes)
//   - table: Table name
//   - id: ID of the record to delete
//
// Returns:
//   - Number of rows affected
//   - Error if any
//
// Example:
//
//	rowsAffected, err := r.HardDelete(ctx, tx, "ma_users", userID)
func (r *BaseRepository) HardDelete(
	ctx context.Context,
	tx *sql.Tx,
	table string,
	id string,
) (int64, error) {

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)

	result, err := r.db.Exec(ctx, tx, query, []interface{}{id}...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute hard delete: %v", err)
	}

	return result.RowsAffected()
}

// ExecuteInTransaction wraps a function in a database transaction
// Automatically commits on success and rolls back on error
//
// Parameters:
//   - fn: Function to execute within the transaction
//
// Returns:
//   - Error if any (from transaction or function)
//
// Example:
//
//	err := r.ExecuteInTransaction(func(tx *sql.Tx) error {
//	    _, err := r.db.Exec(ctx, tx, "INSERT INTO ...", args...)
//	    if err != nil {
//	        return err
//	    }
//	    _, err = r.db.Exec(ctx, tx, "UPDATE ...", args...)
//	    return err
//	})
func (r *BaseRepository) ExecuteInTransaction(fn func(tx *sql.Tx) error) error {
	return r.db.WithTransaction(fn)
}

// Helper returns the query helper for building SQL fragments
func (r *BaseRepository) Helper() db.QueryHelper {
	return r.helper
}
