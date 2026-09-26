package models

import (
	"time"
)

type UserModel struct {
	UserId          int64
	UserName        string
	Phone           *string
	Email           *string
	IsEmailVerified bool
	AvatarKey       *string
	Role            *string
	IdentityCode    *string
	UserStatus      *string
	Status          string
	RptFlg          *string
	Kwords          *string
	Note            *string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
