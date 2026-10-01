package models

import (
	"time"
)

type ExamSessionModel struct {
	EsessId            int64
	Uid                int64
	ProfileId          int64
	ReqExamType        string
	ResTotalQuestions  int
	ResCorrectNumber   int
	ResSkippedNumber   int
	ResScorePercentage *int
	ResReview          *string
	EsessFlag          *bool // TINYINT(1): NULL = no verdict
	AiShortText        *string
	AiReviewShort      *string
	AiReviewLong       *string
	CurrentGrade       *int
	CurrentLevel       *int
	LastSubmittedDt    *time.Time
	EndedDt            *time.Time
	RptFlg             *string
	Kwords             *string
	Note               *string
	EsessStatus        *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
