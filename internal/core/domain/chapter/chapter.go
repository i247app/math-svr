package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type Chapter struct {
	id            string
	gradeID       string
	termID        string
	chapterNumber int
	title         string // From translation
	description   string // From translation
}

func NewChapter() *Chapter {
	return &Chapter{}
}

func (c *Chapter) ID() string {
	return c.id
}

func (c *Chapter) GenerateID() {
	c.id = uuid.New().String()
}

func (c *Chapter) SetID(id string) {
	c.id = id
}

func (c *Chapter) GradeID() string {
	return c.gradeID
}

func (c *Chapter) SetGradeID(gradeID string) {
	c.gradeID = gradeID
}

func (c *Chapter) TermID() string {
	return c.termID
}

func (c *Chapter) SetTermID(termID string) {
	c.termID = termID
}

func (c *Chapter) ChapterNumber() int {
	return c.chapterNumber
}

func (c *Chapter) SetChapterNumber(chapterNumber int) {
	c.chapterNumber = chapterNumber
}

func (c *Chapter) Title() string {
	return c.title
}

func (c *Chapter) SetTitle(title string) {
	c.title = title
}

func (c *Chapter) Description() string {
	return c.description
}

func (c *Chapter) SetDescription(description string) {
	c.description = description
}

func BuildChapterFromModel(model *models.ChapterModel) *Chapter {
	return &Chapter{
		id:            model.ID,
		gradeID:       model.GradeID,
		termID:        model.TermID,
		chapterNumber: model.ChapterNumber,
	}
}
