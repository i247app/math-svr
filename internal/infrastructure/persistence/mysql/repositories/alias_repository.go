package repositories

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"

	"math-ai.com/math-ai/internal/domain/shared/time"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

const (
	aliasTable = "ma_aliases"

	aliasColumns = `id, alias_id, user_id, aka, alias_status, note, create_id, create_dt, modify_id, modify_dt`

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
	if err := s.Scan(&m.Id, &m.AliasId, &m.UserId, &m.Aka, &m.AliasStatus, &m.Note, &m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *AliasRepository) findOneBy(ctx context.Context, where string, args ...any) (*user.Alias, error) {
	query := `SELECT ` + aliasColumns + ` FROM ` + aliasTable + ` WHERE ` + where
	m, err := scanAlias(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, err
	}
	return ModelToDomainAlias(m), nil
}

func (r *AliasRepository) Create(ctx context.Context, alias *user.Alias) (*user.Alias, error) {
	query := `
		INSERT INTO ` + aliasTable + ` (alias_id, user_id, aka, alias_status, note)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query, alias.AliasId(), alias.UserId(), alias.Aka(), alias.AliasStatus(), alias.Note())
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
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
	query := `SELECT ` + aliasColumns + ` FROM ` + aliasTable + ` WHERE user_id = ?`
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []*user.Alias
	for rows.Next() {
		m, err := scanAlias(rows)
		if err != nil {
			return nil, err
		}
		aliases = append(aliases, ModelToDomainAlias(m))
	}
	return aliases, nil
}

func (r *AliasRepository) UpdateByAliasId(ctx context.Context, alias *user.Alias) error {
	query := `
		UPDATE ` + aliasTable + `
		SET aka = COALESCE(?, aka),
			alias_status = COALESCE(?, alias_status),
			note = COALESCE(?, note),
			modify_id = COALESCE(?, modify_id),
			modify_dt = COALESCE(?, modify_dt)
		WHERE alias_id = ?
	`

	if _, err := r.db.Exec(ctx, query, alias.Aka(), alias.AliasStatus(), alias.Note(), alias.ModifyId(), alias.ModifyDt(), alias.AliasId()); err != nil {
		return fmt.Errorf("alias repo update by alias id: %w", err)
	}
	return nil
}

func (r *AliasRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM `+aliasTable+` WHERE user_id = ?`, userId); err != nil {
		return fmt.Errorf("alias repo delete by uid: %w", err)
	}
	return nil
}

func (r *AliasRepository) MarkStatusByUserId(ctx context.Context, userId int64, status enum.UserAliasStatusType) error {
	query := `
		UPDATE ` + aliasTable + `
		SET alias_status = ?,
			modify_dt = ?
		WHERE user_id = ?
	`

	if _, err := r.db.Exec(ctx, query, status, time.Now().Time, userId); err != nil {
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
		WHERE user_id = ?
	`

	if _, err := r.db.Exec(ctx, query, enum.UserAliasStatusTypeDeleted, enum.StatusInactive, time.Now().Time, userId); err != nil {
		return fmt.Errorf("alias repo soft delete by user id: %w", err)
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
	a.SetNote(m.Note)
	a.SetCreateId(m.CreateId)
	a.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	a.SetModifyId(m.ModifyId)
	a.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return a
}
