package models

import (
	"time"
)

type QuestionModel struct {
	QuestionId     int64
	ExamId         int64
	QuestionNumber int
	RenderType     string
	QuestionKind   *string
	QuestionText   string
	Options        string // JSON [{"id","text"}]
	CorrectAnswer  string // JSON {"option_ids"} / {"value"} / {"values"}
	QuestionTopic  *string
	QuestionGrade  *int
	QuestionStatus *string
	RptFlg         *string
	Kwords         *string
	Note           *string
	Status         string
	CreateId       *int64
	CreateDt       time.Time
	ModifyId       *int64
	ModifyDt       time.Time
}
