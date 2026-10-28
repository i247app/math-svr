package exam

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"math-ai.com/math-ai/internal/application/dto/aiquestion"
	"math-ai.com/math-ai/internal/shared/utils"
)

// ExamOption is one option of an exam question on the wire: its id (the key
// a child's answer names it by, stable wherever the option is shown — the
// client draws the letters A, B, C by position) and what is shown for it.
// ImageKey is the S3 key of an image shown with it (bank sets only);
// ImageURL is its presigned URL, set when a response is built and never
// stored. It is the shape of an option in the question bank's import file.
type ExamOption struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	ImageKey string `json:"image_key,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// ExamQuestion is the exam flow's wire shape for one question: what the
// model is asked to write (the OUTPUT FORMAT of the exam prompts), what
// ma_exam_pools.questions_json stores, and what the client is served.
//
// Its field names are the vocabulary of the columns they end up in
// (render_type, question_text, options, correct_answer, ...) and of the
// question bank, so a question can be read from the JSON, written to
// ma_questions / ma_exam_session_lines and rendered by the client without
// anybody renaming anything on the way.
// There is no reader for the older names (question_type, question_name,
// answers[{label, content}], right_answer_label(s), right_answer_content,
// right_option_ids, right_values):
// CheckSet refuses a set written under them rather than let it parse as empty questions.
//
// The exercise flow keeps its own names on aiquestion.Question, which it
// shares with the scorer; this type is the translation at the edge, and
// the conversions below are the single seam where the two meet.
//
//   - QuestionKind is SINGLE_CHOICE (the only kind a model writes),
//     MULTIPLE_CHOICE, TRUE_FALSE or NUMERIC_INPUT; absent = single-choice.
//   - CorrectAnswer is the key, in the shape of ma_question_pools.correct_answer
//     (ExamCorrectAnswer). It is blanked on a live exam.
//   - RightAnswerText is the model's working: it writes the correct option's
//     text first and picks the id from it (ReconcileKey checks the two
//     agree). It is never stored — the displayed key text is derived from
//     the options — and never served.
//   - QuestionImageKey / QuestionImageURL: the stem's image, as for options.
type ExamQuestion struct {
	QuestionNumber   int                `json:"question_number"`
	QuestionKind     string             `json:"question_kind,omitempty"`
	RenderType       string             `json:"render_type,omitempty"`
	QuestionText     string             `json:"question_text"`
	QuestionImageKey string             `json:"question_image_key,omitempty"`
	QuestionImageURL string             `json:"question_image_url,omitempty"`
	Options          []ExamOption       `json:"options"`
	CorrectAnswer    *ExamCorrectAnswer `json:"correct_answer,omitempty"`
	RightAnswerText  string             `json:"right_answer_text,omitempty"`
	QuestionTopic    string             `json:"question_topic,omitempty"`
	QuestionGrade    *int               `json:"question_grade,omitempty"`
}

// ExamCorrectAnswer is the key of one question, the shape of
// ma_question_pools.correct_answer: exactly one part is used, by kind —
// OptionIDs for SINGLE_CHOICE (one id) and MULTIPLE_CHOICE (one or more),
// Value for TRUE_FALSE, Values (accepted numbers) for NUMERIC_INPUT. A new
// kind is a new part here, not a new top-level key in the prompt's output.
type ExamCorrectAnswer struct {
	OptionIDs []string `json:"option_ids,omitempty"`
	Value     *bool    `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"`
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
		if n, err := utils.StringToIntErr(m[1]); err == nil {
			return &n
		}
	}
	return nil
}

