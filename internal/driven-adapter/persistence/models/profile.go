package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type ProfileModel struct {
	ID        string
	UID       string
	Name      string
	Phone     string
	Email     string
	AvatarKey *string
	Dob       *time.MathTime
	Grade     string
	Semester  string
	Status    string
	CreateID  *int64
	CreateDT  time.MathTime
	ModifyID  *int64
	ModifyDT  time.MathTime
	DeletedDT *time.MathTime
}
