package models

import (
	"time"
)

type QuestionModel struct {
	QuestionId       int64
	ExamId           int64
	QuestionNumber   int
	QuestionKind     *string
	QuestionText     string
	Options          string // JSON [{"id","text"}]
	CorrectAnswer    string // JSON {"option_ids"} / {"value"} / {"values"}
	Explanation      *string
	TtsText          *string
	QuestionLanguage string
	Translations     *string // JSON {"<lang>": {...}}; NULL = none
	QuestionTags     string  // JSON ["addition", …], never NULL
	QuestionTopic    *string
	QuestionGrade    *int
	QuestionLevel    *int
	QuestionSubLevel *int
	QuestionStatus   *string
	RptFlg           *string
	Kwords           *string
	Note             *string
	Status           string
	CreateId         *int64
	CreateDt         time.Time
	ModifyId         *int64
	ModifyDt         time.Time
}
