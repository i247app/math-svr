package exam

import (
	"encoding/json"

	"math-ai.com/math-ai/internal/application/dto/question"
	domain "math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/pagination"
	"math-ai.com/math-ai/internal/shared/utils"
)

// GenerateExamReq asks for one exam.
//
// Grade may be pinned by the client; left out, the server derives it —
// measured ability first, then the class the profile attends.
//
// Level is part of the contract but NOT yet part of the behaviour. The
// teaching team has not defined what a level is, so the server does not
// validate, resolve, prompt with or store it — req_level stays NULL no
// matter what is sent. The field is kept so the mobile contract does not
// have to change twice: once to drop it now and again to add it back when
// the rule lands. Do not read req.Level anywhere until then.
type GenerateExamReq struct {
	UserID       *int64 `json:"-"`
	ProfileID    int64  `json:"profile_id"`
	ExamType     string `json:"exam_type"`
	Grade        *int   `json:"grade,omitempty"`
	Level        *int   `json:"level,omitempty"` // accepted, ignored — see type doc
	NumQuestions int    `json:"num_questions,omitempty"`
	Semester     string `json:"semester,omitempty"`
	Program      string `json:"program,omitempty"`
}

// SubmitExamReq carries only the attempt id and the answers.
//
// It deliberately does NOT accept the questions. The answer key lives in
// ma_ai_exams and is read from there, so a client cannot declare its own
// correct answers and grade itself upward.
type SubmitExamReq struct {
	UserID       *int64                   `json:"-"`
	ProfileID    int64                    `json:"profile_id"`
	UserAiExamID int64                    `json:"user_ai_exam_id"`
	Answers      []question.StudentAnswer `json:"answers"`
}

type GetExamReq struct {
	UserID       *int64 `json:"-"`
	ProfileID    int64  `json:"profile_id"`
	UserAiExamID int64  `json:"user_ai_exam_id"`
}

// ListExamsReq pages a child's history. Status narrows to unfinished
// attempts ("IN_PROGRESS"), which is how the resume surface is built.
type ListExamsReq struct {
	UserID    *int64  `json:"-"`
	ProfileID int64   `json:"profile_id"`
	ExamType  *string `json:"exam_type,omitempty"`
	Status    *string `json:"status,omitempty"`
	Page      int     `json:"page,omitempty"`
	Size      int     `json:"size,omitempty"`
}

