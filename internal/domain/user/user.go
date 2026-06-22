package user

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type User struct {
	id         int64
	userId     int64
	userName   string
	phone      string
	email      *string
	avatarKey  *string
	role       string
	userStatus *string
	status     string
	note       *string
	createId   *int64
	createDt   mtime.MathTime
	modifyId   *int64
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

func (u *User) UserId() int64 {
	return u.userId
}

func (u *User) SetUserId(userId int64) {
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

// Role is the account-level role on ma_users (STUDENT / TEACHER / PARENT),
// mirroring ma_profiles.role. NOT NULL at the schema layer — the create
// command defaults it to STUDENT when the caller omits it.
func (u *User) Role() string {
	return u.role
}

func (u *User) SetRole(role string) {
	u.role = role
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

func (u *User) CreateId() *int64 {
	return u.createId
}

func (u *User) SetCreateId(createId *int64) {
	u.createId = createId
}

func (u *User) CreateDt() mtime.MathTime {
	return u.createDt
}

func (u *User) SetCreateDt(createDt mtime.MathTime) {
	u.createDt = createDt
}

func (u *User) ModifyId() *int64 {
	return u.modifyId
}

func (u *User) SetModifyId(modifyId *int64) {
	u.modifyId = modifyId
}

func (u *User) ModifyDt() mtime.MathTime {
	return u.modifyDt
}

func (u *User) SetModifyDt(modifyDt mtime.MathTime) {
	u.modifyDt = modifyDt
}
