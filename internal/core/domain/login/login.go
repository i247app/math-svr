package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math-ai.com/math-ai/internal/shared/logger"
)

type Login struct {
	id        string
	uid       string
	hasspass  string
	status    string
	createDT  time.Time
	modifyDT  time.Time
	deletedDT *time.Time
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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("failed to hash password: %v", err)
	}

	l.hasspass = string(hash)
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
