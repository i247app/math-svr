package models

type UserModel struct {
	Audit

	Uid             int64
	UserName        string
	Phone           *string
	Email           *string
	IsEmailVerified bool
	AvatarKey       *string
	Role            *string
	IdentityCode    *string
	UserStatus      *string
}
