package semester

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

type Semester struct {
	audit.Audit

	semesterId     int64
	name           string
	description    string
	imageKey       *string
	displayOrder   int8
	semesterStatus *string
}

func NewSemester() *Semester {
	return &Semester{}
}

func (s *Semester) SemesterId() int64              { return s.semesterId }
func (s *Semester) SetSemesterId(semesterId int64) { s.semesterId = semesterId }

func (s *Semester) Name() string        { return s.name }
func (s *Semester) SetName(name string) { s.name = name }

func (s *Semester) Description() string               { return s.description }
func (s *Semester) SetDescription(description string) { s.description = description }

func (s *Semester) ImageKey() *string            { return s.imageKey }
func (s *Semester) SetImageKey(imageKey *string) { s.imageKey = imageKey }

func (s *Semester) DisplayOrder() int8                { return s.displayOrder }
func (s *Semester) SetDisplayOrder(displayOrder int8) { s.displayOrder = displayOrder }

func (s *Semester) SemesterStatus() *string                  { return s.semesterStatus }
func (s *Semester) SetSemesterStatus(semesterStatus *string) { s.semesterStatus = semesterStatus }
