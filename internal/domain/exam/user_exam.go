package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// UserExam is one JOURNEY: a stretch of one exam type that a (user,
// profile) pair works through and then closes. Every submission of that
// type folds into the open journey; once it is marked COMPLETE or CANCEL,
// the next submission opens a new one. A user account holds several
// profiles (one per child), so the pair, not the user alone, is the
// subject of the statistics.
//
// At most one journey per triple is ACTIVE at a time — the database holds
// that line (uk_active_journey), not this type.
//
// resTotalQuestions accumulates ANSWERED questions, never the size of the
// exams. Three ten-question rounds with six answers each give 18, not 30,
// and resCorrectNumber can therefore never exceed it.
//
// currentGrade and currentLevel are where the child is WORKING right
// now, as the client states it on each hand-out (grade / level on
// /exams/generate). The server records them; it no longer derives them
// from results — the teaching side owns that judgement. They are not
// written back to the profile: ma_profiles.grade_id records the class
// the child attends, which is a different fact.
type UserExam struct {
	userExamId   int64
	userId       int64
	profileId    int64
	reqExamType  string
	currentGrade *int
	currentLevel *int

	resTotalQuestions  int
	resCorrectNumber   int
	resSkippedNumber   int
	resScorePercentage *int
	resReview          *string

	lastSubmittedDt mtime.MathTime
	endedDt         mtime.MathTime

	rptFlg *string

	kwords *string

	note           *string
	userExamStatus *string
	status         string
	createId       *int64
	createDt       mtime.MathTime
	modifyId       *int64
	modifyDt       mtime.MathTime
}

func NewUserExam() *UserExam { return &UserExam{} }

func (u *UserExam) UserExamId() int64                   { return u.userExamId }
func (u *UserExam) SetUserExamId(id int64)              { u.userExamId = id }
func (u *UserExam) UserId() int64                       { return u.userId }
func (u *UserExam) SetUserId(id int64)                  { u.userId = id }
func (u *UserExam) ProfileId() int64                    { return u.profileId }
func (u *UserExam) SetProfileId(id int64)               { u.profileId = id }
func (u *UserExam) ReqExamType() string                 { return u.reqExamType }
func (u *UserExam) SetReqExamType(t string)             { u.reqExamType = t }
func (u *UserExam) ResTotalQuestions() int              { return u.resTotalQuestions }
func (u *UserExam) SetResTotalQuestions(n int)          { u.resTotalQuestions = n }
func (u *UserExam) ResCorrectNumber() int               { return u.resCorrectNumber }
func (u *UserExam) SetResCorrectNumber(n int)           { u.resCorrectNumber = n }
func (u *UserExam) ResSkippedNumber() int               { return u.resSkippedNumber }
func (u *UserExam) SetResSkippedNumber(n int)           { u.resSkippedNumber = n }
func (u *UserExam) ResScorePercentage() *int            { return u.resScorePercentage }
func (u *UserExam) SetResScorePercentage(n *int)        { u.resScorePercentage = n }
func (u *UserExam) ResReview() *string                  { return u.resReview }
func (u *UserExam) SetResReview(s *string)              { u.resReview = s }
func (u *UserExam) CurrentGrade() *int                  { return u.currentGrade }
func (u *UserExam) SetCurrentGrade(g *int)              { u.currentGrade = g }
func (u *UserExam) CurrentLevel() *int                  { return u.currentLevel }
func (u *UserExam) SetCurrentLevel(l *int)              { u.currentLevel = l }
func (u *UserExam) LastSubmittedDt() mtime.MathTime     { return u.lastSubmittedDt }
func (u *UserExam) SetLastSubmittedDt(t mtime.MathTime) { u.lastSubmittedDt = t }
func (u *UserExam) EndedDt() mtime.MathTime             { return u.endedDt }
func (u *UserExam) SetEndedDt(t mtime.MathTime)         { u.endedDt = t }
func (u *UserExam) Note() *string                       { return u.note }
func (u *UserExam) SetNote(s *string)                   { u.note = s }
func (u *UserExam) RptFlg() *string                     { return u.rptFlg }
func (u *UserExam) SetRptFlg(v *string)                 { u.rptFlg = v }
func (u *UserExam) Kwords() *string                     { return u.kwords }
func (u *UserExam) SetKwords(v *string)                 { u.kwords = v }
func (u *UserExam) UserExamStatus() *string             { return u.userExamStatus }
func (u *UserExam) SetUserExamStatus(s *string)         { u.userExamStatus = s }
func (u *UserExam) Status() string                      { return u.status }
func (u *UserExam) SetStatus(s string)                  { u.status = s }
func (u *UserExam) CreateId() *int64                    { return u.createId }
func (u *UserExam) SetCreateId(id *int64)               { u.createId = id }
func (u *UserExam) CreateDt() mtime.MathTime            { return u.createDt }
func (u *UserExam) SetCreateDt(t mtime.MathTime)        { u.createDt = t }
func (u *UserExam) ModifyId() *int64                    { return u.modifyId }
func (u *UserExam) SetModifyId(id *int64)               { u.modifyId = id }
func (u *UserExam) ModifyDt() mtime.MathTime            { return u.modifyDt }
func (u *UserExam) SetModifyDt(t mtime.MathTime)        { u.modifyDt = t }
