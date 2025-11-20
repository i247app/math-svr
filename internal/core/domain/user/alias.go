package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type Alias struct {
	id        string
	uid       string
	aka       string
	status    string
	createDT  time.Time
	modifyDT  time.Time
	deletedDT *time.Time
}

func NewAliasDomain() *Alias {
	return &Alias{}
}

func (u *Alias) ID() string {
	return u.id
}

func (u *Alias) GenerateID() {
	u.id = uuid.New().String()
}

func (a *Alias) UID() string {
	return a.uid
}

func (a *Alias) SetUID(uid string) {
	a.uid = uid
}

func (a *Alias) Aka() string {
	return a.aka
}

func (a *Alias) SetAka(aka string) {
	a.aka = aka
}

func (a *Alias) Status() string {
	return a.status
}

func (a *Alias) SetStatus(status string) {
	a.status = status
}

func (a *Alias) CreateDT() time.Time {
	return a.createDT
}

func (a *Alias) SetCreateDT(createDT time.Time) {
	a.createDT = createDT
}

func (a *Alias) ModifyDT() time.Time {
	return a.modifyDT
}

func (a *Alias) SetModifyDT(modifyDT time.Time) {
	a.modifyDT = modifyDT
}

func (a *Alias) DeletedDT() *time.Time {
	return a.deletedDT
}

func BuildAliasDomainFromModel(model *models.AliasUserModel) *Alias {
	return &Alias{
		id:       model.ID,
		uid:      model.UID,
		aka:      model.Aka,
		status:   model.Status,
		createDT: model.CreateDT,
		modifyDT: model.ModifyDT,
	}
}
