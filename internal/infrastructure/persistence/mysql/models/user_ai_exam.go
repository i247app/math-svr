package models

import (
	"time"
)

type UserAiExamModel struct {
	Id                 int64
	UserAiExamId       int64
	UserId             int64
	ProfileId          int64
	AiExamId           int64
	UserExamId         *int64
	ShuffleMap         *string
	ReqExamType        string
	ReqGrade           int
	ReqLevel           *int
	ResTotalQuestions  *int
	ResCorrectNumber   *int
	ResSkippedNumber   *int
	ResScorePercentage *int
	StartedDt          *time.Time
	SubmittedDt        *time.Time
	Note               *string
	UserAiExamStatus   *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
