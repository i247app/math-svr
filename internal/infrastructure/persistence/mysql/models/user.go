package models

import (
	"time"
)

type UserModel struct {
	Id         int64
	UserId     int64
	UserName   string
	Phone      string
	Email      *string
	AvatarKey  *string
	Role       string
	UserStatus *string
	Status     string
	Note       *string
	CreateId   *int64
	CreateDt   time.Time
	ModifyId   *int64
	ModifyDt   time.Time
}
