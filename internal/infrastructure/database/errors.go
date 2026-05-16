package database

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

// MySQL error numbers
const (
	mysqlDuplicateEntry = 1062
	mysqlForeignKey     = 1452
	mysqlDataTooLong    = 1406
	mysqlDeadlock       = 1213
	mysqlLockTimeout    = 1205
)

// DBErrorKind classifies database errors for the application layer.
type DBErrorKind int

const (
	DBErrorUnknown     DBErrorKind = iota
	DBErrorNotFound                // row not found
	DBErrorDuplicate               // unique constraint violation
	DBErrorForeignKey              // foreign key violation
	DBErrorDataTooLong             // data exceeds column length
	DBErrorDeadlock                // deadlock or lock timeout
	DBErrorConnection              // connection-level failure
)

// ClassifyError maps a database error to a DBErrorKind.
func ClassifyError(err error) DBErrorKind {
	if err == nil {
		return DBErrorUnknown
	}

	if errors.Is(err, sql.ErrNoRows) {
		return DBErrorNotFound
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case mysqlDuplicateEntry:
			return DBErrorDuplicate
		case mysqlForeignKey:
			return DBErrorForeignKey
		case mysqlDataTooLong:
			return DBErrorDataTooLong
		case mysqlDeadlock, mysqlLockTimeout:
			return DBErrorDeadlock
		}
	}

	return DBErrorUnknown
}

// IsNotFound is a convenience check for sql.ErrNoRows.
func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// IsDuplicate checks if the error is a duplicate entry violation.
func IsDuplicate(err error) bool {
	return ClassifyError(err) == DBErrorDuplicate
}
