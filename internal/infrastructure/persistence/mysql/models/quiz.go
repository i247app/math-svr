package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizModel struct {
	Id              int64
	QuizId          uuid.UUID
	UserId          uuid.UUID
	ProfileId       uuid.UUID
	QuizType        string
	Questions       *string
	Answers         *string
	AIReview        *string
	AIDetectGrade   *string
	TotalQuestions  *int
	CorrectNumber   *int
	ScorePercentage *int
	PreviousQuizId  *uuid.UUID
	Note            *string
	QuizStatus      *string
	Status          string
	CreateId        *uuid.UUID
	CreateDt        time.Time
	ModifyId        *uuid.UUID
	ModifyDt        time.Time
}
