package models

import (
	"time"
)

type ExamPoolModel struct {
	ExamId        int64
	ReqExamType   string
	ReqGrade      int
	ReqLevel      *int
	ReqNumQues    int
	ReqSemester   *string
	ReqProgram    *string
	ReqExtras     *string
	ExamSource    string
	Title         *string
	ShortText     *string
	QuestionsJson string
	VerifiedCount int
	RptFlg        *string
	Kwords        *string
	Note          *string
	ExamStatus    *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
