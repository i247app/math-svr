package models

import (
	"time"
)

type ExerciseSubmissionModel struct {
	Audit

	ClassroomExerciseSubmissionId int64
	ClassroomExerciseId           int64
	ClassroomId                   int64
	ProfileId                     int64
	Answers                       *string
	Review                        *string
	TotalQuestions                *int64
	CorrectNumber                 *int64
	ScorePercentage               *int64
	SubmittedDt                   *time.Time
	GradedDt                      *time.Time
	SubmissionStatus              *string
}
