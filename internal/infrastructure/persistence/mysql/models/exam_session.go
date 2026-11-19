package models

import (
	"time"
)

type ExamSessionModel struct {
	Audit

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
	LastTitle          *string
	LastShortText      *string
	AiReviewShort      *string
	AiReviewLong       *string
	LastRespId         *string
	CurrentGrade       *int
	CurrentLevel       *int
	LastSubmittedDt    *time.Time
	EndedDt            *time.Time
	EsessStatus        *string
}
