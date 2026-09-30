package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/role"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	roleTable = "ma_roles"

	roleColumns = `r.role_id, r.role_code, r.role_name, r.description, r.role_image_key,
		r.role_status, r.rpt_flg, r.kwords, r.note, r.status,
		r.create_id, r.create_dt, r.modify_id, r.modify_dt`

	roleActiveWhere = `r.status IN (?) AND r.deleted_dt IS NULL`
)

func roleActiveArgs() []any {
	return []any{enum.StatusActive}
}

type RoleRepository struct {
	db database.Executor
}

func NewRoleRepository(db database.Executor) role.IRepository {
	return &RoleRepository{db: db}
}

func scanRole(s database.RowScanner) (*models.RoleModel, error) {
	var m models.RoleModel
	if err := s.Scan(&m.RoleId, &m.RoleCode, &m.RoleName, &m.Description, &m.RoleImageKey,
		&m.RoleStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; roleActiveWhere is appended last so soft-deleted and
// inactive rows never come back.
func (r *RoleRepository) findOneBy(ctx context.Context, where string, args ...any) (*role.Role, error) {
	fullArgs := slices.Concat(args, roleActiveArgs())
	query := `SELECT ` + roleColumns + ` FROM ` + roleTable + ` r WHERE (` +
		where + `) AND ` + roleActiveWhere

	m, err := scanRole(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("role repo find (%s): %w", where, err)
	}
	return ModelToDomainRole(m), nil
}

func (r *RoleRepository) FindByRoleId(ctx context.Context, roleId int64) (*role.Role, error) {
	return r.findOneBy(ctx, "r.role_id = ?", roleId)
}

func (r *RoleRepository) FindByRoleCode(ctx context.Context, roleCode string) (*role.Role, error) {
	return r.findOneBy(ctx, "r.role_code = ?", roleCode)
}

func (r *RoleRepository) ListRoles(ctx context.Context, params *role.ListRolesParams) ([]*role.Role, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildRoleListFilterClause(params)

	countArgs := slices.Concat(filterArgs, roleActiveArgs())
	countQuery := `SELECT COUNT(*) FROM ` + roleTable + ` r` +
		whereActive(filterWhere, roleActiveWhere)

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("role repo count: %w", err)
	}

	page, limit := int64(1), int64(20)
	if params != nil {
		page, limit = params.Page, params.Limit
	}
	pg := pagination.NewPagination(page, limit, total)

	listArgs := slices.Concat(filterArgs, roleActiveArgs(), []any{pg.Size, pg.Skip})
	query := `SELECT ` + roleColumns + ` FROM ` + roleTable + ` r` +
		whereActive(filterWhere, roleActiveWhere) +
		` ORDER BY r.role_code ASC, r.role_id ASC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("role repo list: %w", err)
	}
	defer rows.Close()

	var roles []*role.Role
	for rows.Next() {
		m, err := scanRole(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("role repo scan row: %w", err)
		}
		roles = append(roles, ModelToDomainRole(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("role repo rows iteration: %w", err)
	}
	return roles, pg, nil
}

// buildRoleListFilterClause appends the optional narrowing predicates, each
// guarded so placeholders and args stay in step.
func buildRoleListFilterClause(params *role.ListRolesParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause string
		args   []any
	)
	if params.Search != nil {
		if needle := strings.TrimSpace(*params.Search); needle != "" {
			clause += ` AND (r.role_code LIKE ? OR r.role_name LIKE ?)`
			like := "%" + needle + "%"
			args = append(args, like, like)
		}
	}
	if params.RoleStatus != nil && strings.TrimSpace(*params.RoleStatus) != "" {
		clause += ` AND r.role_status = ?`
		args = append(args, strings.TrimSpace(*params.RoleStatus))
	}
	if len(params.RoleIds) > 0 {
		placeholders := make([]string, len(params.RoleIds))
		for i, id := range params.RoleIds {
			placeholders[i] = "?"
			args = append(args, id)
		}
		clause += ` AND r.role_id IN (` + strings.Join(placeholders, ",") + `)`
	}
	return clause, args
}

func (r *RoleRepository) Create(ctx context.Context, ro *role.Role) (*role.Role, error) {
	query := `
		INSERT INTO ` + roleTable + `
			(role_id, role_code, role_name, description, role_image_key, role_status,
			 rpt_flg, kwords, note, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := mtime.Now().Time
	_, err := r.db.Exec(ctx, query,
		ro.RoleId(), ro.RoleCode(), ro.RoleName(), ro.Description(), ro.RoleImageKey(), ro.RoleStatus(),
		ro.RptFlg(), ro.Kwords(), ro.Note(), ro.CreateId(), now, now)
	if err != nil {
		if isDuplicateEntry(err) {
			return nil, fmt.Errorf("role repo create: %w", role.ErrDuplicateRoleCode)
		}
		return nil, fmt.Errorf("role repo create: %w", err)
	}
	return r.FindByRoleId(ctx, ro.RoleId())
}

// Update applies a partial update: nullable columns use COALESCE(?, col) so a
// nil pointer leaves them unchanged, and role_name (NOT NULL) is skipped when
// empty. role_code is deliberately not updatable — it is the role's
// identity, and a future permission model will reference it.
func (r *RoleRepository) Update(ctx context.Context, ro *role.Role) error {
	var roleNameArg any
	if ro.RoleName() != "" {
		roleNameArg = ro.RoleName()
	}

	query := `
		UPDATE ` + roleTable + `
		SET role_name   = COALESCE(?, role_name),
			description = COALESCE(?, description),
			role_image_key = COALESCE(?, role_image_key),
			role_status = COALESCE(?, role_status),
			rpt_flg     = COALESCE(?, rpt_flg),
			kwords      = COALESCE(?, kwords),
			note        = COALESCE(?, note),
			modify_id   = COALESCE(?, modify_id),
			modify_dt   = ?
		WHERE role_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		roleNameArg, ro.Description(), ro.RoleImageKey(), ro.RoleStatus(),
		ro.RptFlg(), ro.Kwords(), ro.Note(),
		ro.ModifyId(), mtime.Now().Time, ro.RoleId()); err != nil {
		return fmt.Errorf("role repo update: %w", err)
	}
	return nil
}

func (r *RoleRepository) ClearRoleImageKey(ctx context.Context, roleId int64) error {
	query := `
		UPDATE ` + roleTable + `
		SET role_image_key = NULL,
			modify_dt      = ?
		WHERE role_id = ?
	`
	if _, err := r.db.Exec(ctx, query, mtime.Now().Time, roleId); err != nil {
		return fmt.Errorf("role repo clear image key: %w", err)
	}
	return nil
}

// SoftDeleteByRoleId stamps deleted_dt, which also releases the role_code
// from uk_live_role_code so the code can be created again.
func (r *RoleRepository) SoftDeleteByRoleId(ctx context.Context, roleId int64) error {
	query := `
		UPDATE ` + roleTable + `
		SET role_status = ?,
			status      = ?,
			deleted_dt  = ?,
			modify_dt   = ?
		WHERE role_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.RoleRecordStatusDeleted, enum.StatusInactive, now, now, roleId); err != nil {
		return fmt.Errorf("role repo soft delete: %w", err)
	}
	return nil
}

func (r *RoleRepository) ForceDeleteByRoleId(ctx context.Context, roleId int64) error {
	query := `DELETE FROM ` + roleTable + ` WHERE role_id = ?`
	if _, err := r.db.Exec(ctx, query, roleId); err != nil {
		return fmt.Errorf("role repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainRole(m *models.RoleModel) *role.Role {
	ro := role.NewRole()
	ro.SetRoleId(m.RoleId)
	ro.SetRoleCode(m.RoleCode)
	ro.SetRoleName(m.RoleName)
	ro.SetDescription(m.Description)
	ro.SetRoleImageKey(m.RoleImageKey)
	ro.SetRoleStatus(m.RoleStatus)
	ro.SetRptFlg(m.RptFlg)
	ro.SetKwords(m.Kwords)
	ro.SetNote(m.Note)
	ro.SetStatus(m.Status)
	ro.SetCreateId(m.CreateId)
	ro.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	ro.SetModifyId(m.ModifyId)
	ro.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return ro
}
