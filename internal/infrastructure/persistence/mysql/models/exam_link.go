package models

import (
	"time"
)

type ExamLinkModel struct {
	Audit

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
	ElinkStatus        *string
}
