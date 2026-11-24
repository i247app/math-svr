package models

import "time"

type GameSessionModel struct {
	Audit

	GsessId            int64
	UID                int64
	ProfileId          int64
	ReqGrade           int
	ReqLevel           *int
	ReqNumQues         int
	QuestionsJson      string
	AnswersJson        *string
	ResTotalQuestions  *int
	ResCorrectNumber   *int
	ResSkippedNumber   *int
	ResScorePercentage *int
	StartedDt          time.Time
	SubmittedDt        *time.Time
	GsessStatus        *string
}
