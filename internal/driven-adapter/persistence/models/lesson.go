package models

type LessonModel struct {
	ID           string
	ChapterID    string
	LessonNumber int
	DurationMin  *int
}
