package models

import (
	"time"
)

type UserExamDetailModel struct {
	Id                 int64
	UserExamDetailId   int64
	UserAiExamId       int64
	UserExamId         int64
	AiExamId           int64
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
	DetailStatus       *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
