package models

import (
	"time"
)

type UserModel struct {
	ID           string
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
	ID       string
	UID      string
	Aka      string
	Status   string
	CreateID *int64
	CreateDT time.Time
	ModifyID *int64
	ModifyDT time.Time
}
