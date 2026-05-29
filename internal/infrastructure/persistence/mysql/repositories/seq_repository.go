package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"math-ai.com/math-ai/internal/domain/seq"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
)

const (
	seqTable   = "ma_seqs"
	seqColumns = `seq_name, current_value, prefix, padding, increment_by, modify_dt`
)

type SeqRepository struct {
	db database.Executor
}

func NewSeqRepository(db database.Executor) seq.IRepository {
	return &SeqRepository{db: db}
}

// Next atomically advances current_value for the named sequence and
// returns the formatted ID. It is safe under high concurrency and across
// multiple application instances because:
//
//  1. The UPDATE acquires a row-level X lock on ma_seqs(seq_name).
//     Concurrent UPDATEs to the same name block until commit, so two
//     callers can never observe the same post-increment value.
//  2. The SELECT runs on the same connection inside the same tx and
//     therefore sees the freshly-written current_value (own-write
//     visibility, regardless of isolation level).
//
// The method MUST be invoked inside a transaction (i.e. from a
// transaction.UnitOfWork callback). Outside a tx, MySQL's driver may
// route the two statements to different pooled connections, breaking
// the own-write guarantee.
func (r *SeqRepository) Next(ctx context.Context, name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("seq repo next: name is required")
	}

	updateQuery := `UPDATE ` + seqTable + `
		SET current_value = current_value + increment_by
		WHERE seq_name = ?`
	result, err := r.db.Exec(ctx, updateQuery, trimmed)
	if err != nil {
		return "", fmt.Errorf("seq repo next update (%s): %w", trimmed, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("seq repo next rows affected (%s): %w", trimmed, err)
	}
	if affected == 0 {
		return "", fmt.Errorf("seq repo next (%s): %w", trimmed, seq.ErrNotFound)
	}

	selectQuery := `SELECT current_value, prefix, padding FROM ` + seqTable +
		` WHERE seq_name = ?`
	var (
		current uint64
		prefix  string
		padding uint8
	)
	if err := r.db.QueryRow(ctx, selectQuery, trimmed).
		Scan(&current, &prefix, &padding); err != nil {
		return "", fmt.Errorf("seq repo next select (%s): %w", trimmed, err)
	}

	return formatSeqID(prefix, padding, current), nil
}

// Find reads the row for inspection / admin use. It never advances the
// counter; callers that need an ID must use Next.
func (r *SeqRepository) Find(ctx context.Context, name string) (*seq.Sequence, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("seq repo find: name is required")
	}
	query := `SELECT ` + seqColumns + ` FROM ` + seqTable + ` WHERE seq_name = ?`

	var m models.SeqModel
	err := r.db.QueryRow(ctx, query, trimmed).Scan(
		&m.SeqName, &m.CurrentValue, &m.Prefix, &m.Padding,
		&m.IncrementBy, &m.ModifyDt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("seq repo find (%s): %w", trimmed, err)
	}
	return ModelToDomainSeq(&m), nil
}

// formatSeqID renders the prefix + zero-padded counter. When padding is
// 0 (admin opted out) we emit the raw decimal — never truncate, since a
// missing leading zero is a silent collision risk.
func formatSeqID(prefix string, padding uint8, value uint64) string {
	str := strconv.FormatUint(value, 10)
	if int(padding) > len(str) {
		str = strings.Repeat("0", int(padding)-len(str)) + str
	}
	return prefix + str
}

func ModelToDomainSeq(m *models.SeqModel) *seq.Sequence {
	s := seq.NewSequence()
	s.SetName(m.SeqName)
	s.SetCurrentValue(m.CurrentValue)
	s.SetPrefix(m.Prefix)
	s.SetPadding(m.Padding)
	s.SetIncrementBy(m.IncrementBy)
	s.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return s
}
