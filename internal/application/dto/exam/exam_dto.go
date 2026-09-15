package exam

import (
	"encoding/json"
	"sort"

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
// UserExamID names the journey a PRACTICE round is drawn inside, and is
// required for that type: the round is aimed at the journey's latest
// submitted sitting, so there is nothing to aim at without one. For a
// PRACTICE round Grade is ignored — it follows that sitting's grade.
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
	UserExamID   *int64 `json:"user_exam_id,omitempty"`
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

// GetExamReq reads either ONE sitting or ONE row of a journey:
//
//   - user_ai_exam_id → that sitting: the exam as served, plus its answers.
//   - user_exam_id    → that journey: its running totals, every sitting
//     that fed them, and every question answered across them.
//     ReqExamType picks which row of the journey — ASSESSMENT (the
//     default, so older clients keep working) or PRACTICE — since the
//     two share the id. It is ignored when a sitting is asked for.
type GetExamReq struct {
	UserID       *int64  `json:"-"`
	ProfileID    int64   `json:"profile_id"`
	UserAiExamID int64   `json:"user_ai_exam_id,omitempty"`
	UserExamID   int64   `json:"user_exam_id,omitempty"`
	ExamType     *string `json:"exam_type,omitempty"`
}

// ListExamsReq pages a child's history. Status narrows to unfinished
// attempts ("IN_PROGRESS"), which is how the resume surface is built.
// UserExamID narrows to one journey, and is required when ExamType is
// PRACTICE: practice rounds only exist inside a journey, so "the practice
// rounds" is only a question about one.
type ListExamsReq struct {
	UserID     *int64  `json:"-"`
	ProfileID  int64   `json:"profile_id"`
	ExamType   *string `json:"exam_type,omitempty"`
	UserExamID *int64  `json:"user_exam_id,omitempty"`
	Status     *string `json:"status,omitempty"`
	Page       int     `json:"page,omitempty"`
	Size       int     `json:"size,omitempty"`
}

// GetExamStatsReq reads a child's journeys. Status narrows to one
// lifecycle state — "ACTIVE" is the natural filter for a dashboard
// showing where the child is now; omit it for the full history.
//
// ExamType names a JOURNEY type. PRACTICE is not one — its totals ride
// along on the ASSESSMENT journey they belong to (ExamStats.Practice) —
// so asking for it is rejected rather than answered with orphan rows.
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
// UserAiExamID says which sitting it came from — redundant on a single
// sitting's review, load-bearing on a journey's, where the rows span
// many.
//
// Answers is every option the question offered, in the order and under
// the labels the child saw, so the screen can lay all of them out and
// mark the right one and the chosen one among them. It is resolved at
// read time from the question set the sitting was drawn from — the log
// itself stores only the two labels that matter for grading — and is
// omitted when that set can no longer be found.
type ExamAnswerDetail struct {
	UserAiExamID       int64                   `json:"user_ai_exam_id"`
	QuestionNumber     int                     `json:"question_number"`
	QuestionType       *string                 `json:"question_type,omitempty"`
	QuestionName       *string                 `json:"question_name,omitempty"`
	Answers            []question.AnswerChoice `json:"answers,omitempty"`
	QuestionTopic      *string                 `json:"question_topic,omitempty"`
	QuestionGrade      *int                    `json:"question_grade,omitempty"`
	RightAnswerLabel   *string                 `json:"right_answer_label,omitempty"`
	RightAnswerContent *string                 `json:"right_answer_content,omitempty"`
	SelectedLabel      string                  `json:"selected_label"`
	SelectedContent    *string                 `json:"selected_content,omitempty"`
	IsCorrect          bool                    `json:"is_correct"`
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
	// Practice is the journey's PRACTICE row, present once the child has
	// submitted a practice round in it. Only an ASSESSMENT journey carries
	// one; it shares user_exam_id and never has a grade.
	Practice *ExamStats `json:"practice,omitempty"`
	// InProgressExams are the journey's sittings — ASSESSMENT or PRACTICE,
	// see each one's exam_type — that were handed out and never submitted,
	// oldest first; absent when there are none. Each carries its questions
	// as served, so the app can put the child back in front of the paper
	// without another read.
	InProgressExams []*ExamResponse `json:"in_progress_exams,omitempty"`
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

// GetExamRes is shaped by which id was asked for.
//
//   - by user_ai_exam_id: Exam + Details (Stats and Exams stay empty).
//   - by user_exam_id:    Stats + Exams + Details (Exam stays empty). Exams
//     are cards — no questions blob — in chronological order; Details span
//     every sitting and carry user_ai_exam_id so they can be grouped.
type GetExamRes struct {
	Exam    *ExamResponse      `json:"exam,omitempty"`
	Stats   *ExamStats         `json:"stats,omitempty"`
	Exams   []*ExamResponse    `json:"exams,omitempty"`
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
		// The sitting's own ordering, not the stored one: two children
		// served the same cached set each see their own arrangement.
		res.Questions = servedQuestions(e.AiQuestionsJson(), shuffleOf(a), includeAnswerKey)
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

// DetailsToResponse renders the per-question log the way the child SAW
// it. The log is stored canonically (so analytics line up with the shared
// question set), but a review screen has to say "câu 3" and "đáp án B"
// in the numbering the child actually met, or nothing on it will match
// their memory. shuffles maps each sitting to its ordering; a sitting
// absent from the map — or mapped to nil — is rendered canonically.
//
// aiExams supplies the question sets the sittings were drawn from, keyed
// by ai_exam_id, so every row can carry its full option list. Each set is
// parsed once, however many rows point at it.
func DetailsToResponse(details []*domain.UserExamDetail, shuffles map[int64]*question.Shuffle, aiExams map[int64]*domain.AiExam) []ExamAnswerDetail {
	questions := newQuestionIndex(aiExams)

	out := make([]ExamAnswerDetail, 0, len(details))
	for _, d := range details {
		sh := shuffles[d.UserAiExamId()]

		var answers []question.AnswerChoice
		if q, ok := questions.lookup(d.AiExamId(), d.QuestionNumber()); ok {
			answers = sh.ServedAnswers(d.QuestionNumber(), q)
		}

		out = append(out, ExamAnswerDetail{
			UserAiExamID:       d.UserAiExamId(),
			QuestionNumber:     sh.ServedNumber(d.QuestionNumber()),
			QuestionType:       d.QuestionType(),
			QuestionName:       d.QuestionName(),
			Answers:            answers,
			QuestionTopic:      d.QuestionTopic(),
			QuestionGrade:      d.QuestionGrade(),
			RightAnswerLabel:   servedLabelPtr(sh, d.QuestionNumber(), d.RightAnswerLabel()),
			RightAnswerContent: d.RightAnswerContent(),
			SelectedLabel:      sh.ServedLabel(d.QuestionNumber(), d.SelectedLabel()),
			SelectedContent:    d.SelectedContent(),
			IsCorrect:          d.IsCorrect(),
		})
	}

	// Rows arrive in canonical order; the student saw them in served
	// order. Re-sort per sitting so the review reads like the paper did.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UserAiExamID != out[j].UserAiExamID {
			return out[i].UserAiExamID < out[j].UserAiExamID
		}
		return out[i].QuestionNumber < out[j].QuestionNumber
	})
	return out
}

