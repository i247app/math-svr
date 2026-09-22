package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

const (
	aliasTable = "ma_aliases"

	aliasColumns = `id, alias_id, uid, aka, alias_status, rpt_flg, kwords, note, create_id, create_dt, modify_id, modify_dt`

	// Login resolution (alias.FindByAka -> user.FindByUserId) relies on this
	// filter so a soft-deleted account cannot log back in through its alias.
	aliasActiveWhere = `status IN (?) AND deleted_dt IS NULL`
)

func aliasActiveArgs() []any {
	return []any{enum.StatusActive}
}

type AliasRepository struct {
	db database.Executor
}

func NewAliasRepository(db database.Executor) user.IAliasRepository {
	return &AliasRepository{db: db}
}

func scanAlias(s database.RowScanner) (*models.AliasModel, error) {
	var m models.AliasModel
	if err := s.Scan(&m.Id, &m.AliasId, &m.UserId, &m.Aka, &m.AliasStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy runs a single-row lookup. `where` is a package-controlled SQL
// fragment (never user input). aliasActiveWhere is appended last;
// a missing row is reported as (nil, nil), never sql.ErrNoRows.
func (r *AliasRepository) findOneBy(ctx context.Context, where string, args ...any) (*user.Alias, error) {
	fullArgs := slices.Concat(args, aliasActiveArgs())
	query := `SELECT ` + aliasColumns + ` FROM ` + aliasTable +
		` WHERE (` + where + `) AND ` + aliasActiveWhere

	m, err := scanAlias(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("alias repo find (%s): %w", where, err)
	}
	return ModelToDomainAlias(m), nil
}

func (r *AliasRepository) Create(ctx context.Context, alias *user.Alias) (*user.Alias, error) {
	query := `
		INSERT INTO ` + aliasTable + ` (alias_id, uid, aka, alias_status, rpt_flg, kwords, note, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query, alias.AliasId(), alias.UserId(),
		alias.Aka(), alias.AliasStatus(), alias.RptFlg(), alias.Kwords(), alias.Note(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("alias repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("alias repo last insert id: %w", err)
	}
	alias.SetId(id)

	return alias, nil
}

func (r *AliasRepository) FindByAliasId(ctx context.Context, aliasId int64) (*user.Alias, error) {
	return r.findOneBy(ctx, "alias_id = ?", aliasId)
}

func (r *AliasRepository) FindByAka(ctx context.Context, aka string) (*user.Alias, error) {
	return r.findOneBy(ctx, "aka = ?", aka)
}

func (r *AliasRepository) FindByUserId(ctx context.Context, userId int64) ([]*user.Alias, error) {
	args := slices.Concat([]any{userId}, aliasActiveArgs())
	query := `SELECT ` + aliasColumns + ` FROM ` + aliasTable +
		` WHERE (uid = ?) AND ` + aliasActiveWhere
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("alias repo find by uid: %w", err)
	}
	defer rows.Close()

	var aliases []*user.Alias
	for rows.Next() {
		m, err := scanAlias(rows)
		if err != nil {
			return nil, fmt.Errorf("alias repo scan row: %w", err)
		}
		aliases = append(aliases, ModelToDomainAlias(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("alias repo rows iteration: %w", err)
	}
	return aliases, nil
}

func (r *AliasRepository) UpdateByAliasId(ctx context.Context, alias *user.Alias) error {
	query := `
		UPDATE ` + aliasTable + `
		SET aka = COALESCE(?, aka),
			alias_status = COALESCE(?, alias_status),
			rpt_flg= COALESCE(?, rpt_flg),
			kwords= COALESCE(?, kwords),
			note = COALESCE(?, note),
			modify_id = COALESCE(?, modify_id),
			modify_dt = COALESCE(?, modify_dt)
		WHERE alias_id = ?
	`

	if _, err := r.db.Exec(ctx, query, alias.Aka(), alias.AliasStatus(), alias.RptFlg(), alias.Kwords(), alias.Note(), alias.ModifyId(), alias.ModifyDt(), alias.AliasId()); err != nil {
		return fmt.Errorf("alias repo update by alias id: %w", err)
	}
	return nil
}

func (r *AliasRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM `+aliasTable+` WHERE uid = ?`, userId); err != nil {
		return fmt.Errorf("alias repo delete by uid: %w", err)
	}
	return nil
}

func (r *AliasRepository) MarkStatusByUserId(ctx context.Context, userId int64, status enum.UserAliasStatusType) error {
	query := `
		UPDATE ` + aliasTable + `
		SET alias_status = ?,
			modify_dt = ?
		WHERE uid = ?
	`

	if _, err := r.db.Exec(ctx, query, status, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("alias repo mark status by uid: %w", err)
	}
	return nil
}

func (r *AliasRepository) SoftDeleteByUserId(ctx context.Context, userId int64) error {
	query := `
		UPDATE ` + aliasTable + `
		SET alias_status = ?,
			status = ?,
			deleted_dt = ?
		WHERE uid = ?
	`

	if _, err := r.db.Exec(ctx, query, enum.UserAliasStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("alias repo soft delete by user id: %w", err)
	}
	return nil
}

// SoftDeleteByAliasId retires a single login key. Both status and
// deleted_dt are written because the active filter keys on those two, not
// on alias_status.
func (r *AliasRepository) SoftDeleteByAliasId(ctx context.Context, aliasId int64) error {
	query := `
		UPDATE ` + aliasTable + `
		SET alias_status = ?,
			status = ?,
			deleted_dt = ?
		WHERE alias_id = ?
	`

	if _, err := r.db.Exec(ctx, query, enum.UserAliasStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, aliasId); err != nil {
		return fmt.Errorf("alias repo soft delete by alias id: %w", err)
	}
	return nil
}

func ModelToDomainAlias(m *models.AliasModel) *user.Alias {
	a := user.NewAlias()
	a.SetId(m.Id)
	a.SetAliasId(m.AliasId)
	a.SetUserId(m.UserId)
	a.SetAka(m.Aka)
	a.SetAliasStatus(m.AliasStatus)
	a.SetRptFlg(m.RptFlg)
	a.SetKwords(m.Kwords)
	a.SetNote(m.Note)
	a.SetCreateId(m.CreateId)
	a.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	a.SetModifyId(m.ModifyId)
	a.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return a
}
