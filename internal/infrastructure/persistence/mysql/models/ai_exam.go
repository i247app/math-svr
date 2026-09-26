package models

import (
	"time"
)

type AiExamModel struct {
	AiExamId        int64
	ReqExamType     string
	ReqGrade        int
	ReqLevel        *int
	ReqNumQues      int
	ReqSemester     *string
	ReqProgram      *string
	ReqExtras       *string
	AiTitle         *string
	AiShortText     *string
	AiQuestionsJson string
	RptFlg          *string
	Kwords          *string
	Note            *string
	AiExamStatus    *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
