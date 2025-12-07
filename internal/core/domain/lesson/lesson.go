package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type Lesson struct {
	id           string
	chapterID    string
	lessonNumber int
	durationMin  *int
	title        string // From translation
	content      string // From translation
}

func NewLesson() *Lesson {
	return &Lesson{}
}

func (l *Lesson) ID() string {
	return l.id
}

func (l *Lesson) GenerateID() {
	l.id = uuid.New().String()
}

func (l *Lesson) SetID(id string) {
	l.id = id
}

func (l *Lesson) ChapterID() string {
	return l.chapterID
}

func (l *Lesson) SetChapterID(chapterID string) {
	l.chapterID = chapterID
}

func (l *Lesson) LessonNumber() int {
	return l.lessonNumber
}

func (l *Lesson) SetLessonNumber(lessonNumber int) {
	l.lessonNumber = lessonNumber
}

func (l *Lesson) DurationMin() *int {
	return l.durationMin
}

func (l *Lesson) SetDurationMin(durationMin *int) {
	l.durationMin = durationMin
}

func (l *Lesson) Title() string {
	return l.title
}

func (l *Lesson) SetTitle(title string) {
	l.title = title
}

func (l *Lesson) Content() string {
	return l.content
}

func (l *Lesson) SetContent(content string) {
	l.content = content
}

func BuildLessonFromModel(model *models.LessonModel) *Lesson {
	return &Lesson{
		id:           model.ID,
		chapterID:    model.ChapterID,
		lessonNumber: model.LessonNumber,
		durationMin:  model.DurationMin,
	}
}
