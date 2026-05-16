package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/colors"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const (
	argsColor   = colors.FGYellow
	errColor    = colors.FGRed
	okColor     = colors.FGGreen
	bgUnchanged = colors.BGReset
)

func logInputSQL(ctx context.Context, queryColor colors.Color, query string, args ...any) {
	lg := logger.From(ctx)
	q := strings.TrimSpace(query)

	if !lg.ColorEnabled() {
		if len(args) > 0 {
			lg.Info("[SQL] "+q, "args", args)
			return
		}
		lg.Info("[SQL] " + q)
		return
	}

	coloredQuery := colors.Colorize(queryColor, bgUnchanged, q)
	if len(args) == 0 {
		lg.Info("[SQL] " + coloredQuery)
		return
	}
	coloredArgs := colors.Colorize(argsColor, bgUnchanged, fmt.Sprintf("%v", args))
	lg.Info("[SQL] " + coloredQuery + " args=" + coloredArgs)
}

func logQueryError(ctx context.Context, err error) {
	lg := logger.From(ctx)
	if !lg.ColorEnabled() {
		lg.Error("[SQL ERROR]", "err", err)
		return
	}
	lg.Error("[SQL ERROR] " + colors.Colorize(errColor, bgUnchanged, err.Error()))
}

func logExecResult(ctx context.Context, result sql.Result) {
	lg := logger.From(ctx)
	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	if !lg.ColorEnabled() {
		lg.Info("[SQL OK]",
			"rows_affected", rowsAffected,
			"last_insert_id", lastInsertID,
		)
		return
	}

	body := fmt.Sprintf("rows_affected=%d last_insert_id=%d", rowsAffected, lastInsertID)
	lg.Info("[SQL OK] " + colors.Colorize(okColor, bgUnchanged, body))
}

func logQueryRowsResult(ctx context.Context, rows *sql.Rows) {
	cols, err := rows.Columns()
	if err != nil {
		return
	}

	lg := logger.From(ctx)
	if !lg.ColorEnabled() {
		lg.Info("[SQL OK]", "columns", cols)
		return
	}

	lg.Info("[SQL OK] " + colors.Colorize(okColor, bgUnchanged, fmt.Sprintf("columns=%v", cols)))
}

func logQueryRowResult(ctx context.Context, _ *sql.Row) {
	lg := logger.From(ctx)
	if !lg.ColorEnabled() {
		lg.Info("[SQL OK] row queried")
		return
	}
	lg.Info("[SQL OK] " + colors.Colorize(okColor, bgUnchanged, "row queried"))
}
