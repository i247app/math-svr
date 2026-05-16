package models

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	Id         int64
	UserId     uuid.UUID
	Phone      string
	Email      *string
	UserStatus *string
	Status     string
	Note       *string
	CreateId   *uuid.UUID
	CreateDt   time.Time
	ModifyId   *uuid.UUID
	ModifyDt   time.Time
}
