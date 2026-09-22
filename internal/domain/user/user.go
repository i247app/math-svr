package user

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

type User struct {
	id              int64
	userId          int64
	userName        string
	phone           *string
	email           *string
	isEmailVerified bool
	avatarKey       *string
	role            *string
	identityCode    *string
	userStatus      *string
	status          string
	rptFlg          *string
	kwords          *string
	note            *string
	createId        *int64
	createDt        mtime.MathTime
	modifyId        *int64
	modifyDt        mtime.MathTime
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

// Phone is the primary login key. Nullable since migration 031: a guest
// has not declared one, and is found through their ma_aliases row on
// device_uuid instead. Every registering path still requires it.
func (u *User) Phone() *string {
	return u.phone
}

func (u *User) SetPhone(phone *string) {
	u.phone = phone
}

func (u *User) Email() *string {
	return u.email
}

func (u *User) SetEmail(email *string) {
	u.email = email
}

func (u *User) IsEmailVerified() bool {
	return u.isEmailVerified
}

func (u *User) SetIsEmailVerified(isEmailVerified bool) {
	u.isEmailVerified = isEmailVerified
}

func (u *User) AvatarKey() *string {
	return u.avatarKey
}

func (u *User) SetAvatarKey(avatarKey *string) {
	u.avatarKey = avatarKey
}

// Role is the account-level role on ma_users (STUDENT / TEACHER / PARENT),
// mirroring ma_profiles.role. Nullable since migration 031: a guest has
// not declared one yet. The create command defaults it to STUDENT when a
// registering caller omits it.
func (u *User) Role() *string {
	return u.role
}

func (u *User) SetRole(role *string) {
	u.role = role
}

// IdentityCode is GUEST / USER / VERIFIED — see enum.IdentityCodeType.
// Nil only on a row written before migration 031 backfilled it.
func (u *User) IdentityCode() *string {
	return u.identityCode
}

func (u *User) SetIdentityCode(identityCode *string) {
	u.identityCode = identityCode
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

// RptFlg is the client-side report flag (rpt_flg). Nil means "not reported".
func (u *User) RptFlg() *string {
	return u.rptFlg
}

func (u *User) SetRptFlg(rptFlg *string) {
	u.rptFlg = rptFlg
}

// Kwords holds search keywords (kwords) for a future text-search index.
func (u *User) Kwords() *string {
	return u.kwords
}

func (u *User) SetKwords(kwords *string) {
	u.kwords = kwords
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
