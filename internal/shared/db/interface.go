package db

import (
	"context"
	"database/sql"
)

type HanderlerWithTx func(tx *sql.Tx) error

type IDatabase interface {
	GetDB() *sql.DB
	WithTransaction(function HanderlerWithTx) error // Updated to pass tx
	Query(ctx context.Context, tx *sql.Tx, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row
	Exec(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}
