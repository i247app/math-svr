package models

import "time"

type ChapterModel struct {
	Id            int64
	ChapterId     int64
	ProgramId     int64
	GradeId       int64
	SemesterId    int64
	Label         string
	Description   string
	DisplayOrder  int8
	Note          *string
	ChapterStatus *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}

type ChapterTranslationModel struct {
	Id                   int64
	ChapterTranslationId int64
	ChapterId            int64
	Language             string
	Label                string
	Description          string
	Note                 *string
	CtStatus             *string
	Status               string
	CreateId             *int64
	CreateDt             time.Time
	ModifyId             *int64
	ModifyDt             time.Time
}
