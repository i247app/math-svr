package models

import "time"

type ChapterModel struct {
	Id            int64
	ChapterId     string
	ProgramId     string
	GradeId       string
	SemesterId    string
	Label         string
	Description   string
	DisplayOrder  int8
	Note          *string
	ChapterStatus *string
	Status        string
	CreateId      *string
	CreateDt      time.Time
	ModifyId      *string
	ModifyDt      time.Time
}

type ChapterTranslationModel struct {
	Id                   int64
	ChapterTranslationId string
	ChapterId            string
	Language             string
	Label                string
	Description          string
	Note                 *string
	CtStatus             *string
	Status               string
	CreateId             *string
	CreateDt             time.Time
	ModifyId             *string
	ModifyDt             time.Time
}
