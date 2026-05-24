package models

import (
	"time"
)

type QuizModel struct {
	Id              int64
	QuizId          string
	UserId          *string
	ProfileId       *string
	QuizType        string
	Questions       *string
	Answers         *string
	AIReview        *string
	AIDetectGrade   *string
	TotalQuestions  *int
	CorrectNumber   *int
	ScorePercentage *int
	PreviousQuizId  *string
	Note            *string
	QuizStatus      *string
	Status          string
	CreateId        *string
	CreateDt        time.Time
	ModifyId        *string
	ModifyDt        time.Time
}