// questionIndex parses each question set lazily and at most once, then
// answers "which question is canonical number N of set S" in O(1). A
// journey review touches every row of every sitting, so this is the
// difference between parsing each set once and parsing it per row.
type questionIndex struct {
	sets   map[int64]*domain.AiExam
	parsed map[int64]map[int]question.Question
}

func newQuestionIndex(sets map[int64]*domain.AiExam) *questionIndex {
	return &questionIndex{sets: sets, parsed: make(map[int64]map[int]question.Question, len(sets))}
}

func (x *questionIndex) lookup(aiExamID int64, canonicalQN int) (question.Question, bool) {
	byNumber, ok := x.parsed[aiExamID]
	if !ok {
		byNumber = x.parse(aiExamID)
		x.parsed[aiExamID] = byNumber
	}
	q, ok := byNumber[canonicalQN]
	return q, ok
}

// parse decodes one set. A missing or unreadable set yields an empty
// index rather than an error: the review still renders, just without the
// option list for those rows — which is the graceful outcome when a
// shared question set has been retired from under a child's history.
func (x *questionIndex) parse(aiExamID int64) map[int]question.Question {
	set, ok := x.sets[aiExamID]
	if !ok || set == nil {
		return nil
	}
	var stored []ExamQuestion
	if err := json.Unmarshal([]byte(set.AiQuestionsJson()), &stored); err != nil {
		return nil
	}
	byNumber := make(map[int]question.Question, len(stored))
	for _, q := range stored {
		byNumber[q.QuestionNumber] = q.ToShared()
	}
	return byNumber
}

