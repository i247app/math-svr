package repositories

import "strings"

// Active-row filter convention (project-wide, 2026-09-21)
//
// Every table exposes one constant of the exact shape
//
//	<x>ActiveWhere = `x.status IN (?) AND x.deleted_dt IS NULL`
//	<x>ActiveArgs() = []any{enum.StatusActive}
//
// and every read places it LAST in the WHERE clause, after the business
// predicates:
//
//	WHERE (u.uid = ?) AND u.status IN (?) AND u.deleted_dt IS NULL
//	args:  uid, enum.StatusActive
//
// The <entity>_status column is a business lifecycle (PENDING, ARCHIVED,
// REVOKED, …) and is filtered per query when it matters; it is never part
// of the active-row filter. Soft delete stamps deleted_dt (and flips status
// to INACTIVE), so `deleted_dt IS NULL` alone hides deleted rows.

// whereActive renders a WHERE clause from an optional business fragment and
// the table's active-row filter. `filter` is what the build*Filter helpers
// return: zero or more predicates each prefixed with " AND ". The active
// filter always comes last.
func whereActive(filter, active string) string {
	filter = strings.TrimPrefix(strings.TrimSpace(filter), "AND ")
	if filter == "" {
		return ` WHERE ` + active
	}
	return ` WHERE ` + filter + ` AND ` + active
}
