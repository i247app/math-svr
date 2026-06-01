package models

import (
	"time"
)

type ClassroomExerciseSubmissionModel struct {
	Id                            int64
	ClassroomExerciseSubmissionId int64
	ClassroomExerciseId           int64
	ClassroomId                   int64
	ProfileId                     int64
	Answers                       *string
	AIReview                      *string
	TotalQuestions                *int64
	CorrectNumber                 *int64
	ScorePercentage               *int64
	SubmittedDt                   *time.Time
	GradedDt                      *time.Time
	Note                          *string
	SubmissionStatus              *string
	Status                        string
	CreateId                      *int64
	CreateDt                      time.Time
	ModifyId                      *int64
	ModifyDt                      time.Time
}
