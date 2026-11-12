package semester

import (
	"time"
)

type Semester struct {
	semesterId     int64
	name           string
	description    string
	imageKey       *string
	displayOrder   int8
	rptFlg         *string
	kwords         *string
	note           *string
	semesterStatus *string
	status         string
	createId       *int64
	createDt       time.Time
	modifyId       *int64
	modifyDt       time.Time
}

func NewSemester() *Semester {
	return &Semester{}
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

// RptFlg is the client-side report flag (rpt_flg). Nil means "not reported".
func (s *Semester) RptFlg() *string {
	return s.rptFlg
}

func (s *Semester) SetRptFlg(rptFlg *string) {
	s.rptFlg = rptFlg
}

// Kwords holds search keywords (kwords) for a future text-search index.
func (s *Semester) Kwords() *string {
	return s.kwords
}

func (s *Semester) SetKwords(kwords *string) {
	s.kwords = kwords
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

func (s *Semester) CreateDt() time.Time {
	return s.createDt
}

func (s *Semester) SetCreateDt(createDt time.Time) {
	s.createDt = createDt
}

func (s *Semester) ModifyId() *int64 {
	return s.modifyId
}

func (s *Semester) SetModifyId(modifyId *int64) {
	s.modifyId = modifyId
}

func (s *Semester) ModifyDt() time.Time {
	return s.modifyDt
}

func (s *Semester) SetModifyDt(modifyDt time.Time) {
	s.modifyDt = modifyDt
}
