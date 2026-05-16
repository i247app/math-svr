package database

import (
	"context"
	"database/sql"

	"math-ai.com/math-ai/internal/shared/colors"
)

// DatabaseWithLogs wraps *sql.DB and logs every query through the project
// logger (logger.From). Because *sql.DB is process-shared, this wrapper
// cannot hold a request ctx — non-tx queries log under context.Background()
// (anon prefix) by default. Callers that have a request ctx and want it
// reflected in the [SQL] line can opt in via WithContext(ctx).
type DatabaseWithLogs struct {
	db *sql.DB
	// ctx context.Context
}

func NewDatabaseWithLogs(db *sql.DB) *DatabaseWithLogs {
	return &DatabaseWithLogs{db: db}
}

// DB returns the underlying *sql.DB (useful for migration tools or direct access).
func (d *DatabaseWithLogs) DB() *sql.DB {
	return d.db
}

// WithContext returns a logging clone that emits SQL log lines under ctx.
// The underlying *sql.DB is shared — only the logging context changes. Use
// this from non-tx read paths (query handlers) when you want token tail /
// uid prefixes on the SQL trace.
func (d *DatabaseWithLogs) WithContext(ctx context.Context) *DatabaseWithLogs {
	if ctx == nil {
		ctx = context.Background()
	}
	cp := *d
	// cp.ctx = ctx
	return &cp
}

func (d *DatabaseWithLogs) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	logInputSQL(ctx, colors.FGCyan, query, args...)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		logQueryError(ctx, err)
		return nil, err
	}
	logQueryRowsResult(ctx, rows)
	return rows, nil
}

func (d *DatabaseWithLogs) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	logInputSQL(ctx, colors.FGCyan, query, args...)
	row := d.db.QueryRow(query, args...)
	logQueryRowResult(ctx, row)
	return row
}

func (d *DatabaseWithLogs) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	logInputSQL(ctx, colors.FGYellow, query, args...)
	result, err := d.db.Exec(query, args...)
	if err != nil {
		logQueryError(ctx, err)
		return nil, err
	}
	logExecResult(ctx, result)
	return result, nil
}

func (d *DatabaseWithLogs) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	logInputSQL(ctx, colors.FGMagenta, "BEGIN TRANSACTION")
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		logQueryError(ctx, err)
		return nil, err
	}
	return tx, nil
}

func (d *DatabaseWithLogs) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DatabaseWithLogs) Close() error {
	return d.db.Close()
}
