package models

import (
	"time"
)

type QuizModel struct {
	Id              int64
	QuizId          int64
	UserId          *int64
	ProfileId       *int64
	Purpose         string
	TypeOfQuiz      *string
	Title           *string
	ShortText       *string
	Questions       *string
	Answers         *string
	AIReview        *string
	AIDetectGrade   *string
	TotalQuestions  *int
	CorrectNumber   *int
	ScorePercentage *int
	PreviousQuizId  *int64
	Note            *string
	QuizStatus      *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
