package models

import (
	"time"
)

type UserModel struct {
	ID           int64
	Name         string
	Phone        string
	Email        string
	AvatarUrl    *string
	Role         string
	HashPassword string
	Status       string
	CreateID     *int64
	CreateDT     time.Time
	ModifyID     *int64
	ModifyDT     time.Time
}

type AliasUserModel struct {
	ID       int64
	UID      int64
	Aka      string
	Status   string
	CreateID *int64
	CreateDT time.Time
	ModifyID *int64
	ModifyDT time.Time
}
