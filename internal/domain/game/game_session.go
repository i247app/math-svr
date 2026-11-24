package game

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// GameSession is one round of the game a child plays. It owns its question
// set (questionsJson, answer key included, in the order the child is
// served) and, once submitted, the graded answers and the result.
//
// reqLevel is nil when the client stated none. The score is over ALL the
// round's questions, skipped ones included. A round never counts toward the
// child's assessed ability: it is linked to no exam or journey.
type GameSession struct {
	audit.Audit

	gsessId            int64
	uid                int64
	profileId          int64
	reqGrade           int
	reqLevel           *int
	reqNumQues         int
	questionsJson      string
	answersJson        *string
	resTotalQuestions  *int
	resCorrectNumber   *int
	resSkippedNumber   *int
	resScorePercentage *int
	startedDt          time.Time
	submittedDt        time.Time
	gsessStatus        *string
}

func NewGameSession() *GameSession { return &GameSession{} }

func (s *GameSession) GsessId() int64     { return s.gsessId }
func (s *GameSession) SetGsessId(v int64) { s.gsessId = v }

func (s *GameSession) UID() int64     { return s.uid }
func (s *GameSession) SetUID(v int64) { s.uid = v }

func (s *GameSession) ProfileId() int64     { return s.profileId }
func (s *GameSession) SetProfileId(v int64) { s.profileId = v }

func (s *GameSession) ReqGrade() int     { return s.reqGrade }
func (s *GameSession) SetReqGrade(v int) { s.reqGrade = v }

func (s *GameSession) ReqLevel() *int     { return s.reqLevel }
func (s *GameSession) SetReqLevel(v *int) { s.reqLevel = v }

func (s *GameSession) ReqNumQues() int     { return s.reqNumQues }
func (s *GameSession) SetReqNumQues(v int) { s.reqNumQues = v }

func (s *GameSession) QuestionsJson() string     { return s.questionsJson }
func (s *GameSession) SetQuestionsJson(v string) { s.questionsJson = v }

func (s *GameSession) AnswersJson() *string     { return s.answersJson }
func (s *GameSession) SetAnswersJson(v *string) { s.answersJson = v }

func (s *GameSession) ResTotalQuestions() *int     { return s.resTotalQuestions }
func (s *GameSession) SetResTotalQuestions(v *int) { s.resTotalQuestions = v }

func (s *GameSession) ResCorrectNumber() *int     { return s.resCorrectNumber }
func (s *GameSession) SetResCorrectNumber(v *int) { s.resCorrectNumber = v }

func (s *GameSession) ResSkippedNumber() *int     { return s.resSkippedNumber }
func (s *GameSession) SetResSkippedNumber(v *int) { s.resSkippedNumber = v }

func (s *GameSession) ResScorePercentage() *int     { return s.resScorePercentage }
func (s *GameSession) SetResScorePercentage(v *int) { s.resScorePercentage = v }

func (s *GameSession) StartedDt() time.Time     { return s.startedDt }
func (s *GameSession) SetStartedDt(v time.Time) { s.startedDt = v }

func (s *GameSession) SubmittedDt() time.Time     { return s.submittedDt }
func (s *GameSession) SetSubmittedDt(v time.Time) { s.submittedDt = v }

func (s *GameSession) GsessStatus() *string     { return s.gsessStatus }
func (s *GameSession) SetGsessStatus(v *string) { s.gsessStatus = v }