// ToShared converts into the internal shape the scorer and normalisers
// operate on. The key text (RightAnswerContent) is the text of the option
// the key names, never the model's right_answer_text.
func (q ExamQuestion) ToShared() aiquestion.Question {
	out := aiquestion.Question{
		QuestionNumber: q.QuestionNumber,
		QuestionType:   q.RenderType,
		QuestionName:   q.QuestionText,
		Answers:        optionsToChoices(q.Options),
		QuestionTopic:  q.QuestionTopic,
		QuestionGrade:  q.QuestionGrade,
		Kind:           q.QuestionKind,
		ImageKey:       q.QuestionImageKey,
	}
	ca := q.CorrectAnswer
	if ca == nil {
		return out
	}
	switch q.QuestionKind {
	case aiquestion.KindMultipleChoice:
		out.RightOptionIDs = ca.OptionIDs
	case aiquestion.KindNumericInput:
		out.RightValues = ca.Values
	case aiquestion.KindTrueFalse:
		if ca.Value != nil {
			out.RightAnswerLabel = aiquestion.FalseOptionID
			if *ca.Value {
				out.RightAnswerLabel = aiquestion.TrueOptionID
			}
		}
	default:
		if len(ca.OptionIDs) > 0 {
			out.RightAnswerLabel = ca.OptionIDs[0]
		}
	}
	if out.RightAnswerLabel != "" {
		out.RightAnswerContent = q.optionText(out.RightAnswerLabel)
	}
	return out
}

// optionText is the text of the option with that id, "" when none.
func (q ExamQuestion) optionText(id string) string {
	for _, o := range q.Options {
		if o.ID == id {
			return o.Text
		}
	}
	return ""
}

// ExamQuestionFrom converts back out to the wire shape.
func ExamQuestionFrom(q aiquestion.Question) ExamQuestion {
	out := ExamQuestion{
		QuestionNumber:   q.QuestionNumber,
		QuestionKind:     q.Kind,
		RenderType:       q.QuestionType,
		QuestionText:     q.QuestionName,
		QuestionImageKey: q.ImageKey,
		Options:          choicesToOptions(q.Answers),
		QuestionTopic:    q.QuestionTopic,
		QuestionGrade:    q.QuestionGrade,
	}
	switch q.EffectiveKind() {
	case aiquestion.KindMultipleChoice:
		if len(q.RightOptionIDs) > 0 {
			out.CorrectAnswer = &ExamCorrectAnswer{OptionIDs: q.RightOptionIDs}
		}
	case aiquestion.KindNumericInput:
		if len(q.RightValues) > 0 {
			out.CorrectAnswer = &ExamCorrectAnswer{Values: q.RightValues}
		}
	case aiquestion.KindTrueFalse:
		if q.RightAnswerLabel != "" {
			v := q.RightAnswerLabel == aiquestion.TrueOptionID
			out.CorrectAnswer = &ExamCorrectAnswer{Value: &v}
		}
	default:
		if q.RightAnswerLabel != "" {
			out.CorrectAnswer = &ExamCorrectAnswer{OptionIDs: []string{q.RightAnswerLabel}}
		}
	}
	return out
}

func ToSharedQuestions(in []ExamQuestion) []aiquestion.Question {
	out := make([]aiquestion.Question, 0, len(in))
	for _, q := range in {
		out = append(out, q.ToShared())
	}
	return out
}

func FromSharedQuestions(in []aiquestion.Question) []ExamQuestion {
	out := make([]ExamQuestion, 0, len(in))
	for _, q := range in {
		out = append(out, ExamQuestionFrom(q))
	}
	return out
}

// OptionsFrom is a served option list on the wire. It is never nil, so a
// question without options (a numeric one) renders as [] and not null.
func OptionsFrom(in []aiquestion.AnswerChoice) []ExamOption { return choicesToOptions(in) }

func choicesToOptions(in []aiquestion.AnswerChoice) []ExamOption {
	out := make([]ExamOption, 0, len(in))
	for _, c := range in {
		out = append(out, ExamOption{ID: c.Label, Text: c.Content, ImageKey: c.ImageKey})
	}
	return out
}

func optionsToChoices(in []ExamOption) []aiquestion.AnswerChoice {
	out := make([]aiquestion.AnswerChoice, 0, len(in))
	for _, o := range in {
		out = append(out, aiquestion.AnswerChoice{Label: o.ID, Content: o.Text, ImageKey: o.ImageKey})
	}
	return out
}

