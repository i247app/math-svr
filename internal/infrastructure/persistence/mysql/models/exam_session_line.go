package models

import (
	"time"
)

type ExamSessionLineModel struct {
	EsessLnId       int64
	ElinkId         int64
	EsessId         int64
	ExamId          int64
	ReqExamType     string
	QuestionNumber  int
	QuestionKind    *string
	QuestionText    *string
	QuestionTopic   *string
	QuestionTags    string // JSON ["addition", …], never NULL
	Explanation     *string
	QuestionGrade   *int
	QuestionLevel   *int
	RightOptionIDs  *string
	RightAnswerText *string
	ChildOptionIDs  string
	ChildAnswerText *string
	IsCorrect       bool
	RptFlg          *string
	Kwords          *string
	Note            *string
	EsessLnStatus   *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
