package models

import (
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type PermissionModel struct {
	ID           string
	Name         string
	Description  *string
	HTTPMethod   string
	EndpointPath string
	Resource     *string
	Action       *string
	Status       string
	CreateID     *string
	CreateDT     time.MathTime
	ModifyID     *string
	ModifyDT     time.MathTime
	DeletedDT    *time.MathTime
}
