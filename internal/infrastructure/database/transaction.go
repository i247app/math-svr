package database

import (
	"context"
	"database/sql"
	"fmt"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// WithTransaction executes fn inside a database transaction. It commits on
// success and rolls back on error or panic. The caller's ctx is forwarded
// to fn so that anything constructed inside (notably TxWithLogs) can bind
// to the same request scope and produce ctx-tagged log lines.
func WithTransaction(ctx context.Context, db SqlHandler, fn func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("database: begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			logger.From(ctx).Error("[TX] Rolled back (panic)", "panic", fmt.Sprintf("%v", p))
			panic(p) // re-panic after rollback
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.From(ctx).Error("[TX] Rollback failed", "err", rbErr)
		} else {
			logger.From(ctx).Warn("[TX] Rolled back")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("database: commit tx: %w", err)
	}

	logger.From(ctx).Info("[TX] Committed")
	return nil
}
