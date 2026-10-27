package models

import (
	"time"
)

type QuestionModel struct {
	QuestionId         int64
	ExamId             int64
	QuestionNumber     int
	QuestionType       string
	QuestionName       string
	Answers            string // JSON [{"label","content"}]
	RightAnswerLabel   string
	RightAnswerContent *string
	QuestionTopic      *string
	QuestionGrade      *int
	QuestionStatus     *string
	RptFlg             *string
	Kwords             *string
	Note               *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
