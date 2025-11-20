package user

import (
	"time"

	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type User struct {
	ID        string
	Name      string
	Phone     string
	Email     string
	Role      enum.ERole
	Password  string
	Status    string
	CreateID  *int64
	CreateDT  time.Time
	ModifyID  *int64
	ModifyDT  time.Time
	DeletedDT *time.Time
}
