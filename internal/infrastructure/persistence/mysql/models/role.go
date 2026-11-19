package models

type RoleModel struct {
	Audit

	RoleId       int64
	RoleCode     string
	RoleName     string
	Description  *string
	RoleImageKey *string
	RoleStatus   *string
}
