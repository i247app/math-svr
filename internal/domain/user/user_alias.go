package user

import (
	"github.com/google/uuid"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type Alias struct {
	id          int64
	aliasId     uuid.UUID
	userId      uuid.UUID
	aka         string
	aliasStatus *string
	note        *string
	status      string
	createId    *uuid.UUID
	createDt    mtime.MathTime
	modifyId    *uuid.UUID
	modifyDt    mtime.MathTime
}

func NewAlias() *Alias {
	return &Alias{}
}

func (a *Alias) Id() int64 {
	return a.id
}

func (a *Alias) SetId(id int64) {
	a.id = id
}

func (a *Alias) AliasId() uuid.UUID {
	return a.aliasId
}

func (a *Alias) SetAliasId(aliasId uuid.UUID) {
	a.aliasId = aliasId
}

func (a *Alias) UserId() uuid.UUID {
	return a.userId
}

func (a *Alias) SetUserId(userId uuid.UUID) {
	a.userId = userId
}

func (a *Alias) Aka() string {
	return a.aka
}

func (a *Alias) SetAka(aka string) {
	a.aka = aka
}

func (a *Alias) AliasStatus() *string {
	return a.aliasStatus
}

func (a *Alias) SetAliasStatus(aliasStatus *string) {
	a.aliasStatus = aliasStatus
}

func (a *Alias) Note() *string {
	return a.note
}

func (a *Alias) SetNote(note *string) {
	a.note = note
}

func (a *Alias) Status() string {
	return a.status
}

func (a *Alias) SetStatus(status string) {
	a.status = status
}

func (a *Alias) CreateId() *uuid.UUID {
	return a.createId
}

func (a *Alias) SetCreateId(createId *uuid.UUID) {
	a.createId = createId
}

func (a *Alias) CreateDt() mtime.MathTime {
	return a.createDt
}

func (a *Alias) SetCreateDt(createDt mtime.MathTime) {
	a.createDt = createDt
}

func (a *Alias) ModifyId() *uuid.UUID {
	return a.modifyId
}

func (a *Alias) SetModifyId(modifyId *uuid.UUID) {
	a.modifyId = modifyId
}

func (a *Alias) ModifyDt() mtime.MathTime {
	return a.modifyDt
}

func (a *Alias) SetModifyDt(modifyDt mtime.MathTime) {
	a.modifyDt = modifyDt
}
