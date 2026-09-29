package models

import (
	"time"
)

type ExamSessionLineModel struct {
	EsessLnId          int64
	ElinkId            int64
	EsessId            int64
	ExamId             int64
	ReqExamType        string
	QuestionNumber     int
	QuestionType       *string
	QuestionName       *string
	QuestionTopic      *string
	QuestionGrade      *int
	QuestionLevel      *int
	RightAnswerLabel   *string
	RightAnswerContent *string
	SelectedLabel      string
	SelectedContent    *string
	IsCorrect          bool
	RptFlg             *string
	Kwords             *string
	Note               *string
	EsessLnStatus      *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
