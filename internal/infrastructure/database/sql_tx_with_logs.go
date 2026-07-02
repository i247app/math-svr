package database

import (
	"context"
	"database/sql"

	"math-ai.com/math-ai/internal/shared/colors"
)

// TxWithLogs wraps *sql.Tx so queries inside a transaction produce the same
// [SQL] log lines as DatabaseWithLogs. It satisfies Executor, which is what
// repositories depend on.
//
// Unlike DatabaseWithLogs, a TxWithLogs is short-lived (one transaction)
// and therefore safe to bind to the request ctx that started the
// transaction. Every query routed through it inherits the request's token
// tail and uid in the log output for free, with no repo-layer changes.
type TxWithLogs struct {
	tx  *sql.Tx
	ctx context.Context
}

// NewTxWithLogs binds tx to ctx so every query logged via this wrapper
// carries the request prefix. ctx is typically the same one passed to
// BeginTx. A nil ctx is silently replaced with context.Background().
func NewTxWithLogs(ctx context.Context, tx *sql.Tx) *TxWithLogs {
	if ctx == nil {
		ctx = context.Background()
	}
	return &TxWithLogs{tx: tx, ctx: ctx}
}

func (t *TxWithLogs) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	span := startDBSpan(ctx, query)
	logInputSQL(ctx, colors.FGCyan, query, args...)
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		logQueryError(ctx, err)
		endDBSpan(span, err)
		return nil, err
	}
	logQueryRowsResult(t.ctx, rows)
	endDBSpan(span, nil)
	return rows, nil
}

func (t *TxWithLogs) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	span := startDBSpan(ctx, query)
	defer span.End()
	logInputSQL(ctx, colors.FGCyan, query, args...)
	row := t.tx.QueryRow(query, args...)
	logQueryRowResult(ctx, row)
	return row
}

func (t *TxWithLogs) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	span := startDBSpan(ctx, query)
	logInputSQL(ctx, colors.FGYellow, query, args...)
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		logQueryError(ctx, err)
		endDBSpan(span, err)
		return nil, err
	}
	logExecResult(ctx, result)
	endDBSpan(span, nil)
	return result, nil
}
