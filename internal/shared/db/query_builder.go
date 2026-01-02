package db

import (
	"fmt"
	"strings"
)

// QueryBuilder helps build SQL queries in a more structured way
// without external dependencies. Uses strings.Builder internally.
type QueryBuilder struct {
	query strings.Builder
	args  []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		args: make([]interface{}, 0),
	}
}

// WriteString writes a raw SQL string to the query
func (qb *QueryBuilder) WriteString(sql string) *QueryBuilder {
	qb.query.WriteString(sql)
	return qb
}

// AddCondition adds a WHERE/AND condition with placeholder
// Example: AddCondition("name = ?", "John")
func (qb *QueryBuilder) AddCondition(condition string, args ...interface{}) *QueryBuilder {
	qb.query.WriteString(condition)
	qb.args = append(qb.args, args...)
	return qb
}

// AddArg adds an argument for a placeholder
func (qb *QueryBuilder) AddArg(arg interface{}) *QueryBuilder {
	qb.args = append(qb.args, arg)
	return qb
}

// AddArgs adds multiple arguments
func (qb *QueryBuilder) AddArgs(args ...interface{}) *QueryBuilder {
	qb.args = append(qb.args, args...)
	return qb
}

// String returns the built query string
func (qb *QueryBuilder) String() string {
	return qb.query.String()
}

// Args returns the query arguments
func (qb *QueryBuilder) Args() []interface{} {
	return qb.args
}

// ToSQL returns both query and args
func (qb *QueryBuilder) ToSQL() (string, []interface{}) {
	return qb.query.String(), qb.args
}

// QueryHelper provides static helper methods for common SQL patterns
type QueryHelper struct{}

// BuildWhereClause builds a WHERE clause from a map of conditions
// Example: BuildWhereClause(map[string]interface{}{"name": "John", "age": 25})
// Returns: "WHERE name = ? AND age = ?", []interface{}{"John", 25}
func (QueryHelper) BuildWhereClause(conditions map[string]interface{}) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var clauses []string
	var args []interface{}

	for field, value := range conditions {
		if value == nil {
			clauses = append(clauses, fmt.Sprintf("%s IS NULL", field))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s = ?", field))
			args = append(args, value)
		}
	}

	whereClause := "WHERE " + strings.Join(clauses, " AND ")
	return whereClause, args
}

// BuildLikeConditions builds LIKE conditions for search
// Example: BuildLikeConditions([]string{"name", "email"}, "john")
// Returns: "(name LIKE ? OR email LIKE ?)", []interface{}{"%john%", "%john%"}
func (QueryHelper) BuildLikeConditions(fields []string, searchTerm string) (string, []interface{}) {
	if len(fields) == 0 || searchTerm == "" {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	searchPattern := "%" + searchTerm + "%"

	for _, field := range fields {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		args = append(args, searchPattern)
	}

	condition := "(" + strings.Join(conditions, " OR ") + ")"
	return condition, args
}

// BuildUpdateSet builds SET clause for UPDATE
// Example: BuildUpdateSet(map[string]interface{}{"name": "John", "age": 25})
// Returns: "SET name = ?, age = ?", []interface{}{"John", 25}
func (QueryHelper) BuildUpdateSet(updates map[string]interface{}) (string, []interface{}) {
	if len(updates) == 0 {
		return "", nil
	}

	var setPairs []string
	var args []interface{}

	for field, value := range updates {
		if value == nil {
			setPairs = append(setPairs, fmt.Sprintf("%s = NULL", field))
		} else {
			setPairs = append(setPairs, fmt.Sprintf("%s = ?", field))
			args = append(args, value)
		}
	}

	setClause := "SET " + strings.Join(setPairs, ", ")
	return setClause, args
}

// BuildInCondition builds an IN condition
// Example: BuildInCondition("id", []interface{}{"1", "2", "3"})
// Returns: "id IN (?, ?, ?)", []interface{}{"1", "2", "3"}
func (QueryHelper) BuildInCondition(field string, values []interface{}) (string, []interface{}) {
	if len(values) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}

	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
	return condition, values
}

// BuildOrderBy builds ORDER BY clause
// Example: BuildOrderBy("name", true) returns "ORDER BY name DESC"
func (QueryHelper) BuildOrderBy(field string, desc bool) string {
	if field == "" {
		return ""
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", field, order)
}

// BuildPagination builds LIMIT OFFSET clause
// Example: BuildPagination(10, 20) returns "LIMIT 10 OFFSET 20"
func (QueryHelper) BuildPagination(limit, offset int64) string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
}
