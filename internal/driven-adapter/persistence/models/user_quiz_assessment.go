package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type UserQuizAssessmentModel struct {
	ID            string
	UID           string
	Questions     string
	Answers       string
	AIReview      string
	AIDetectGrade string
	Status        string
	CreateID      *int64
	CreateDT      time.MathTime
	ModifyID      *int64
	ModifyDT      time.MathTime
	DeletedDT     *time.MathTime
}
