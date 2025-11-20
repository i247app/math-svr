package user

import (
	"time"

	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type User struct {
	id        int64
	name      string
	phone     string
	email     string
	avatarUrl *string
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

func (u *User) ID() int64 {
	return u.id
}

func (u *User) SetID(id int64) {
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

func (u *User) AvatarURL() *string {
	return u.avatarUrl
}

func (u *User) SetAvatarURL(avatarURL *string) {
	u.avatarUrl = avatarURL
}

func (u *User) Role() string {
	return u.role
}

func (u *User) SetRole(role string) {
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

func UserDomainFromModel(model *models.UserModel) *User {
	return &User{
		id:        model.ID,
		name:      model.Name,
		phone:     model.Phone,
		email:     model.Email,
		avatarUrl: model.AvatarUrl,
		role:      model.Role,
		password:  model.HashPassword,
		status:    model.Status,
		createDT:  model.CreateDT,
		modifyDT:  model.ModifyDT,
	}
}
