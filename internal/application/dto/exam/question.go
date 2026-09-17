package exam

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"math-ai.com/math-ai/internal/application/dto/question"
)

// ExamQuestion is the exam flow's wire shape for one MCQ item.
//
// It exists because the exam vocabulary is its own: the field names here
// match the columns they end up in (right_answer_label,
// right_answer_content, question_topic, question_grade),
// so a question can be read from the JSON, written to ma_user_exam_details
// and rendered by the client without anybody renaming anything on the way.
//
// The quiz and exercise flows keep the older names (right_answer,
// correct_answer, topic, difficulty) on question.Question, which they
// share. That struct stays untouched: it is the model's contract with two
// other aggregates and with an already-shipped client, and renaming its
// tags to suit a third would rewrite what those two store and return.
//
// So this type is a translation at the edges only. Everything in between —
// parsing, normalisation, scoring — works on question.Question, and the
// conversions below are the single seam where the two vocabularies meet.
type ExamQuestion struct {
	QuestionNumber     int                     `json:"question_number"`
	QuestionType       string                  `json:"question_type,omitempty"`
	QuestionName       string                  `json:"question_name"`
	Answers            []question.AnswerChoice `json:"answers"`
	RightAnswerLabel   string                  `json:"right_answer_label,omitempty"`
	RightAnswerContent string                  `json:"right_answer_content,omitempty"`
	QuestionTopic      string                  `json:"question_topic,omitempty"`
	QuestionGrade      *int                    `json:"question_grade,omitempty"`
}

// UnmarshalJSON reads question_grade as either a number or a band label.
// The prompt asks the model for the label ("Mẫu giáo", "Lớp 3") because
// that is the vocabulary the rules are written in; the column is an int.
// A label that is neither is read as absent rather than failing the whole
// round — the server re-stamps question_grade from the question's
// position afterwards and only logs what the model claimed.
func (q *ExamQuestion) UnmarshalJSON(data []byte) error {
	type plain ExamQuestion
	aux := struct {
		*plain
		QuestionGrade json.RawMessage `json:"question_grade,omitempty"`
	}{plain: (*plain)(q)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	q.QuestionGrade = parseGrade(aux.QuestionGrade)
	return nil
}

var gradeLabelRe = regexp.MustCompile(`(?i)^\s*(?:lớp|grade)\s*(\d+)\s*$`)

// parseGrade maps the model's question_grade to the stored int, or nil.
func parseGrade(raw json.RawMessage) *int {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return &n
	}
	var label string
	if err := json.Unmarshal(raw, &label); err != nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "mẫu giáo", "mầm non", "kindergarten":
		zero := 0
		return &zero
	}
	if m := gradeLabelRe.FindStringSubmatch(label); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return &n
		}
	}
	return nil
}

// ToShared converts into the internal shape the scorer and normalisers
// operate on.
func (q ExamQuestion) ToShared() question.Question {
	return question.Question{
		QuestionNumber:     q.QuestionNumber,
		QuestionType:       q.QuestionType,
		QuestionName:       q.QuestionName,
		Answers:            q.Answers,
		RightAnswerLabel:   q.RightAnswerLabel,
		RightAnswerContent: q.RightAnswerContent,
		QuestionTopic:      q.QuestionTopic,
		QuestionGrade:      q.QuestionGrade,
	}
}

// ExamQuestionFrom converts back out to the wire shape.
func ExamQuestionFrom(q question.Question) ExamQuestion {
	return ExamQuestion{
		QuestionNumber:     q.QuestionNumber,
		QuestionType:       q.QuestionType,
		QuestionName:       q.QuestionName,
		Answers:            q.Answers,
		RightAnswerLabel:   q.RightAnswerLabel,
		RightAnswerContent: q.RightAnswerContent,
		QuestionTopic:      q.QuestionTopic,
		QuestionGrade:      q.QuestionGrade,
	}
}

func ToSharedQuestions(in []ExamQuestion) []question.Question {
	out := make([]question.Question, 0, len(in))
	for _, q := range in {
		out = append(out, q.ToShared())
	}
	return out
}

func FromSharedQuestions(in []question.Question) []ExamQuestion {
	out := make([]ExamQuestion, 0, len(in))
	for _, q := range in {
		out = append(out, ExamQuestionFrom(q))
	}
	return out
}
