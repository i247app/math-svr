package models

import (
	"time"

	"github.com/google/uuid"
)

type AliasModel struct {
	Id          int64
	AliasId     uuid.UUID
	UserId      uuid.UUID
	Aka         string
	AliasStatus *string
	Note        *string
	CreateId    *uuid.UUID
	CreateDt    time.Time
	ModifyId    *uuid.UUID
	ModifyDt    time.Time
}
