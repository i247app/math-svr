package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/user"

	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"

	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// Table name
const (
	userTable = "ma_users"

	userColumns = `u.uid, u.name, u.phone, u.email, u.is_email_verified, u.avatar_key, u.role, u.identity_code, u.user_status, u.status,
	u.rpt_flg, u.kwords, u.note, u.create_id, u.create_dt, u.modify_id, u.modify_dt`

	userFrom = userTable + ` u`

	// User reads never JOIN ma_aliases. The alias table is a separate
	// aggregate (login-key registry) owned by AliasRepository; resolving a
	// login name to a user is composed in the application layer
	// (alias.FindByAka -> user.FindByUserId), not here.
	userActiveWhere = `u.status IN (?) AND u.deleted_dt IS NULL`
)

func userActiveArgs() []any {
	return []any{enum.StatusActive}
}

type UserRepository struct {
	db database.Executor
}

func NewUserRepository(db database.Executor) user.IRepository {
	return &UserRepository{db: db}
}

func scanUser(s database.RowScanner) (*models.UserModel, error) {
	var m models.UserModel
	if err := s.Scan(&m.UserId, &m.UserName, &m.Phone, &m.Email, &m.IsEmailVerified, &m.AvatarKey, &m.Role, &m.IdentityCode, &m.UserStatus, &m.Status,
		&m.RptFlg, &m.Kwords, &m.Note, &m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy runs a single-row lookup. `where` is a package-controlled SQL
// fragment (never user input); args supply the placeholder values.
// userActiveWhere is appended last so every read excludes system-INACTIVE
// and soft-deleted rows.
func (r *UserRepository) findOneBy(ctx context.Context, where string, args ...any) (*user.User, error) {
	fullArgs := slices.Concat(args, userActiveArgs())
	query := `SELECT ` + userColumns + ` FROM ` + userFrom +
		` WHERE (` + where + `) AND ` + userActiveWhere

	m, err := scanUser(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo find (%s): %w", where, err)
	}
	return ModelToDomain(m), nil
}

// func (r *UserRepository) FindById(ctx context.Context, id int64) (*user.User, error) {
// 	return r.findOneBy(ctx, "u.uid = ?", id)
// }

func (r *UserRepository) FindByUserId(ctx context.Context, userId int64) (*user.User, error) {
	return r.findOneBy(ctx, "u.uid = ?", userId)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	return r.findOneBy(ctx, "u.email = ?", email)
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*user.User, error) {
	return r.findOneBy(ctx, "u.phone = ?", phone)
}

func (r *UserRepository) FindByUserName(ctx context.Context, userName string) (*user.User, error) {
	return r.findOneBy(ctx, "u.name = ?", userName)
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	query := `
		INSERT INTO ` + userTable + ` (uid, name, phone, email, is_email_verified, avatar_key, role, identity_code, user_status, rpt_flg, kwords, note, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(ctx, query, u.UserId(), u.UserName(), u.Phone(), u.Email(), u.IsEmailVerified(), u.AvatarKey(), u.Role(), u.IdentityCode(), u.UserStatus(), u.RptFlg(), u.Kwords(), u.Note(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("user repo create: %w", err)
	}

	return r.FindByUserId(ctx, u.UserId())
}

func (r *UserRepository) ListUsers(ctx context.Context, params *user.ListUsersParams) ([]*user.User, *pagination.Pagination, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + userFrom +
		` WHERE ` + userActiveWhere
	if err := r.db.QueryRow(ctx, countQuery, userActiveArgs()...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("user repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := slices.Concat(userActiveArgs(), []any{pg.Size, pg.Skip})
	query := `SELECT ` + userColumns + ` FROM ` + userFrom +
		` WHERE ` + userActiveWhere +
		` ORDER BY u.uid DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("user repo list: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		m, err := scanUser(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("user repo scan row: %w", err)
		}
		users = append(users, ModelToDomain(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("user repo rows iteration: %w", err)
	}

	return users, pg, nil
}

func (r *UserRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM `+userTable+` WHERE uid = ?`, userId); err != nil {
		return fmt.Errorf("user repo delete by user id: %w", err)
	}
	return nil
}

// Update patches a user row. Most columns use COALESCE so a nil from the
// domain means "leave it alone", but is_email_verified is written
// DIRECTLY: false is a real value there, not an absence, and COALESCE
// would make it impossible to ever un-verify an address. Callers must
// therefore load the row, change what they mean to change, and pass the
// whole aggregate — which both of them do.
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	// name is NOT NULL in ma_users, so an empty string from the
	// domain means "do not change" — pass NULL into COALESCE so the
	// existing value is preserved. Callers that intentionally want to
	// overwrite must set a non-empty name.
	var userName any
	if u.UserName() != "" {
		userName = u.UserName()
	}

	query := `
		UPDATE ` + userTable + `
		SET
			name = COALESCE(?, name),
			email = COALESCE(?, email),
			phone = COALESCE(?, phone),
			avatar_key = COALESCE(?, avatar_key),
			role = COALESCE(?, role),
			identity_code = COALESCE(?, identity_code),
			is_email_verified = ?,
			rpt_flg = COALESCE(?, rpt_flg),
			kwords = COALESCE(?, kwords)
		WHERE uid = ?
	`

	if _, err := r.db.Exec(ctx, query, userName, u.Email(), u.Phone(), u.AvatarKey(), u.Role(), u.IdentityCode(), u.IsEmailVerified(), u.RptFlg(), u.Kwords(), u.UserId()); err != nil {
		return fmt.Errorf("user repo update: %w", err)
	}
	return nil
}

// UpdateAvatarKey persists a newly-uploaded avatar's S3 key onto a user
// row. Split out from the general Update path so the multipart upload
// flow doesn't need to materialise the rest of the User aggregate just
// to write one column.
func (r *UserRepository) UpdateAvatarKey(ctx context.Context, userId int64, avatarKey string) error {
	query := `UPDATE ` + userTable + ` SET avatar_key = ? WHERE uid = ?`
	if _, err := r.db.Exec(ctx, query, avatarKey, userId); err != nil {
		return fmt.Errorf("user repo update avatar key: %w", err)
	}
	return nil
}

func (r *UserRepository) MarkStatusByUserId(ctx context.Context, userId int64, status enum.UserStatusType) error {
	query := `
		UPDATE ` + userTable + `
		SET user_status = ?,
			modify_dt = ?
		WHERE uid = ?
	`

	if _, err := r.db.Exec(ctx, query, status, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("user repo mark status by user id: %w", err)
	}
	return nil
}

func (r *UserRepository) SoftDeleteByUserId(ctx context.Context, userId int64) error {
	query := `
		UPDATE ` + userTable + `
		SET user_status = ?,
			status = ?,
			deleted_dt = ?
		WHERE uid = ?
	`

	if _, err := r.db.Exec(ctx, query, enum.UserStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("user repo soft delete by user id: %w", err)
	}
	return nil
}

func DomainToModel(u *user.User) *models.UserModel {
	return &models.UserModel{
		UserId:     u.UserId(),
		UserName:   u.UserName(),
		Email:      u.Email(),
		Phone:      u.Phone(),
		AvatarKey:  u.AvatarKey(),
		Role:       u.Role(),
		UserStatus: u.UserStatus(),
		Status:     u.Status(),
		RptFlg:     u.RptFlg(),
		Kwords:     u.Kwords(),
		Note:       u.Note(),
		CreateId:   u.CreateId(),
		CreateDt:   u.CreateDt().ToTime(),
		ModifyId:   u.ModifyId(),
		ModifyDt:   u.ModifyDt().ToTime(),
	}
}

func ModelToDomain(m *models.UserModel) *user.User {
	u := user.NewUser()
	u.SetUserId(m.UserId)
	u.SetUserName(m.UserName)
	u.SetEmail(m.Email)
	u.SetIsEmailVerified(m.IsEmailVerified)
	u.SetPhone(m.Phone)
	u.SetAvatarKey(m.AvatarKey)
	u.SetRole(m.Role)
	u.SetIdentityCode(m.IdentityCode)
	u.SetUserStatus(m.UserStatus)
	u.SetStatus(m.Status)
	u.SetRptFlg(m.RptFlg)
	u.SetKwords(m.Kwords)
	u.SetNote(m.Note)
	u.SetCreateId(m.CreateId)
	u.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	u.SetModifyId(m.ModifyId)
	u.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})

	return u
}