// GetExamStatsReq reads a child's journeys. Status narrows to one
// lifecycle state — "ACTIVE" is the natural filter for a dashboard
// showing where the child is now; omit it for the full history.
type GetExamStatsReq struct {
	UserID    *int64  `json:"-"`
	ProfileID int64   `json:"profile_id"`
	ExamType  *string `json:"exam_type,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// MarkExamJourneyReq ends one journey. Status is COMPLETE or CANCEL; the
// difference is what the next journey of that type inherits (see
// enum.UserExamStatusType).
type MarkExamJourneyReq struct {
	UserID     *int64 `json:"-"`
	ProfileID  int64  `json:"profile_id"`
	UserExamID int64  `json:"user_exam_id"`
	Status     string `json:"status"`
}

type MarkExamJourneyRes struct {
	Stats *ExamStats `json:"stats"`
}

// ExamResult is one sitting's score. Nil until the exam is submitted.
//
// TotalQuestions counts the questions the child ANSWERED — the exam
// model's counting rule — while SkippedNumber keeps the untouched ones
// visible so "3 of 3" is distinguishable from "10 of 10".
type ExamResult struct {
	TotalQuestions  int `json:"total_questions"`
	CorrectNumber   int `json:"correct_number"`
	SkippedNumber   int `json:"skipped_number"`
	ScorePercentage int `json:"score_percentage"`
}

// ExamResponse is one attempt as the client sees it: the sitting plus the
// question set it was drawn from, flattened into a single object so the
// app never has to join two payloads.
type ExamResponse struct {
	UserAiExamID int64  `json:"user_ai_exam_id"`
	AiExamID     int64  `json:"ai_exam_id"`
	ProfileID    int64  `json:"profile_id"`
	ExamType     string `json:"exam_type"`
	Grade        int    `json:"grade"`
	// Level mirrors req_level, which is NULL on every row today, so this is
	// always omitted on the wire. Kept for the same reason as the request
	// field: the contract is settled even though the value is not.
	Level *int `json:"level,omitempty"`

	Title     *string `json:"title,omitempty"`
	ShortText *string `json:"short_text,omitempty"`
	NumQues   int     `json:"num_questions"`

	Questions []ExamQuestion `json:"questions,omitempty"`
	Result    *ExamResult    `json:"result,omitempty"`

	Status      string `json:"status"`
	StartedDt   string `json:"started_dt,omitempty"`
	SubmittedDt string `json:"submitted_dt,omitempty"`
	CreateDt    string `json:"create_dt"`
}

// ExamAnswerDetail is one answered question on the review screen.
type ExamAnswerDetail struct {
	QuestionNumber     int     `json:"question_number"`
	QuestionType       *string `json:"question_type,omitempty"`
	QuestionName       *string `json:"question_name,omitempty"`
	QuestionTopic      *string `json:"question_topic,omitempty"`
	QuestionGrade      *int    `json:"question_grade,omitempty"`
	RightAnswerLabel   *string `json:"right_answer_label,omitempty"`
	RightAnswerContent *string `json:"right_answer_content,omitempty"`
	SelectedLabel      string  `json:"selected_label"`
	SelectedContent    *string `json:"selected_content,omitempty"`
	IsCorrect          bool    `json:"is_correct"`
}

// ExamStats is one JOURNEY — a stretch of one exam type the child worked
// through. UserExamID is what a client sends back to end it; Status says
// whether it is the open one (ACTIVE) or history (COMPLETE / CANCEL).
type ExamStats struct {
	UserExamID      int64   `json:"user_exam_id"`
	ExamType        string  `json:"exam_type"`
	Status          string  `json:"status"`
	TotalQuestions  int     `json:"total_questions"`
	CorrectNumber   int     `json:"correct_number"`
	SkippedNumber   int     `json:"skipped_number"`
	ScorePercentage *int    `json:"score_percentage,omitempty"`
	Review          *string `json:"review,omitempty"`
	Grade           *int    `json:"grade,omitempty"`
	LastSubmittedDt string  `json:"last_submitted_dt,omitempty"`
	EndedDt         string  `json:"ended_dt,omitempty"`
	CreateDt        string  `json:"create_dt"`
}

type GenerateExamRes struct {
	Exam *ExamResponse `json:"exam"`
}

// SubmitExamRes answers the question the child just asked — how did I do
// on THIS one — and hands back the updated running record beside it.
type SubmitExamRes struct {
	Exam  *ExamResponse `json:"exam"`
	Stats *ExamStats    `json:"stats,omitempty"`
}

type GetExamRes struct {
	Exam    *ExamResponse      `json:"exam"`
	Details []ExamAnswerDetail `json:"details,omitempty"`
}

type ListExamsRes struct {
	Exams      []*ExamResponse        `json:"exams"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type GetExamStatsRes struct {
	Stats []ExamStats `json:"stats"`
}

// AttemptToResponse flattens an attempt and its question set.
//
// includeAnswerKey gates right_answer / correct_answer: a live exam must
// not ship the key, or the client holds the answers to the paper it is
// currently sitting. A submitted one does, so the review screen can mark
// each question.
func AttemptToResponse(a *domain.UserAiExam, e *domain.AiExam, includeAnswerKey bool) *ExamResponse {
	if a == nil {
		return nil
	}

	res := &ExamResponse{
		UserAiExamID: a.UserAiExamId(),
		AiExamID:     a.AiExamId(),
		ProfileID:    a.ProfileId(),
		ExamType:     a.ReqExamType(),
		Grade:        a.ReqGrade(),
		Level:        a.ReqLevel(),
		CreateDt:     a.CreateDt().String(),
	}
	if s := a.UserAiExamStatus(); s != nil {
		res.Status = *s
	}
	if !a.StartedDt().IsZero() {
		res.StartedDt = a.StartedDt().String()
	}
	if !a.SubmittedDt().IsZero() {
		res.SubmittedDt = a.SubmittedDt().String()
	}

	if e != nil {
		res.Title = e.AiTitle()
		res.ShortText = e.AiShortText()
		res.NumQues = e.ReqNumQues()
		res.Questions = parseQuestions(e.AiQuestionsJson(), includeAnswerKey)
	}

	if total := a.ResTotalQuestions(); total != nil {
		res.Result = &ExamResult{
			TotalQuestions:  *total,
			CorrectNumber:   utils.DerefInt(a.ResCorrectNumber()),
			SkippedNumber:   utils.DerefInt(a.ResSkippedNumber()),
			ScorePercentage: utils.DerefInt(a.ResScorePercentage()),
		}
	}
	return res
}

// AttemptListToResponse maps a page of attempts against the question sets
// hydrated alongside them. The answer key is exposed only for submitted
// rows, so a history list cannot leak the answers to an exam still open.
func AttemptListToResponse(attempts []*domain.UserAiExam, aiExams map[int64]*domain.AiExam) []*ExamResponse {
	out := make([]*ExamResponse, 0, len(attempts))
	for _, a := range attempts {
		e := aiExams[a.AiExamId()]
		submitted := a.UserAiExamStatus() != nil && *a.UserAiExamStatus() == "SUBMITTED"
		res := AttemptToResponse(a, e, submitted)
		// A list renders cards, not papers: drop the questions blob so a
		// twenty-row page does not ship twenty question sets.
		res.Questions = nil
		out = append(out, res)
	}
	return out
}

func DetailsToResponse(details []*domain.UserExamDetail) []ExamAnswerDetail {
	out := make([]ExamAnswerDetail, 0, len(details))
	for _, d := range details {
		out = append(out, ExamAnswerDetail{
			QuestionNumber:     d.QuestionNumber(),
			QuestionType:       d.QuestionType(),
			QuestionName:       d.QuestionName(),
			QuestionTopic:      d.QuestionTopic(),
			QuestionGrade:      d.QuestionGrade(),
			RightAnswerLabel:   d.RightAnswerLabel(),
			RightAnswerContent: d.RightAnswerContent(),
			SelectedLabel:      d.SelectedLabel(),
			SelectedContent:    d.SelectedContent(),
			IsCorrect:          d.IsCorrect(),
		})
	}
	return out
}

func StatsToResponse(rows []*domain.UserExam) []ExamStats {
	out := make([]ExamStats, 0, len(rows))
	for _, r := range rows {
		s := ExamStats{
			UserExamID:      r.UserExamId(),
			ExamType:        r.ReqExamType(),
			CreateDt:        r.CreateDt().String(),
			TotalQuestions:  r.ResTotalQuestions(),
			CorrectNumber:   r.ResCorrectNumber(),
			SkippedNumber:   r.ResSkippedNumber(),
			ScorePercentage: r.ResScorePercentage(),
			Review:          r.ResReview(),
			Grade:           r.ResGrade(),
		}
		if st := r.UserExamStatus(); st != nil {
			s.Status = *st
		}
		if !r.LastSubmittedDt().IsZero() {
			s.LastSubmittedDt = r.LastSubmittedDt().String()
		}
		if !r.EndedDt().IsZero() {
			s.EndedDt = r.EndedDt().String()
		}
		out = append(out, s)
	}
	return out
}

func StatsToSingleResponse(row *domain.UserExam) *ExamStats {
	if row == nil {
		return nil
	}
	all := StatsToResponse([]*domain.UserExam{row})
	return &all[0]
}

// parseQuestions decodes the stored payload and, when the key must stay
// hidden, blanks it field by field rather than dropping the question —
// the child still needs the stem and the options to answer.
func parseQuestions(raw string, includeAnswerKey bool) []ExamQuestion {
	if raw == "" {
		return nil
	}
	var out []ExamQuestion
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	if !includeAnswerKey {
		for i := range out {
			out[i].RightAnswerLabel = ""
			out[i].RightAnswerContent = ""
		}
	}
	return out
}

// ExamProgressReq is the request for the learning-progress chart.
// UserID comes from the session, never the body. FromDt/ToDt are optional
// datetime strings; Tz is a numeric offset (IANA names are rejected — see
// enum.IsValidTzOffset). ExamType optionally narrows to one type; Limit
// caps the number of chart points.
type ExamProgressReq struct {
	UserID    *int64  `json:"-"`
	ProfileID int64   `json:"profile_id"`
	ExamType  *string `json:"exam_type"`
	FromDt    string  `json:"from_dt"`
	ToDt      string  `json:"to_dt"`
	Tz        string  `json:"tz"`
	Limit     int     `json:"limit"`
}

// ExamPoint is one chart point — a single submitted attempt. Sequence is
// the 1..N positional label. Score is on the 10-point scale the product
// displays; ScorePct is the raw 0-100 value it was derived from.
type ExamPoint struct {
	Sequence       int64   `json:"sequence"`
	UserAiExamID   int64   `json:"user_ai_exam_id"`
	ExamType       string  `json:"exam_type"`
	Grade          int     `json:"grade"`
	CompletedDt    string  `json:"completed_dt"`
	Score          float64 `json:"score"`
	ScorePct       int64   `json:"score_pct"`
	CorrectNumber  *int64  `json:"correct_number"`
	TotalQuestions *int64  `json:"total_questions"`
}

// ExamProgressSummary backs the banner. Nullable score fields serialise
// as null when the window has no data; AverageDelta compares against the
// prior same-size window and is null when that window is empty.
type ExamProgressSummary struct {
	Count           int64    `json:"count"`
	AverageScore    *float64 `json:"average_score"`
	AverageScorePct *float64 `json:"average_score_pct"`
	AverageDelta    *float64 `json:"average_delta"`
	HighestScore    *float64 `json:"highest_score"`
	HighestScorePct *int64   `json:"highest_score_pct"`
	HighestExamID   *int64   `json:"highest_user_ai_exam_id"`
	LowestScore     *float64 `json:"lowest_score"`
	Trend           string   `json:"trend"`
}

type ExamProgressRes struct {
	ProfileID int64               `json:"profile_id"`
	FromDt    string              `json:"from_dt"`
	ToDt      string              `json:"to_dt"`
	Tz        string              `json:"tz"`
	ExamType  *string             `json:"exam_type"`
	Limit     int                 `json:"limit"`
	Series    []ExamPoint         `json:"series"`
	Summary   ExamProgressSummary `json:"summary"`
}
