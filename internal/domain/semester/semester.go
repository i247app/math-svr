package semester

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

type Semester struct {
	id             int64
	semesterId     int64
	name           string
	description    string
	imageKey       *string
	displayOrder   int8
	note           *string
	semesterStatus *string
	status         string
	createId       *int64
	createDt       mtime.MathTime
	modifyId       *int64
	modifyDt       mtime.MathTime
}

func NewSemester() *Semester {
	return &Semester{}
}

func (s *Semester) Id() int64 {
	return s.id
}

func (s *Semester) SetId(id int64) {
	s.id = id
}

func (s *Semester) SemesterId() int64 {
	return s.semesterId
}

func (s *Semester) SetSemesterId(semesterId int64) {
	s.semesterId = semesterId
}

func (s *Semester) Name() string {
	return s.name
}

func (s *Semester) SetName(name string) {
	s.name = name
}

func (s *Semester) Description() string {
	return s.description
}

func (s *Semester) SetDescription(description string) {
	s.description = description
}

func (s *Semester) ImageKey() *string {
	return s.imageKey
}

func (s *Semester) SetImageKey(imageKey *string) {
	s.imageKey = imageKey
}

func (s *Semester) DisplayOrder() int8 {
	return s.displayOrder
}

func (s *Semester) SetDisplayOrder(displayOrder int8) {
	s.displayOrder = displayOrder
}

func (s *Semester) Note() *string {
	return s.note
}

func (s *Semester) SetNote(note *string) {
	s.note = note
}

func (s *Semester) SemesterStatus() *string {
	return s.semesterStatus
}

func (s *Semester) SetSemesterStatus(semesterStatus *string) {
	s.semesterStatus = semesterStatus
}

func (s *Semester) Status() string {
	return s.status
}

func (s *Semester) SetStatus(status string) {
	s.status = status
}

func (s *Semester) CreateId() *int64 {
	return s.createId
}

func (s *Semester) SetCreateId(createId *int64) {
	s.createId = createId
}

func (s *Semester) CreateDt() mtime.MathTime {
	return s.createDt
}

func (s *Semester) SetCreateDt(createDt mtime.MathTime) {
	s.createDt = createDt
}

func (s *Semester) ModifyId() *int64 {
	return s.modifyId
}

func (s *Semester) SetModifyId(modifyId *int64) {
	s.modifyId = modifyId
}

func (s *Semester) ModifyDt() mtime.MathTime {
	return s.modifyDt
}

func (s *Semester) SetModifyDt(modifyDt mtime.MathTime) {
	s.modifyDt = modifyDt
}
