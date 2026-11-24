package dto

import (
	"encoding/json"
	"slices"
	"strings"

	domain "math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/mtime"
	"math-ai.com/math-ai/internal/shared/pagination"
	"math-ai.com/math-ai/internal/shared/utils"
)

// StartGameReq opens a round. Grade nil = the profile's; Level nil = any
// level; NumQuestions nil = enum.GameNumQuestionsDefault.
type StartGameReq struct {
	UID          *int64 `json:"-"`
	ProfileID    int64  `json:"profile_id"`
	Grade        *int   `json:"grade,omitempty"`
	Level        *int   `json:"level,omitempty"`
	NumQuestions *int   `json:"num_questions,omitempty"`
}

// SubmitGameReq submits a round. A question left out of Answers is skipped;
// no answers at all is a valid submit that scores 0.
type SubmitGameReq struct {
	UID       *int64       `json:"-"`
	ProfileID int64        `json:"profile_id"`
	GsessID   int64        `json:"gsess_id"`
	Answers   []ExamAnswer `json:"answers"`
}

// ListGamesReq pages a child's COMPLETE rounds, newest first.
type ListGamesReq struct {
	UID       *int64 `json:"-"`
	ProfileID int64  `json:"profile_id"`
	pagination.Request
}

type GetGameReq struct {
	UID       *int64 `json:"-"`
	ProfileID int64  `json:"profile_id"`
	GsessID   int64  `json:"gsess_id"`
}

// GameSessionResponse is one round. Questions always carry the answer key —
// the game client needs it while the child plays (product owner's decision,
// 2026-11-24) — and the explanation only once the round is COMPLETE; a list
// leaves the questions out.
type GameSessionResponse struct {
	Language     enum.LanguageType `json:"lang"`
	GsessID      int64             `json:"gsess_id"`
	ProfileID    int64             `json:"profile_id"`
	Grade        int               `json:"grade"`
	Level        *int              `json:"level"`
	NumQuestions int               `json:"num_questions"`
	Status       string            `json:"status"`
	StartedDt    mtime.MathTime    `json:"started_dt"`
	SubmittedDt  *mtime.MathTime   `json:"submitted_dt"`
	Questions    []ExamQuestion    `json:"questions,omitempty"`
	Result       *GameResult       `json:"result,omitempty"`
}

// GameResult is a COMPLETE round's score, over ALL its questions.
type GameResult struct {
	TotalQuestions  int `json:"total_questions"`
	CorrectNumber   int `json:"correct_number"`
	SkippedNumber   int `json:"skipped_number"`
	ScorePercentage int `json:"score_percentage"`
}

// GameAnswer is one graded answer, as stored in answers_json (answered
// questions only).
type GameAnswer struct {
	QuestionNumber  int      `json:"question_number"`
	ChildOptionIDs  []string `json:"child_option_ids"`
	ChildAnswerText string   `json:"child_answer_text"`
	IsCorrect       bool     `json:"is_correct"`
}

// GameAnswerDetail is how the child did on one question of a COMPLETE round.
// There is one per question, skipped ones included; the stem, the options and
// the key are on the question of the same number.
type GameAnswerDetail struct {
	GameAnswer
	IsSkipped bool `json:"is_skipped"`
}

type GameSessionRes struct {
	GameSession *GameSessionResponse `json:"game_session"`
	Details     []GameAnswerDetail   `json:"details,omitempty"`
}

type ListGamesRes struct {
	GameSessions []*GameSessionResponse       `json:"game_sessions"`
	Pagination   *pagination.Pagination       `json:"pagination,omitempty"`
	Cursor       *pagination.CursorPagination `json:"cursor,omitempty"`
}

func isGameComplete(s *domain.GameSession) bool {
	return utils.DerefString(s.GsessStatus()) == enum.GsessStatusComplete.String()
}

// GameSessionToResponse renders a round in lang. The answer key is always
// shown; whether the explanation is follows from the round's own status,
// never from the caller.
func GameSessionToResponse(s *domain.GameSession, lang enum.LanguageType) *GameSessionResponse {
	if s == nil {
		return nil
	}
	res := &GameSessionResponse{
		Language:     LanguageOf(lang),
		GsessID:      s.GsessId(),
		ProfileID:    s.ProfileId(),
		Grade:        s.ReqGrade(),
		Level:        s.ReqLevel(),
		NumQuestions: s.ReqNumQues(),
		Status:       utils.DerefString(s.GsessStatus()),
		StartedDt:    mtime.NewTime(s.StartedDt()),
		SubmittedDt:  mtime.NewMathTimePtr(s.SubmittedDt()),
		// Stored in served order: there is no shuffle to apply.
		Questions: servedQuestions(s.QuestionsJson(), nil, true, lang),
	}
	if !isGameComplete(s) {
		for i := range res.Questions {
			res.Questions[i].Explanation = ""
		}
	}
	if total := s.ResTotalQuestions(); total != nil {
		res.Result = &GameResult{
			TotalQuestions:  *total,
			CorrectNumber:   utils.DerefInt(s.ResCorrectNumber()),
			SkippedNumber:   utils.DerefInt(s.ResSkippedNumber()),
			ScorePercentage: utils.DerefInt(s.ResScorePercentage()),
		}
	}
	return res
}

func GameSessionsToResponse(rows []*domain.GameSession, lang enum.LanguageType) []*GameSessionResponse {
	out := make([]*GameSessionResponse, 0, len(rows))
	for _, s := range rows {
		out = append(out, GameSessionToResponse(s, lang))
	}
	return out
}

// GameSessionToRes is the full view of one round: the round and, once it is
// COMPLETE, how the child did on each question.
func GameSessionToRes(s *domain.GameSession, lang enum.LanguageType) *GameSessionRes {
	res := &GameSessionRes{GameSession: GameSessionToResponse(s, lang)}
	if s == nil || !isGameComplete(s) {
		return res
	}
	var stored []GameAnswer
	if raw := utils.DerefString(s.AnswersJson()); raw != "" {
		// Written by the submit command; an unreadable value shows every
		// question as skipped rather than failing the read.
		_ = json.Unmarshal([]byte(raw), &stored)
	}
	byNumber := make(map[int]GameAnswer, len(stored))
	for _, a := range stored {
		byNumber[a.QuestionNumber] = a
	}
	for _, q := range res.GameSession.Questions {
		a, answered := byNumber[q.QuestionNumber]
		a.QuestionNumber = q.QuestionNumber
		if a.ChildOptionIDs == nil {
			a.ChildOptionIDs = []string{}
		}
		// A choice answer reads as the (localized) texts of its options; a
		// numeric one is the number the child typed.
		if len(a.ChildOptionIDs) > 0 {
			var texts []string
			for _, o := range q.Options {
				if slices.Contains(a.ChildOptionIDs, o.ID) {
					texts = append(texts, o.Text)
				}
			}
			a.ChildAnswerText = strings.Join(texts, ContentSeparator)
		}
		res.Details = append(res.Details, GameAnswerDetail{GameAnswer: a, IsSkipped: !answered})
	}
	return res
}
