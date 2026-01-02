package domain

import (
	"time"

	"github.com/google/uuid"
	hasher "math-ai.com/math-ai/internal/shared/utils/hash"
)

type Login struct {
	id          string
	uid         string
	hasspass    string
	note        *string
	loginStatus string
	status      string
	createID    *int64
	createDT    time.Time
	modifyID    *int64
	modifyDT    time.Time
	deletedDT   *time.Time
}

func NewLoginDomain() *Login {
	return &Login{}
}

func (l *Login) ID() string {
	return l.id
}

func (l *Login) SetID(id string) {
	l.id = id
}

func (l *Login) GenerateID() {
	l.id = uuid.New().String()
}

func (l *Login) UID() string {
	return l.uid
}

func (l *Login) SetUID(uid string) {
	l.uid = uid
}

func (l *Login) HassPass() string {
	return l.hasspass
}

func (l *Login) SetHassPass(password string) {
	hash, err := hasher.DefaultHasher.Hash(password)
	if err != nil {
		////logger.Errorf("failed to hash password: %v", err)
	}

	l.hasspass = string(hash)
}

func (l *Login) Note() *string {
	return l.note
}

func (l *Login) SetNote(note *string) {
	l.note = note
}

func (l *Login) LoginStatus() string {
	return l.loginStatus
}

func (l *Login) SetLoginStatus(loginStatus string) {
	l.loginStatus = loginStatus
}

func (l *Login) Status() string {
	return l.status
}

func (l *Login) SetStatus(status string) {
	l.status = status
}

func (l *Login) CreatedBy() *int64 {
	return l.createID
}

func (l *Login) SetCreatedBy(createID *int64) {
	l.createID = createID
}

func (l *Login) CreateDT() time.Time {
	return l.createDT
}

func (l *Login) SetCreateDT(createDT time.Time) {
	l.createDT = createDT
}

func (l *Login) ModifiedBy() *int64 {
	return l.modifyID
}

func (l *Login) SetModifiedBy(modifyID *int64) {
	l.modifyID = modifyID
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
