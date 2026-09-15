package models

import (
	"time"
)

type UserExamModel struct {
	Id                 int64
	UserExamId         int64
	UserId             int64
	ProfileId          int64
	ReqExamType        string
	ResTotalQuestions  int
	ResCorrectNumber   int
	ResSkippedNumber   int
	ResScorePercentage *int
	ResReview          *string
	CurrentGrade       *int
	CurrentLevel       *int
	LastSubmittedDt    *time.Time
	EndedDt            *time.Time
	Note               *string
	UserExamStatus     *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
