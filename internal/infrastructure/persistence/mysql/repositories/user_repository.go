package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Table name
const (
	userTable = "ma_users"

	userColumns = `u.id, u.user_id, u.name, u.phone, u.email, u.avatar_key, u.role, u.user_status, u.status,
	u.note, u.create_id, u.create_dt, u.modify_id, u.modify_dt`

	userFromJoin = userTable + ` u
	JOIN ma_aliases a ON u.user_id = a.user_id`

	userActiveWhere = `
	u.status IN (?) AND u.deleted_dt IS NULL
	AND a.status IN (?) AND a.deleted_dt IS NULL`

	userBareActiveWhere = `u.status IN (?) AND u.deleted_dt IS NULL`
)

func userActiveArgs() []any {
	return []any{
		enum.StatusActive,
		enum.StatusActive,
	}
}

func userBareActiveArgs() []any {
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
	if err := s.Scan(&m.Id, &m.UserId, &m.UserName, &m.Phone, &m.Email, &m.AvatarKey, &m.Role, &m.UserStatus, &m.Status,
		&m.Note, &m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy runs a single-row lookup. `where` is a package-controlled SQL
// fragment (never user input); args supply the placeholder values. The query
// uses SELECT DISTINCT because userFromJoin can multiply rows via aliases.
// userActiveWhere is automatically prepended so every read excludes
// system-INACTIVE rows and any row with a *_status of DELETED.
func (r *UserRepository) findOneBy(ctx context.Context, where string, args ...any) (*user.User, error) {
	fullArgs := append(userActiveArgs(), args...)
	query := `SELECT DISTINCT ` + userColumns + ` FROM ` + userFromJoin +
		` WHERE ` + userActiveWhere + ` AND (` + where + `)`

	m, err := scanUser(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo find (%s): %w", where, err)
	}
	return ModelToDomain(m), nil
}

func (r *UserRepository) findBareById(ctx context.Context, id int64) (*user.User, error) {
	args := append(userBareActiveArgs(), id)
	query := `SELECT ` + userColumns + ` FROM ` + userTable + ` u WHERE ` +
		userBareActiveWhere + ` AND u.id = ?`

	m, err := scanUser(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repo find bare by id: %w", err)
	}
	return ModelToDomain(m), nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (*user.User, error) {
	return r.findOneBy(ctx, "u.id = ?", id)
}

func (r *UserRepository) FindByUserId(ctx context.Context, userId int64) (*user.User, error) {
	return r.findOneBy(ctx, "u.user_id = ?", userId)
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
		INSERT INTO ` + userTable + ` (user_id, name, phone, email, avatar_key, role, user_status, note, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(ctx, query, u.UserId(), u.UserName(), u.Phone(), u.Email(), u.AvatarKey(), u.Role(), u.UserStatus(), u.Note(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("user repo create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("user repo last insert id: %w", err)
	}

	// Hydrate via the bare `users` table — the alias/login rows don't exist
	// yet at this point in the CreateUser transaction, so the JOIN'd
	// FindById would return nil.
	return r.findBareById(ctx, id)
}

func (r *UserRepository) ListUsers(ctx context.Context, params *user.ListUsersParams) ([]*user.User, *pagination.Pagination, error) {
	var total int64
	countQuery := `SELECT COUNT(DISTINCT u.id) FROM ` + userFromJoin +
		` WHERE ` + userActiveWhere
	if err := r.db.QueryRow(ctx, countQuery, userActiveArgs()...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("user repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(userActiveArgs(), pg.Size, pg.Skip)
	query := `SELECT DISTINCT ` + userColumns + ` FROM ` + userFromJoin +
		` WHERE ` + userActiveWhere +
		` ORDER BY u.id DESC LIMIT ? OFFSET ?`
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

func (r *UserRepository) DeleteById(ctx context.Context, id int64) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM `+userTable+` WHERE id = ?`, id); err != nil {
		return fmt.Errorf("user repo delete by id: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM `+userTable+` WHERE user_id = ?`, userId); err != nil {
		return fmt.Errorf("user repo delete by user id: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	// name is NOT NULL in ma_users, so an empty string from the
	// domain means "do not change" — pass NULL into COALESCE so the
	// existing value is preserved. Callers that intentionally want to
	// overwrite must set a non-empty name.
	var userName any
	if u.UserName() != "" {
		userName = u.UserName()
	}

	// role is NOT NULL in ma_users — an empty string from the domain means
	// "do not change", so pass NULL into COALESCE to preserve the existing
	// value. Mirrors the user_name handling above and ma_profiles.role.
	var roleArg any
	if u.Role() != "" {
		roleArg = u.Role()
	}

	query := `
		UPDATE ` + userTable + `
		SET
			name = COALESCE(?, name),
			email = COALESCE(?, email),
			phone = COALESCE(?, phone),
			avatar_key = COALESCE(?, avatar_key),
			role = COALESCE(?, role)
		WHERE id = ?
	`

	if _, err := r.db.Exec(ctx, query, userName, u.Email(), u.Phone(), u.AvatarKey(), roleArg, u.Id()); err != nil {
		return fmt.Errorf("user repo update: %w", err)
	}
	return nil
}

// UpdateAvatarKey persists a newly-uploaded avatar's S3 key onto a user
// row. Split out from the general Update path so the multipart upload
// flow doesn't need to materialise the rest of the User aggregate just
// to write one column.
func (r *UserRepository) UpdateAvatarKey(ctx context.Context, userId int64, avatarKey string) error {
	query := `UPDATE ` + userTable + ` SET avatar_key = ? WHERE user_id = ?`
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
		WHERE user_id = ?
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
		WHERE user_id = ?
	`

	if _, err := r.db.Exec(ctx, query, enum.UserStatusTypeDeleted, enum.StatusInactive, mtime.Now().Time, userId); err != nil {
		return fmt.Errorf("user repo soft delete by user id: %w", err)
	}
	return nil
}

func DomainToModel(u *user.User) *models.UserModel {
	return &models.UserModel{
		Id:         u.Id(),
		UserId:     u.UserId(),
		UserName:   u.UserName(),
		Email:      u.Email(),
		Phone:      u.Phone(),
		AvatarKey:  u.AvatarKey(),
		Role:       u.Role(),
		UserStatus: u.UserStatus(),
		Status:     u.Status(),
		Note:       u.Note(),
		CreateId:   u.CreateId(),
		CreateDt:   u.CreateDt().ToTime(),
		ModifyId:   u.ModifyId(),
		ModifyDt:   u.ModifyDt().ToTime(),
	}
}

func ModelToDomain(m *models.UserModel) *user.User {
	u := user.NewUser()
	u.SetId(m.Id)
	u.SetUserId(m.UserId)
	u.SetUserName(m.UserName)
	u.SetEmail(m.Email)
	u.SetPhone(m.Phone)
	u.SetAvatarKey(m.AvatarKey)
	u.SetRole(m.Role)
	u.SetUserStatus(m.UserStatus)
	u.SetStatus(m.Status)
	u.SetNote(m.Note)
	u.SetCreateId(m.CreateId)
	u.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	u.SetModifyId(m.ModifyId)
	u.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})

	return u
}
