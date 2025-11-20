package domain

import "time"

type Login struct {
	id        int64
	uid       int64
	hasspass  string
	status    string
	createDT  time.Time
	modifyDT  time.Time
	deletedDT *time.Time
}

func NewLoginDomain() *Login {
	return &Login{}
}

func (l *Login) ID() int64 {
	return l.id
}

func (l *Login) SetID(id int64) {
	l.id = id
}

func (l *Login) UID() int64 {
	return l.uid
}

func (l *Login) SetUID(uid int64) {
	l.uid = uid
}

func (l *Login) HassPass() string {
	return l.hasspass
}

func (l *Login) SetHassPass(hasspass string) {
	l.hasspass = hasspass
}

func (l *Login) Status() string {
	return l.status
}

func (l *Login) SetStatus(status string) {
	l.status = status
}

func (l *Login) CreateDT() time.Time {
	return l.createDT
}

func (l *Login) SetCreateDT(createDT time.Time) {
	l.createDT = createDT
}

func (l *Login) ModifyDT() time.Time {
	return l.modifyDT
}

func (l *Login) SetModifyDT(modifyDT time.Time) {
	l.modifyDT = modifyDT
}

func (l *Login) DeletedDT() *time.Time {
	return l.deletedDT
}
