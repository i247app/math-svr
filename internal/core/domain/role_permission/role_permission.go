package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type RolePermission struct {
	id           string
	roleID       string
	permissionID string
	createID     *string
	createDT     time.MathTime
	modifyID     *string
	modifyDT     time.MathTime
	deletedDT    *time.MathTime
}

func NewRolePermissionDomain() *RolePermission {
	return &RolePermission{}
}

func (rp *RolePermission) ID() string {
	return rp.id
}

func (rp *RolePermission) GenerateID() {
	rp.id = uuid.New().String()
}

func (rp *RolePermission) SetID(id string) {
	rp.id = id
}

func (rp *RolePermission) RoleID() string {
	return rp.roleID
}

func (rp *RolePermission) SetRoleID(roleID string) {
	rp.roleID = roleID
}

func (rp *RolePermission) PermissionID() string {
	return rp.permissionID
}

func (rp *RolePermission) SetPermissionID(permissionID string) {
	rp.permissionID = permissionID
}

func (rp *RolePermission) CreateID() *string {
	return rp.createID
}

func (rp *RolePermission) SetCreateID(createID *string) {
	rp.createID = createID
}

func (rp *RolePermission) CreatedAt() time.MathTime {
	return rp.createDT
}

func (rp *RolePermission) SetCreatedAt(createDT time.MathTime) {
	rp.createDT = createDT
}

func (rp *RolePermission) ModifyID() *string {
	return rp.modifyID
}

func (rp *RolePermission) SetModifyID(modifyID *string) {
	rp.modifyID = modifyID
}

func (rp *RolePermission) ModifiedAt() time.MathTime {
	return rp.modifyDT
}

func (rp *RolePermission) SetModifiedAt(modifyDT time.MathTime) {
	rp.modifyDT = modifyDT
}

func (rp *RolePermission) DeletedAt() *time.MathTime {
	return rp.deletedDT
}

func (rp *RolePermission) SetDeletedAt(deletedDT *time.MathTime) {
	rp.deletedDT = deletedDT
}

// BuildRolePermissionDomainFromModel builds a RolePermission from a model
func BuildRolePermissionDomainFromModel(model *models.RolePermissionModel) *RolePermission {
	return &RolePermission{
		id:           model.ID,
		roleID:       model.RoleID,
		permissionID: model.PermissionID,
		createID:     model.CreateID,
		createDT:     model.CreateDT,
		modifyID:     model.ModifyID,
		modifyDT:     model.ModifyDT,
		deletedDT:    model.DeletedDT,
	}
}
