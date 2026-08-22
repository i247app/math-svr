package repositories

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// mysqlDuplicateEntry is MySQL's ER_DUP_ENTRY — a UNIQUE constraint rejected
// the row.
const mysqlDuplicateEntry = 1062

// isDuplicateEntry reports whether err is a unique-constraint violation.
//
// Knowing which driver error means "someone else got there first" is
// infrastructure's job. Repositories use this to translate the driver error
// into a domain sentinel, so command handlers can handle a lost race without
// depending on the SQL driver.
func isDuplicateEntry(err error) bool {
	var myErr *mysql.MySQLError
	return errors.As(err, &myErr) && myErr.Number == mysqlDuplicateEntry
}
