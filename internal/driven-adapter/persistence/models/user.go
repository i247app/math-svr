package models

import (
	"time"

	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type UserModel struct {
	ID           int64
	Name         string
	Phone        string
	Email        string
	Role         *string
	Status       enum.EStatus
	HashPassword string
	CreateID     *int64
	CreateDT     time.Time
	ModifyID     *int64
	ModifyDT     time.Time
}
