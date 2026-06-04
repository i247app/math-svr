package models

import (
	"time"
)

type ExerciseModel struct {
	Id                  int64
	ClassroomExerciseId int64
	ClassroomId         int64
	CreatorProfileId    *int64
	Visibility          string
	ProgramId           *int64
	Title               string
	Description         *string
	ChapterName         string
	LessonName          string
	TotalQuestions      int
	Questions           *string
	Answers             *string
	StartDate           *time.Time
	EndDate             *time.Time
	Note                *string
	ExerciseStatus      *string
	Status              string
	CreateId            *int64
	CreateDt            time.Time
	ModifyId            *int64
	ModifyDt            time.Time
}
