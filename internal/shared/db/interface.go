package db

import (
	"context"
	"database/sql"
)

// IDatabase defines the database interface.
type IDatabase interface {
	GetDB() *sql.DB
	WithTransaction(function func(tx *sql.Tx) error) error // Updated to pass tx
	Query(ctx context.Context, tx *sql.Tx, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row
	Exec(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}

type HanderlerWithTx func(tx *sql.Tx) error