// ShufflesOf parses each attempt's stored ordering into the map
// DetailsToResponse consumes. A malformed ordering is treated as none:
// the review then renders that sitting canonically rather than failing
// the whole screen over one row.
func ShufflesOf(attempts ...*domain.UserAiExam) map[int64]*question.Shuffle {
	out := make(map[int64]*question.Shuffle, len(attempts))
	for _, a := range attempts {
		if a == nil {
			continue
		}
		out[a.UserAiExamId()] = shuffleOf(a)
	}
	return out
}

func shuffleOf(a *domain.UserAiExam) *question.Shuffle {
	sh, err := question.ParseShuffle(a.ShuffleMap())
	if err != nil {
		return nil
	}
	return sh
}

func servedLabelPtr(sh *question.Shuffle, canonicalQN int, label *string) *string {
	if label == nil {
		return nil
	}
	served := sh.ServedLabel(canonicalQN, *label)
	return &served
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

// JourneyStatsToResponse renders journeys with their practice row and
// unfinished sittings tucked inside, so the client sees one entry per
// journey rather than two rows that happen to share an id.
//
// An unfinished sitting ships with its answer key, the same as the
// hand-out response does — the two must agree, or a child resuming an
// exam would see a different paper than the one they started.
func JourneyStatsToResponse(journeys []domain.JourneyStats, aiExams map[int64]*domain.AiExam) []ExamStats {
	out := make([]ExamStats, 0, len(journeys))
	for _, j := range journeys {
		s := StatsToSingleResponse(j.Journey)
		if s == nil {
			continue
		}
		s.Practice = StatsToSingleResponse(j.Practice)
		for _, a := range j.InProgress {
			s.InProgressExams = append(s.InProgressExams, AttemptToResponse(a, aiExams[a.AiExamId()], true))
		}
		out = append(out, *s)
	}
	return out
}

// parseQuestions decodes the stored payload and, when the key must stay
// hidden, blanks it field by field rather than dropping the question —
// the child still needs the stem and the options to answer.
// servedQuestions decodes the stored (canonical) set, arranges it the way
// this sitting was served, and — when the key must stay hidden — blanks
// it field by field rather than dropping the question: the child still
// needs the stem and the options to answer.
func servedQuestions(raw string, sh *question.Shuffle, includeAnswerKey bool) []ExamQuestion {
	if raw == "" {
		return nil
	}
	var stored []ExamQuestion
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return nil
	}
	out := FromSharedQuestions(sh.Apply(ToSharedQuestions(stored)))
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
	UserID     *int64  `json:"-"`
	ProfileID  int64   `json:"profile_id"`
	ExamType   *string `json:"exam_type"`
	UserExamID *int64  `json:"user_exam_id,omitempty"` // required with PRACTICE
	FromDt     string  `json:"from_dt"`
	ToDt       string  `json:"to_dt"`
	Tz         string  `json:"tz"`
	Limit      int     `json:"limit"`
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
