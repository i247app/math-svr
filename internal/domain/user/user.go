package user

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type User struct {
	id         int64
	userId     string
	userName   string
	phone      string
	email      *string
	avatarKey  *string
	userStatus *string
	status     string
	note       *string
	createId   *string
	createDt   mtime.MathTime
	modifyId   *string
	modifyDt   mtime.MathTime
}

func NewUser() *User {
	return &User{}
}

func (u *User) Id() int64 {
	return u.id
}

func (u *User) SetId(id int64) {
	u.id = id
}

func (u *User) UserId() string {
	return u.userId
}

func (u *User) SetUserId(userId string) {
	u.userId = userId
}

func (u *User) UserName() string {
	return u.userName
}

func (u *User) SetUserName(userName string) {
	u.userName = userName
}

func (u *User) Phone() string {
	return u.phone
}

func (u *User) SetPhone(phone string) {
	u.phone = phone
}

func (u *User) Email() *string {
	return u.email
}

func (u *User) SetEmail(email *string) {
	u.email = email
}

func (u *User) AvatarKey() *string {
	return u.avatarKey
}

func (u *User) SetAvatarKey(avatarKey *string) {
	u.avatarKey = avatarKey
}

func (u *User) UserStatus() *string {
	return u.userStatus
}

func (u *User) SetUserStatus(userStatus *string) {
	u.userStatus = userStatus
}

func (u *User) Status() string {
	return u.status
}

func (u *User) SetStatus(status string) {
	u.status = status
}

func (u *User) Note() *string {
	return u.note
}

func (u *User) SetNote(note *string) {
	u.note = note
}

func (u *User) CreateId() *string {
	return u.createId
}

func (u *User) SetCreateId(createId *string) {
	u.createId = createId
}

func (u *User) CreateDt() mtime.MathTime {
	return u.createDt
}

func (u *User) SetCreateDt(createDt mtime.MathTime) {
	u.createDt = createDt
}

func (u *User) ModifyId() *string {
	return u.modifyId
}

func (u *User) SetModifyId(modifyId *string) {
	u.modifyId = modifyId
}

func (u *User) ModifyDt() mtime.MathTime {
	return u.modifyDt
}

func (u *User) SetModifyDt(modifyDt mtime.MathTime) {
	u.modifyDt = modifyDt
}