// Usable reports whether the question can be shown and graded: it has a
// stem (text or image), a correct_answer whose part fits its kind, and,
// unless it is a numeric one, at least two options with distinct non-empty
// ids — a choice key naming options it has (exactly one for single-choice),
// a true/false key holding its value; a numeric one needs its accepted
// values. A question written under an older field vocabulary has no
// correct_answer (or no question_text / option ids), so it is not usable.
func (q ExamQuestion) Usable() bool {
	if strings.TrimSpace(q.QuestionText) == "" && q.QuestionImageKey == "" {
		return false
	}
	ca := q.CorrectAnswer
	if ca == nil {
		return false
	}
	if q.QuestionKind == aiquestion.KindNumericInput {
		return len(ca.Values) > 0
	}
	if len(q.Options) < 2 {
		return false
	}
	ids := make(map[string]bool, len(q.Options))
	for _, o := range q.Options {
		if o.ID == "" || ids[o.ID] {
			return false
		}
		ids[o.ID] = true
	}
	if q.QuestionKind == aiquestion.KindTrueFalse {
		return ca.Value != nil
	}
	if len(ca.OptionIDs) == 0 {
		return false
	}
	if q.QuestionKind != aiquestion.KindMultipleChoice && len(ca.OptionIDs) != 1 {
		return false
	}
	for _, id := range ca.OptionIDs {
		if !ids[id] {
			return false
		}
	}
	return true
}

// NormalizeIDs trims and lower-cases the option ids and the key of a
// model-written question, so "A" and "a " name the option "a" both in its
// list and in its key. Bank ids are already canonical and are not passed
// through here.
func (q *ExamQuestion) NormalizeIDs() {
	clean := func(id string) string { return strings.ToLower(strings.TrimSpace(id)) }
	for i := range q.Options {
		q.Options[i].ID = clean(q.Options[i].ID)
	}
	if q.CorrectAnswer != nil {
		for i := range q.CorrectAnswer.OptionIDs {
			q.CorrectAnswer.OptionIDs[i] = clean(q.CorrectAnswer.OptionIDs[i])
		}
	}
}

// ReconcileKey makes a model-written single-choice key agree with the
// model's right_answer_text. The prompt has the model write that text
// first and pick the id from it, so when the two disagree the text is the
// model's actual answer: if it is the text of exactly one option, the key
// is moved to that option and true is returned (the caller logs it). Any
// other mismatch is left alone — the key stays as written. A no-op for a
// question without right_answer_text or without a one-id key.
func (q *ExamQuestion) ReconcileKey() bool {
	want := strings.TrimSpace(q.RightAnswerText)
	if want == "" || q.CorrectAnswer == nil || len(q.CorrectAnswer.OptionIDs) != 1 {
		return false
	}
	if strings.TrimSpace(q.optionText(q.CorrectAnswer.OptionIDs[0])) == want {
		return false
	}
	match := ""
	for _, o := range q.Options {
		if strings.TrimSpace(o.Text) == want {
			if match != "" {
				return false
			}
			match = o.ID
		}
	}
	if match == "" {
		return false
	}
	q.CorrectAnswer.OptionIDs = []string{match}
	return true
}

// CheckSet refuses a question set with an unusable question, naming it.
// Every reader of a stored or generated set runs it, so a set written under
// the older vocabulary — or with a key that names no option — fails loudly
// instead of being served, or graded, as empty or unanswerable questions.
func CheckSet(qs []ExamQuestion) error {
	if len(qs) == 0 {
		return fmt.Errorf("exam: the question set is empty")
	}
	for _, q := range qs {
		if !q.Usable() {
			return fmt.Errorf("exam: question %d is not usable: it needs a question_text, at least two options with distinct ids, and a correct_answer that fits its kind (option_ids naming its options for a choice question) — a set written with an older field vocabulary? the prompt must emit question_text / options[{id, text}] / correct_answer{option_ids} / right_answer_text", q.QuestionNumber)
		}
	}
	return nil
}
