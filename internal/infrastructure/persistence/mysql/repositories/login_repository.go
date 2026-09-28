package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/domain/login"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	loginTable = "ma_logins"

	loginColumns = `lg.login_id, lg.uid, lg.upw, lg.logins_status, lg.rpt_flg,
		lg.kwords, lg.note, lg.status, lg.create_id, lg.create_dt,
		lg.modify_id, lg.modify_dt`

	loginActiveWhere = `lg.status IN (?) AND lg.deleted_dt IS NULL`
)

func loginActiveArgs() []any {
	return []any{enum.StatusActive}
}

type LoginRepository struct {
	db database.Executor
}

func NewLoginRepository(db database.Executor) login.IRepository {
	return &LoginRepository{db: db}
}

func scanLogin(s database.RowScanner) (*models.LoginModel, error) {
	var m models.LoginModel
	if err := s.Scan(&m.LoginId, &m.Uid, &m.Upw, &m.LoginsStatus, &m.RptFlg,
		&m.Kwords, &m.Note, &m.Status, &m.CreateId, &m.CreateDt,
		&m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *LoginRepository) findOneBy(ctx context.Context, where string, args ...any) (*login.Login, error) {
	fullArgs := slices.Concat(args, loginActiveArgs())
	query := `SELECT ` + loginColumns + ` FROM ` + loginTable + ` lg WHERE (` +
		where + `) AND ` + loginActiveWhere

	m, err := scanLogin(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("login repo find (%s): %w", where, err)
	}
	return ModelToDomainLogin(m), nil
}

func (r *LoginRepository) FindByLoginId(ctx context.Context, loginId int64) (*login.Login, error) {
	return r.findOneBy(ctx, "lg.login_id = ?", loginId)
}

func (r *LoginRepository) FindByUid(ctx context.Context, uid int64) (*login.Login, error) {
	return r.findOneBy(ctx, "lg.uid = ?", uid)
}

func (r *LoginRepository) Create(ctx context.Context, l *login.Login) (*login.Login, error) {
	query := `
		INSERT INTO ` + loginTable + `
			(login_id, uid, upw, logins_status, rpt_flg, kwords, note, status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := mtime.Now().Time
	_, err := r.db.Exec(ctx, query,
		l.LoginId(), l.Uid(), l.Upw(), l.LoginsStatus(), l.RptFlg(), l.Kwords(),
		l.Note(), l.Status(), l.CreateId(), now, now)
	if err != nil {
		return nil, fmt.Errorf("login repo create: %w", err)
	}

	return r.FindByLoginId(ctx, l.LoginId())
}

func ModelToDomainLogin(m *models.LoginModel) *login.Login {
	l := login.NewLogin()
	l.SetLoginId(m.LoginId)
	l.SetUid(m.Uid)
	l.SetUpw(m.Upw)
	l.SetLoginsStatus(m.LoginsStatus)
	l.SetRptFlg(m.RptFlg)
	l.SetKwords(m.Kwords)
	l.SetNote(m.Note)
	l.SetStatus(m.Status)
	l.SetCreateId(m.CreateId)
	l.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	l.SetModifyId(m.ModifyId)
	l.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return l
}
