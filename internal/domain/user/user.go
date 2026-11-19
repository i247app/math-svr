package user

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

type User struct {
	audit.Audit

	uid             int64
	userName        string
	phone           *string
	email           *string
	isEmailVerified bool
	avatarKey       *string
	role            *string
	identityCode    *string
	userStatus      *string
}

func NewUser() *User {
	return &User{}
}

func (u *User) UID() int64       { return u.uid }
func (u *User) SetUid(uid int64) { u.uid = uid }

func (u *User) UserName() string            { return u.userName }
func (u *User) SetUserName(userName string) { u.userName = userName }

// Phone is the primary login key. Nullable since migration 031: a guest
// has not declared one, and is found through their ma_aliases row on
// device_uuid instead. Every registering path still requires it.
func (u *User) Phone() *string         { return u.phone }
func (u *User) SetPhone(phone *string) { u.phone = phone }

func (u *User) Email() *string         { return u.email }
func (u *User) SetEmail(email *string) { u.email = email }

func (u *User) IsEmailVerified() bool                   { return u.isEmailVerified }
func (u *User) SetIsEmailVerified(isEmailVerified bool) { u.isEmailVerified = isEmailVerified }

func (u *User) AvatarKey() *string             { return u.avatarKey }
func (u *User) SetAvatarKey(avatarKey *string) { u.avatarKey = avatarKey }

// Role is the account-level role on ma_users (STUDENT / TEACHER / PARENT),
// mirroring ma_profiles.role. Nullable since migration 031: a guest has
// not declared one yet. The create command defaults it to STUDENT when a
// registering caller omits it.
func (u *User) Role() *string        { return u.role }
func (u *User) SetRole(role *string) { u.role = role }

// IdentityCode is GUEST / USER / VERIFIED — see enum.IdentityCodeType.
// Nil only on a row written before migration 031 backfilled it.
func (u *User) IdentityCode() *string                { return u.identityCode }
func (u *User) SetIdentityCode(identityCode *string) { u.identityCode = identityCode }

func (u *User) UserStatus() *string              { return u.userStatus }
func (u *User) SetUserStatus(userStatus *string) { u.userStatus = userStatus }
