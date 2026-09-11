package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// UserAiExam is ONE ATTEMPT: a (user, profile) pair sitting down with one
// AiExam. It is created the moment the exam is handed out, which is what
// makes "you left this one unfinished" answerable, and it is the only
// place a single sitting's result is kept — ma_user_exams holds lifetime
// totals and has no time dimension at all.
//
// reqExamType / reqGrade / reqLevel are snapshots of where the child was
// placed when they took it. Reading them off the AiExam row would work
// today, but that row is shared cache; the snapshot keeps one child's
// history readable on its own terms.
//
// resTotalQuestions counts the questions the child ANSWERED, not the
// questions the exam contained. A skipped question lands in
// resSkippedNumber instead, so "answered 3 of 10 and got them all right"
// is distinguishable from "answered all 10 correctly".
type UserAiExam struct {
	id           int64
	userAiExamId int64
	userId       int64
	profileId    int64
	aiExamId     int64

	reqExamType string
	reqGrade    int
	reqLevel    *int

	resTotalQuestions  *int
	resCorrectNumber   *int
	resSkippedNumber   *int
	resScorePercentage *int

	startedDt   mtime.MathTime
	submittedDt mtime.MathTime

	note             *string
	userAiExamStatus *string
	status           string
	createId         *int64
	createDt         mtime.MathTime
	modifyId         *int64
	modifyDt         mtime.MathTime
}

func NewUserAiExam() *UserAiExam { return &UserAiExam{} }

func (u *UserAiExam) Id() int64                       { return u.id }
func (u *UserAiExam) SetId(id int64)                  { u.id = id }
func (u *UserAiExam) UserAiExamId() int64             { return u.userAiExamId }
func (u *UserAiExam) SetUserAiExamId(id int64)        { u.userAiExamId = id }
func (u *UserAiExam) UserId() int64                   { return u.userId }
func (u *UserAiExam) SetUserId(id int64)              { u.userId = id }
func (u *UserAiExam) ProfileId() int64                { return u.profileId }
func (u *UserAiExam) SetProfileId(id int64)           { u.profileId = id }
func (u *UserAiExam) AiExamId() int64                 { return u.aiExamId }
func (u *UserAiExam) SetAiExamId(id int64)            { u.aiExamId = id }
func (u *UserAiExam) ReqExamType() string             { return u.reqExamType }
func (u *UserAiExam) SetReqExamType(t string)         { u.reqExamType = t }
func (u *UserAiExam) ReqGrade() int                   { return u.reqGrade }
func (u *UserAiExam) SetReqGrade(g int)               { u.reqGrade = g }
func (u *UserAiExam) ReqLevel() *int                  { return u.reqLevel }
func (u *UserAiExam) SetReqLevel(l *int)              { u.reqLevel = l }
func (u *UserAiExam) ResTotalQuestions() *int         { return u.resTotalQuestions }
func (u *UserAiExam) SetResTotalQuestions(n *int)     { u.resTotalQuestions = n }
func (u *UserAiExam) ResCorrectNumber() *int          { return u.resCorrectNumber }
func (u *UserAiExam) SetResCorrectNumber(n *int)      { u.resCorrectNumber = n }
func (u *UserAiExam) ResSkippedNumber() *int          { return u.resSkippedNumber }
func (u *UserAiExam) SetResSkippedNumber(n *int)      { u.resSkippedNumber = n }
func (u *UserAiExam) ResScorePercentage() *int        { return u.resScorePercentage }
func (u *UserAiExam) SetResScorePercentage(n *int)    { u.resScorePercentage = n }
func (u *UserAiExam) StartedDt() mtime.MathTime       { return u.startedDt }
func (u *UserAiExam) SetStartedDt(t mtime.MathTime)   { u.startedDt = t }
func (u *UserAiExam) SubmittedDt() mtime.MathTime     { return u.submittedDt }
func (u *UserAiExam) SetSubmittedDt(t mtime.MathTime) { u.submittedDt = t }
func (u *UserAiExam) Note() *string                   { return u.note }
func (u *UserAiExam) SetNote(s *string)               { u.note = s }
func (u *UserAiExam) UserAiExamStatus() *string       { return u.userAiExamStatus }
func (u *UserAiExam) SetUserAiExamStatus(s *string)   { u.userAiExamStatus = s }
func (u *UserAiExam) Status() string                  { return u.status }
func (u *UserAiExam) SetStatus(s string)              { u.status = s }
func (u *UserAiExam) CreateId() *int64                { return u.createId }
func (u *UserAiExam) SetCreateId(id *int64)           { u.createId = id }
func (u *UserAiExam) CreateDt() mtime.MathTime        { return u.createDt }
func (u *UserAiExam) SetCreateDt(t mtime.MathTime)    { u.createDt = t }
func (u *UserAiExam) ModifyId() *int64                { return u.modifyId }
func (u *UserAiExam) SetModifyId(id *int64)           { u.modifyId = id }
func (u *UserAiExam) ModifyDt() mtime.MathTime        { return u.modifyDt }
func (u *UserAiExam) SetModifyDt(t mtime.MathTime)    { u.modifyDt = t }
