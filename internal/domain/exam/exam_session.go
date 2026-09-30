package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// JourneyKey names the slot an open journey occupies: a submission folds
// into the ACTIVE journey of its slot, and a slot holds at most one.
//
// For most types the slot is (user, profile, type). A GRADE review adds
// the grade: a Grade 1 review and a Grade 2 review measure different
// things, so a child working both keeps two open journeys side by side
// and neither score is diluted by the other's questions. The grade is
// therefore the identity of a GRADE journey, fixed when it opens —
// current_grade is never moved on one.
//
// uk_active_journey holds the same rule in the database; build a key
// with JourneyKeyOf so code and schema cannot disagree about it.
type JourneyKey struct {
	Uid       int64
	ProfileId int64
	ExamType  string
	// Grade narrows the slot to one grade; set on a GRADE key only. nil
	// on a GRADE key means "whichever grade the child worked last": the
	// most recently touched open journey of the type.
	Grade *int
}

// JourneyKeyOf builds the slot key for one child and exam type. grade is
// kept only for GRADE — for every other type the child has one open
// journey whatever the grade, and passing it is harmless.
func JourneyKeyOf(uid, profileId int64, examType string, grade *int) JourneyKey {
	k := JourneyKey{Uid: uid, ProfileId: profileId, ExamType: examType}
	if examType == string(enum.ExamTypeGrade) && grade != nil {
		g := *grade
		k.Grade = &g
	}
	return k
}

// ExamSession is one JOURNEY: a stretch of one exam type that a (user,
// profile) pair works through and then closes. Every submission of that
// type folds into the open journey; once it is marked COMPLETE or CANCEL,
// the next submission opens a new one. A user account holds several
// profiles (one per child), so the pair, not the user alone, is the
// subject of the statistics.
//
// At most one journey per slot (JourneyKey) is ACTIVE at a time — the
// database holds that line (uk_active_journey), not this type.
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
type ExamSession struct {
	esessId      int64
	uid          int64
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

	note        *string
	esessStatus *string
	status      string
	createId    *int64
	createDt    mtime.MathTime
	modifyId    *int64
	modifyDt    mtime.MathTime
}

func NewExamSession() *ExamSession { return &ExamSession{} }

func (u *ExamSession) EsessId() int64                      { return u.esessId }
func (u *ExamSession) SetEsessId(id int64)                 { u.esessId = id }
func (u *ExamSession) Uid() int64                          { return u.uid }
func (u *ExamSession) SetUid(id int64)                     { u.uid = id }
func (u *ExamSession) ProfileId() int64                    { return u.profileId }
func (u *ExamSession) SetProfileId(id int64)               { u.profileId = id }
func (u *ExamSession) ReqExamType() string                 { return u.reqExamType }
func (u *ExamSession) SetReqExamType(t string)             { u.reqExamType = t }
func (u *ExamSession) ResTotalQuestions() int              { return u.resTotalQuestions }
func (u *ExamSession) SetResTotalQuestions(n int)          { u.resTotalQuestions = n }
func (u *ExamSession) ResCorrectNumber() int               { return u.resCorrectNumber }
func (u *ExamSession) SetResCorrectNumber(n int)           { u.resCorrectNumber = n }
func (u *ExamSession) ResSkippedNumber() int               { return u.resSkippedNumber }
func (u *ExamSession) SetResSkippedNumber(n int)           { u.resSkippedNumber = n }
func (u *ExamSession) ResScorePercentage() *int            { return u.resScorePercentage }
func (u *ExamSession) SetResScorePercentage(n *int)        { u.resScorePercentage = n }
func (u *ExamSession) ResReview() *string                  { return u.resReview }
func (u *ExamSession) SetResReview(s *string)              { u.resReview = s }
func (u *ExamSession) CurrentGrade() *int                  { return u.currentGrade }
func (u *ExamSession) SetCurrentGrade(g *int)              { u.currentGrade = g }
func (u *ExamSession) CurrentLevel() *int                  { return u.currentLevel }
func (u *ExamSession) SetCurrentLevel(l *int)              { u.currentLevel = l }
func (u *ExamSession) LastSubmittedDt() mtime.MathTime     { return u.lastSubmittedDt }
func (u *ExamSession) SetLastSubmittedDt(t mtime.MathTime) { u.lastSubmittedDt = t }
func (u *ExamSession) EndedDt() mtime.MathTime             { return u.endedDt }
func (u *ExamSession) SetEndedDt(t mtime.MathTime)         { u.endedDt = t }
func (u *ExamSession) Note() *string                       { return u.note }
func (u *ExamSession) SetNote(s *string)                   { u.note = s }
func (u *ExamSession) RptFlg() *string                     { return u.rptFlg }
func (u *ExamSession) SetRptFlg(v *string)                 { u.rptFlg = v }
func (u *ExamSession) Kwords() *string                     { return u.kwords }
func (u *ExamSession) SetKwords(v *string)                 { u.kwords = v }
func (u *ExamSession) EsessStatus() *string                { return u.esessStatus }
func (u *ExamSession) SetEsessStatus(s *string)            { u.esessStatus = s }
func (u *ExamSession) Status() string                      { return u.status }
func (u *ExamSession) SetStatus(s string)                  { u.status = s }
func (u *ExamSession) CreateId() *int64                    { return u.createId }
func (u *ExamSession) SetCreateId(id *int64)               { u.createId = id }
func (u *ExamSession) CreateDt() mtime.MathTime            { return u.createDt }
func (u *ExamSession) SetCreateDt(t mtime.MathTime)        { u.createDt = t }
func (u *ExamSession) ModifyId() *int64                    { return u.modifyId }
func (u *ExamSession) SetModifyId(id *int64)               { u.modifyId = id }
func (u *ExamSession) ModifyDt() mtime.MathTime            { return u.modifyDt }
func (u *ExamSession) SetModifyDt(t mtime.MathTime)        { u.modifyDt = t }
