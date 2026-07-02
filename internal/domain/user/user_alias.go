package user

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

type Alias struct {
	id          int64
	aliasId     int64
	userId      int64
	aka         string
	aliasStatus *string
	note        *string
	status      string
	createId    *int64
	createDt    mtime.MathTime
	modifyId    *int64
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

func (a *Alias) AliasId() int64 {
	return a.aliasId
}

func (a *Alias) SetAliasId(aliasId int64) {
	a.aliasId = aliasId
}

func (a *Alias) UserId() int64 {
	return a.userId
}

func (a *Alias) SetUserId(userId int64) {
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

func (a *Alias) CreateId() *int64 {
	return a.createId
}

func (a *Alias) SetCreateId(createId *int64) {
	a.createId = createId
}

func (a *Alias) CreateDt() mtime.MathTime {
	return a.createDt
}

func (a *Alias) SetCreateDt(createDt mtime.MathTime) {
	a.createDt = createDt
}

func (a *Alias) ModifyId() *int64 {
	return a.modifyId
}

func (a *Alias) SetModifyId(modifyId *int64) {
	a.modifyId = modifyId
}

func (a *Alias) ModifyDt() mtime.MathTime {
	return a.modifyDt
}

func (a *Alias) SetModifyDt(modifyDt mtime.MathTime) {
	a.modifyDt = modifyDt
}
