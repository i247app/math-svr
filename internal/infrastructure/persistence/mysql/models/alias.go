package models

import (
	"time"
)

type AliasModel struct {
	Aid         int64
	Uid         int64
	Aka         string
	AliasStatus *string
	RptFlg      *string
	Kwords      *string
	Note        *string
	CreateId    *int64
	CreateDt    time.Time
	ModifyId    *int64
	ModifyDt    time.Time
}
