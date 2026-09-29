package models

import (
	"time"
)

type ExamLinkModel struct {
	ElinkId            int64
	Uid                int64
	ProfileId          int64
	ExamId             int64
	EsessId            *int64
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
	RptFlg             *string
	Kwords             *string
	Note               *string
	ElinkStatus        *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
