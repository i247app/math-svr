package models

import (
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type RoleModel struct {
	ID           string
	Name         string
	Code         string
	Description  *string
	ParentRoleID *string
	IsSystemRole bool
	Status       string
	DisplayOrder int8
	CreateID     *string
	CreateDT     time.MathTime
	ModifyID     *string
	ModifyDT     time.MathTime
	DeletedDT    *time.MathTime
}
