package models

import (
	"time"
)

type UserQuizPracticesModel struct {
	ID        string
	UID       string
	Questions string
	Answers   string
	AIReview  string
	Status    string
	CreateID  *int64
	CreateDT  time.Time
	ModifyID  *int64
	ModifyDT  time.Time
	DeletedDT *time.Time
}
