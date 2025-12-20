package models

import (
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type RolePermissionModel struct {
	ID           string
	RoleID       string
	PermissionID string
	CreateID     *string
	CreateDT     time.MathTime
	ModifyID     *string
	ModifyDT     time.MathTime
	DeletedDT    *time.MathTime
}
