package models

import "time"

type RoleModel struct {
	RoleId       int64
	RoleCode     string
	RoleName     string
	Description  *string
	RoleImageKey *string
	RoleStatus   *string
	RptFlg       *string
	Kwords       *string
	Note         *string
	Status       string
	CreateId     *int64
	CreateDt     time.Time
	ModifyId     *int64
	ModifyDt     time.Time
}
