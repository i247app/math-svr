package models

import (
	"time"
)

type LoginModel struct {
	LoginId      int64
	UserId       int64
	Upass        *string
	LoginsStatus *string
	RptFlg       *string
	Kwords       *string
	Note         *string
	Status       string
	CreateId     *int64
	CreateDt     time.Time
	ModifyId     *int64
	ModifyDt     time.Time
}
