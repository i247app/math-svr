package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type Role struct {
	id           string
	name         string
	code         string
	description  *string
	parentRoleID *string
	isSystemRole bool
	status       string
	displayOrder int8
	createID     *string
	createDT     time.MathTime
	modifyID     *string
	modifyDT     time.MathTime
	deletedDT    *time.MathTime
}

func NewRoleDomain() *Role {
	return &Role{}
}

func (r *Role) ID() string {
	return r.id
}

func (r *Role) GenerateID() {
	r.id = uuid.New().String()
}

func (r *Role) SetID(id string) {
	r.id = id
}

func (r *Role) Name() string {
	return r.name
}

func (r *Role) SetName(name string) {
	r.name = name
}

func (r *Role) Code() string {
	return r.code
}

func (r *Role) SetCode(code string) {
	r.code = code
}

func (r *Role) Description() *string {
	return r.description
}

func (r *Role) SetDescription(description *string) {
	r.description = description
}

func (r *Role) ParentRoleID() *string {
	return r.parentRoleID
}

func (r *Role) SetParentRoleID(parentRoleID *string) {
	r.parentRoleID = parentRoleID
}

func (r *Role) IsSystemRole() bool {
	return r.isSystemRole
}

func (r *Role) SetIsSystemRole(isSystemRole bool) {
	r.isSystemRole = isSystemRole
}

func (r *Role) Status() string {
	return r.status
}

func (r *Role) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	r.status = status
}

func (r *Role) DisplayOrder() int8 {
	return r.displayOrder
}

func (r *Role) SetDisplayOrder(displayOrder int8) {
	r.displayOrder = displayOrder
}

func (r *Role) CreateID() *string {
	return r.createID
}

func (r *Role) SetCreateID(createID *string) {
	r.createID = createID
}

func (r *Role) CreatedAt() time.MathTime {
	return r.createDT
}

func (r *Role) SetCreatedAt(createDT time.MathTime) {
	r.createDT = createDT
}

func (r *Role) ModifyID() *string {
	return r.modifyID
}

func (r *Role) SetModifyID(modifyID *string) {
	r.modifyID = modifyID
}

func (r *Role) ModifiedAt() time.MathTime {
	return r.modifyDT
}

func (r *Role) SetModifiedAt(modifyDT time.MathTime) {
	r.modifyDT = modifyDT
}

func (r *Role) DeletedAt() *time.MathTime {
	return r.deletedDT
}

func (r *Role) SetDeletedAt(deletedDT *time.MathTime) {
	r.deletedDT = deletedDT
}

// BuildRoleDomainFromModel builds a Role from a model
func BuildRoleDomainFromModel(model *models.RoleModel) *Role {
	return &Role{
		id:           model.ID,
		name:         model.Name,
		code:         model.Code,
		description:  model.Description,
		parentRoleID: model.ParentRoleID,
		isSystemRole: model.IsSystemRole,
		status:       model.Status,
		displayOrder: model.DisplayOrder,
		createID:     model.CreateID,
		createDT:     model.CreateDT,
		modifyID:     model.ModifyID,
		modifyDT:     model.ModifyDT,
		deletedDT:    model.DeletedDT,
	}
}
