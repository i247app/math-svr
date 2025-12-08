package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type User struct {
	id        string
	name      string
	phone     string
	email     string
	avatarKey *string
	dob       *time.Time
	role      string
	password  string
	status    string
	createDT  time.Time
	modifyDT  time.Time
	deletedDT *time.Time
}

func NewUserDomain() *User {
	return &User{}
}

func (u *User) ID() string {
	return u.id
}

func (u *User) GenerateID() {
	u.id = uuid.New().String()
}

func (u *User) SetID(id string) {
	u.id = id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) SetName(name string) {
	u.name = name
}

func (u *User) Phone() string {
	return u.phone
}

func (u *User) SetPhone(phone string) {
	u.phone = phone
}

func (u *User) Email() string {
	return u.email
}

func (u *User) SetEmail(email string) {
	u.email = email
}

func (u *User) AvatarKey() *string {
	return u.avatarKey
}

func (u *User) SetAvatarKey(avatarKey *string) {
	u.avatarKey = avatarKey
}

func (u *User) DOB() *time.Time {
	return u.dob
}

func (u *User) SetDOB(dob *time.Time) {
	u.dob = dob
}

func (u *User) Role() string {
	return u.role
}

func (u *User) SetRole(role string) {
	if role == "" {
		role = string(enum.RoleUser)
	}

	u.role = role
}

func (u *User) Password() string {
	return u.password
}

func (u *User) SetPassword(password string) {
	u.password = password
}

func (u *User) Status() string {
	return u.status
}

func (u *User) SetStatus(status string) {
	u.status = status
}

func (u *User) CreatedAt() time.Time {
	return u.createDT
}

func (u *User) SetCreatedAt(createAt time.Time) {
	u.createDT = createAt
}

func (u *User) ModifyAt() time.Time {
	return u.modifyDT
}

func (u *User) SetModifyAt(modifyAt time.Time) {
	u.modifyDT = modifyAt
}

func (u *User) DeletedAt() *time.Time {
	return u.deletedDT
}

func (u *User) SetDeletedAt(deletedAt *time.Time) {
	u.deletedDT = deletedAt
}

func BuildUserDomainFromModel(model *models.UserModel) *User {
	return &User{
		id:        model.ID,
		name:      model.Name,
		phone:     model.Phone,
		email:     model.Email,
		avatarKey: model.AvatarKey,
		dob:       model.Dob,
		role:      model.Role,
		password:  model.HashPassword,
		status:    model.Status,
		createDT:  model.CreateDT,
		modifyDT:  model.ModifyDT,
	}
}
