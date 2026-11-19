package role

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Role models ma_roles — one entry in the permission module's role registry.
// roleCode is the stable machine key (upper-case, e.g. "TEACHER") and never
// changes after creation; roleName / description are display text.
//
// Nothing consults this registry for authorization yet: ma_users.role and
// ma_profiles.role remain free-text columns validated by enum.RoleType.
type Role struct {
	audit.Audit

	roleId      int64
	roleCode    string
	roleName    string
	description *string
	imageKey    *string
	roleStatus  *string
}

func NewRole() *Role {
	return &Role{}
}

func (r *Role) RoleId() int64      { return r.roleId }
func (r *Role) SetRoleId(id int64) { r.roleId = id }

func (r *Role) RoleCode() string     { return r.roleCode }
func (r *Role) SetRoleCode(c string) { r.roleCode = c }

func (r *Role) RoleName() string     { return r.roleName }
func (r *Role) SetRoleName(n string) { r.roleName = n }

func (r *Role) Description() *string     { return r.description }
func (r *Role) SetDescription(d *string) { r.description = d }

func (r *Role) RoleImageKey() *string     { return r.imageKey }
func (r *Role) SetRoleImageKey(k *string) { r.imageKey = k }

func (r *Role) RoleStatus() *string     { return r.roleStatus }
func (r *Role) SetRoleStatus(v *string) { r.roleStatus = v }
