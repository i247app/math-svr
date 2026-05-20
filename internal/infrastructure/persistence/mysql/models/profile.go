package models

import (
	"time"

	"github.com/google/uuid"
)

type ProfileModel struct {
	Id            int64
	ProfileId     uuid.UUID
	UserId        uuid.UUID
	Name          string
	AvatarKey     *string
	Dob           *time.Time
	ProgramId     *uuid.UUID
	GradeId       *uuid.UUID
	SemesterId    *uuid.UUID
	IsDefault     bool
	Note          *string
	ProfileStatus *string
	Status        string
	CreateId      *uuid.UUID
	CreateDt      time.Time
	ModifyId      *uuid.UUID
	ModifyDt      time.Time
}
