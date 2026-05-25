package models

import (
	"time"
)

type UserModel struct {
	Id         int64
	UserId     string
	UserName   string
	Phone      string
	Email      *string
	AvatarKey  *string
	UserStatus *string
	Status     string
	Note       *string
	CreateId   *string
	CreateDt   time.Time
	ModifyId   *string
	ModifyDt   time.Time
}
