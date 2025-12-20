package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type UserModel struct {
	ID           string
	Name         string
	Phone        string
	Email        string
	AvatarKey    *string
	Dob          *time.MathTime
	RoleID       string
	Role         string
	HashPassword string
	Status       string
	CreateID     *int64
	CreateDT     time.MathTime
	ModifyID     *int64
	ModifyDT     time.MathTime
}

type AliasUserModel struct {
	ID       string
	UID      string
	Aka      string
	Status   string
	CreateID *int64
	CreateDT time.MathTime
	ModifyID *int64
	ModifyDT time.MathTime
}
