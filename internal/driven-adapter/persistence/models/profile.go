package models

import (
	"time"
)

type ProfileModel struct {
	ID        string
	UID       string
	Name      string
	Phone     string
	Email     string
	Grade     string
	Level     string
	Status    string
	CreateID  *int64
	CreateDT  time.Time
	ModifyID  *int64
	ModifyDT  time.Time
	DeletedDT *time.Time
}
