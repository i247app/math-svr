package user

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type Alias struct {
	id          int64
	aliasId     string
	userId      string
	aka         string
	aliasStatus *string
	note        *string
	status      string
	createId    *string
	createDt    mtime.MathTime
	modifyId    *string
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

func (a *Alias) AliasId() string {
	return a.aliasId
}

func (a *Alias) SetAliasId(aliasId string) {
	a.aliasId = aliasId
}

func (a *Alias) UserId() string {
	return a.userId
}

func (a *Alias) SetUserId(userId string) {
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

func (a *Alias) CreateId() *string {
	return a.createId
}

func (a *Alias) SetCreateId(createId *string) {
	a.createId = createId
}

func (a *Alias) CreateDt() mtime.MathTime {
	return a.createDt
}

func (a *Alias) SetCreateDt(createDt mtime.MathTime) {
	a.createDt = createDt
}

func (a *Alias) ModifyId() *string {
	return a.modifyId
}

func (a *Alias) SetModifyId(modifyId *string) {
	a.modifyId = modifyId
}

func (a *Alias) ModifyDt() mtime.MathTime {
	return a.modifyDt
}

func (a *Alias) SetModifyDt(modifyDt mtime.MathTime) {
	a.modifyDt = modifyDt
}
