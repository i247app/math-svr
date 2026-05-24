package models

import (
	"time"
)

type AliasModel struct {
	Id          int64
	AliasId     string
	UserId      string
	Aka         string
	AliasStatus *string
	Note        *string
	CreateId    *string
	CreateDt    time.Time
	ModifyId    *string
	ModifyDt    time.Time
}
