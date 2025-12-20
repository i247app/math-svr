package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type Permission struct {
	id           string
	name         string
	description  *string
	httpMethod   string
	endpointPath string
	resource     *string
	action       *string
	status       string
	createID     *string
	createDT     time.MathTime
	modifyID     *string
	modifyDT     time.MathTime
	deletedDT    *time.MathTime
}

func NewPermissionDomain() *Permission {
	return &Permission{}
}

func (p *Permission) ID() string {
	return p.id
}

func (p *Permission) GenerateID() {
	p.id = uuid.New().String()
}

func (p *Permission) SetID(id string) {
	p.id = id
}

func (p *Permission) Name() string {
	return p.name
}

func (p *Permission) SetName(name string) {
	p.name = name
}

func (p *Permission) Description() *string {
	return p.description
}

func (p *Permission) SetDescription(description *string) {
	p.description = description
}

func (p *Permission) HTTPMethod() string {
	return p.httpMethod
}

func (p *Permission) SetHTTPMethod(httpMethod string) {
	p.httpMethod = httpMethod
}

func (p *Permission) EndpointPath() string {
	return p.endpointPath
}

func (p *Permission) SetEndpointPath(endpointPath string) {
	p.endpointPath = endpointPath
}

func (p *Permission) Resource() *string {
	return p.resource
}

func (p *Permission) SetResource(resource *string) {
	p.resource = resource
}

func (p *Permission) Action() *string {
	return p.action
}

func (p *Permission) SetAction(action *string) {
	p.action = action
}

func (p *Permission) Status() string {
	return p.status
}

func (p *Permission) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	p.status = status
}

func (p *Permission) CreateID() *string {
	return p.createID
}

func (p *Permission) SetCreateID(createID *string) {
	p.createID = createID
}

func (p *Permission) CreatedAt() time.MathTime {
	return p.createDT
}

func (p *Permission) SetCreatedAt(createDT time.MathTime) {
	p.createDT = createDT
}

func (p *Permission) ModifyID() *string {
	return p.modifyID
}

func (p *Permission) SetModifyID(modifyID *string) {
	p.modifyID = modifyID
}

func (p *Permission) ModifiedAt() time.MathTime {
	return p.modifyDT
}

func (p *Permission) SetModifiedAt(modifyDT time.MathTime) {
	p.modifyDT = modifyDT
}

func (p *Permission) DeletedAt() *time.MathTime {
	return p.deletedDT
}

func (p *Permission) SetDeletedAt(deletedDT *time.MathTime) {
	p.deletedDT = deletedDT
}

// BuildPermissionDomainFromModel builds a Permission from a model
func BuildPermissionDomainFromModel(model *models.PermissionModel) *Permission {
	return &Permission{
		id:           model.ID,
		name:         model.Name,
		description:  model.Description,
		httpMethod:   model.HTTPMethod,
		endpointPath: model.EndpointPath,
		resource:     model.Resource,
		action:       model.Action,
		status:       model.Status,
		createID:     model.CreateID,
		createDT:     model.CreateDT,
		modifyID:     model.ModifyID,
		modifyDT:     model.ModifyDT,
		deletedDT:    model.DeletedDT,
	}
}
