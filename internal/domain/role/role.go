package role

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Role models ma_roles — one entry in the permission module's role registry.
// roleCode is the stable machine key (upper-case, e.g. "TEACHER") and never
// changes after creation; roleName / description are display text.
//
// Nothing consults this registry for authorization yet: ma_users.role and
// ma_profiles.role remain free-text columns validated by enum.RoleType.
type Role struct {
	roleId      int64
	roleCode    string
	roleName    string
	description *string
	imageKey    *string
	roleStatus  *string
	rptFlg      *string
	kwords      *string
	note        *string
	status      string
	createId    *int64
	createDt    mtime.MathTime
	modifyId    *int64
	modifyDt    mtime.MathTime
}

func NewRole() *Role {
	return &Role{}
}

func (r *Role) RoleId() int64                { return r.roleId }
func (r *Role) SetRoleId(id int64)           { r.roleId = id }
func (r *Role) RoleCode() string             { return r.roleCode }
func (r *Role) SetRoleCode(c string)         { r.roleCode = c }
func (r *Role) RoleName() string             { return r.roleName }
func (r *Role) SetRoleName(n string)         { r.roleName = n }
func (r *Role) Description() *string         { return r.description }
func (r *Role) SetDescription(d *string)     { r.description = d }
func (r *Role) RoleImageKey() *string        { return r.imageKey }
func (r *Role) SetRoleImageKey(k *string)    { r.imageKey = k }
func (r *Role) RoleStatus() *string          { return r.roleStatus }
func (r *Role) SetRoleStatus(v *string)      { r.roleStatus = v }
func (r *Role) RptFlg() *string              { return r.rptFlg }
func (r *Role) SetRptFlg(v *string)          { r.rptFlg = v }
func (r *Role) Kwords() *string              { return r.kwords }
func (r *Role) SetKwords(v *string)          { r.kwords = v }
func (r *Role) Note() *string                { return r.note }
func (r *Role) SetNote(n *string)            { r.note = n }
func (r *Role) Status() string               { return r.status }
func (r *Role) SetStatus(v string)           { r.status = v }
func (r *Role) CreateId() *int64             { return r.createId }
func (r *Role) SetCreateId(id *int64)        { r.createId = id }
func (r *Role) CreateDt() mtime.MathTime     { return r.createDt }
func (r *Role) SetCreateDt(t mtime.MathTime) { r.createDt = t }
func (r *Role) ModifyId() *int64             { return r.modifyId }
func (r *Role) SetModifyId(id *int64)        { r.modifyId = id }
func (r *Role) ModifyDt() mtime.MathTime     { return r.modifyDt }
func (r *Role) SetModifyDt(t mtime.MathTime) { r.modifyDt = t }
