package user

import (
	"time"

	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type User struct {
	id        int64
	name      string
	phone     string
	email     string
	avatarUrl *string
	role      enum.ERole
	password  string
	status    string
	createDT  time.Time
	modifyDT  time.Time
	deletedDT *time.Time
}

func (u *User) ID() int64 {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Phone() string {
	return u.phone
}

func (u *User) Email() string {
	return u.email
}

func (u *User) AvatarURL() *string {
	return u.avatarUrl
}

func (u *User) Role() enum.ERole {
	return u.role
}

func (u *User) Password() string {
	return u.password
}

func (u *User) Status() string {
	return u.status
}

func (u *User) CreatedAt() time.Time {
	return u.createDT
}

func (u *User) ModifyAt() time.Time {
	return u.modifyDT
}

func (u *User) DeletedAt() *time.Time {
	return u.deletedDT
}
