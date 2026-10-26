package models

import (
	"time"
)

type QuestionModel struct {
	QuestionId       int64
	QuestionKind     string
	QuestionText     string
	QuestionImageKey *string
	Options          *string // JSON [{"id","text","image_key"?}]; NULL for kinds with no options
	CorrectAnswer    string  // JSON {"option_ids":[...]} or {"value":bool}
	Explanation      *string
	QuestionGrade    int
	QuestionTopic    *string
	QuestionLevel    *int
	QuestionTags     string // JSON array of strings
	QuestionSource   string
	VerifiedCount    int
	QuestionHash     string
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
