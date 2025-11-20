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
	Status       string
	HashPassword string
	CreateID     *int64
	CreateDT     time.Time
	ModifyID     *int64
	ModifyDT     time.Time
}
