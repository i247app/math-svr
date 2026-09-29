package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// ExamLink is ONE ATTEMPT: a (user, profile) pair sitting down with one
// ExamPool. It is created the moment the exam is handed out, which is what
// makes "you left this one unfinished" answerable, and it is the only
// place a single sitting's result is kept — ma_exam_sessions holds lifetime
// totals and has no time dimension at all.
//
// reqExamType / reqGrade are snapshots of where the child was placed when
// they took it. Reading them off the ExamPool row would work today, but that
// row is shared cache; the snapshot keeps one child's history readable on
// its own terms. reqLevel mirrors the nullable req_level column and is
// always NULL — no level rule exists yet, so nothing sets it.
//
// resTotalQuestions counts the questions the child ANSWERED, not the
// questions the exam contained. A skipped question lands in
// resSkippedNumber instead, so "answered 3 of 10 and got them all right"
// is distinguishable from "answered all 10 correctly".
//
// esessId is the journey this sitting belongs to. It is known at
// hand-out time for a PRACTICE round (the client names the journey) and
// for an ASSESSMENT drawn while a journey is open; the very first
// ASSESSMENT of a journey is handed out before the journey row exists and
// carries nil until submit, which always writes it. After submit every
// sitting knows its journey, so "the latest sitting of journey X" is one
// indexed read rather than a walk through the answer log.
type ExamLink struct {
	elinkId    int64
	uid        int64
	profileId  int64
	examId     int64
	esessId    *int64
	shuffleMap *string

	reqExamType string
	reqGrade    int
	reqLevel    *int

	resTotalQuestions  *int
	resCorrectNumber   *int
	resSkippedNumber   *int
	resScorePercentage *int

	startedDt   mtime.MathTime
	submittedDt mtime.MathTime

	rptFlg *string

	kwords *string

	note        *string
	elinkStatus *string
	status      string
	createId    *int64
	createDt    mtime.MathTime
	modifyId    *int64
	modifyDt    mtime.MathTime
}

func NewExamLink() *ExamLink { return &ExamLink{} }

func (u *ExamLink) ElinkId() int64                  { return u.elinkId }
func (u *ExamLink) SetElinkId(id int64)             { u.elinkId = id }
func (u *ExamLink) Uid() int64                      { return u.uid }
func (u *ExamLink) SetUid(id int64)                 { u.uid = id }
func (u *ExamLink) ProfileId() int64                { return u.profileId }
func (u *ExamLink) SetProfileId(id int64)           { u.profileId = id }
func (u *ExamLink) ExamId() int64                   { return u.examId }
func (u *ExamLink) SetExamId(id int64)              { u.examId = id }
func (u *ExamLink) EsessId() *int64                 { return u.esessId }
func (u *ExamLink) SetEsessId(id *int64)            { u.esessId = id }
func (u *ExamLink) ShuffleMap() *string             { return u.shuffleMap }
func (u *ExamLink) SetShuffleMap(s *string)         { u.shuffleMap = s }
func (u *ExamLink) ReqExamType() string             { return u.reqExamType }
func (u *ExamLink) SetReqExamType(t string)         { u.reqExamType = t }
func (u *ExamLink) ReqGrade() int                   { return u.reqGrade }
func (u *ExamLink) SetReqGrade(g int)               { u.reqGrade = g }
func (u *ExamLink) ReqLevel() *int                  { return u.reqLevel }
func (u *ExamLink) SetReqLevel(l *int)              { u.reqLevel = l }
func (u *ExamLink) ResTotalQuestions() *int         { return u.resTotalQuestions }
func (u *ExamLink) SetResTotalQuestions(n *int)     { u.resTotalQuestions = n }
func (u *ExamLink) ResCorrectNumber() *int          { return u.resCorrectNumber }
func (u *ExamLink) SetResCorrectNumber(n *int)      { u.resCorrectNumber = n }
func (u *ExamLink) ResSkippedNumber() *int          { return u.resSkippedNumber }
func (u *ExamLink) SetResSkippedNumber(n *int)      { u.resSkippedNumber = n }
func (u *ExamLink) ResScorePercentage() *int        { return u.resScorePercentage }
func (u *ExamLink) SetResScorePercentage(n *int)    { u.resScorePercentage = n }
func (u *ExamLink) StartedDt() mtime.MathTime       { return u.startedDt }
func (u *ExamLink) SetStartedDt(t mtime.MathTime)   { u.startedDt = t }
func (u *ExamLink) SubmittedDt() mtime.MathTime     { return u.submittedDt }
func (u *ExamLink) SetSubmittedDt(t mtime.MathTime) { u.submittedDt = t }
func (u *ExamLink) Note() *string                   { return u.note }
func (u *ExamLink) SetNote(s *string)               { u.note = s }
func (u *ExamLink) RptFlg() *string                 { return u.rptFlg }
func (u *ExamLink) SetRptFlg(v *string)             { u.rptFlg = v }
func (u *ExamLink) Kwords() *string                 { return u.kwords }
func (u *ExamLink) SetKwords(v *string)             { u.kwords = v }
func (u *ExamLink) ElinkStatus() *string            { return u.elinkStatus }
func (u *ExamLink) SetElinkStatus(s *string)        { u.elinkStatus = s }
func (u *ExamLink) Status() string                  { return u.status }
func (u *ExamLink) SetStatus(s string)              { u.status = s }
func (u *ExamLink) CreateId() *int64                { return u.createId }
func (u *ExamLink) SetCreateId(id *int64)           { u.createId = id }
func (u *ExamLink) CreateDt() mtime.MathTime        { return u.createDt }
func (u *ExamLink) SetCreateDt(t mtime.MathTime)    { u.createDt = t }
func (u *ExamLink) ModifyId() *int64                { return u.modifyId }
func (u *ExamLink) SetModifyId(id *int64)           { u.modifyId = id }
func (u *ExamLink) ModifyDt() mtime.MathTime        { return u.modifyDt }
func (u *ExamLink) SetModifyDt(t mtime.MathTime)    { u.modifyDt = t }
