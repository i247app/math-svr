package database

import (
	"context"
	"database/sql"
)

// Executor is the query surface shared by *sql.DB wrappers and *sql.Tx.
// Repositories depend on this so they can be reused inside a transaction.
type Executor interface {
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// RowScanner is satisfied by both *sql.Row and *sql.Rows, letting a single
// scan helper serve single-row and iterating queries.
type RowScanner interface {
	Scan(dest ...any) error
}

// SqlHandler is the full DB handle used by bootstrap and the UnitOfWork.
type SqlHandler interface {
	Executor
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PingContext(ctx context.Context) error
	Close() error
}
