package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type ContactModel struct {
	ID             string
	UID            *string
	ContactName    string
	ContactEmail   *string
	ContactPhone   *string
	ContactMessage string
	IsRead         *bool
	Note           *string
	ContactStatus  string
	Status         string
	CreateID       *int64
	CreateDT       time.MathTime
	ModifyID       *int64
	ModifyDT       time.MathTime
	DeletedDT      *time.MathTime
}
