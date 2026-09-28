package models

import (
	"time"
)

type LoginModel struct {
	LoginId      int64
	Uid          int64
	Upw          *string
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
