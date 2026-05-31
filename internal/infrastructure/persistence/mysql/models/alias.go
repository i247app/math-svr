package models

import (
	"time"
)

type AliasModel struct {
	Id          int64
	AliasId     int64
	UserId      int64
	Aka         string
	AliasStatus *string
	Note        *string
	CreateId    *int64
	CreateDt    time.Time
	ModifyId    *int64
	ModifyDt    time.Time
}
