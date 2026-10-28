// Package question holds the MCQ contract shared by every aggregate that
// consumes AI-generated questions — quizzes, exams, and classroom
// exercises. It exists so those modules depend on one neutral package
// instead of on each other: before it, module/exercise imported
// module/quiz purely to reuse the parser, which made the quiz package
// impossible to rename or retire.
//
// The JSON tags here are a wire contract in two directions at once: the
// mobile client reads them, and the LLM is prompted to emit them. Renaming
// a Go field is free; renaming a tag is not.
package aiquestion

import (
	"slices"
	"strings"
)

// AnswerChoice is one option in a multiple-choice question. Label is the
// key the student selects; Content is the text that gets displayed.
//
// In an EXAM question Label holds the option's ID (the wire name is "id"):
// a short stable key ("a".."d" for a model-written set, the bank's own id
// for a bank set). It is never renamed and never shown — the client draws
// the letters A, B, C by position. In an exercise it is the letter "A".."D".
//
// ImageKey is the S3 key of an image shown with the option (question bank
// sets only). ImageURL is its presigned URL, set when a response is built
// and never stored. Both are empty for every model-written option, and
// omitted from the JSON then.
type AnswerChoice struct {
	Label    string `json:"label"`
	Content  string `json:"content"`
	ImageKey string `json:"image_key,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// Question type discriminators. Type tells the client how to render a
// question and its answers; it never affects grading, which is always
// label-based. Missing or unknown values normalize to TypeArithmetic so
// legacy rows and any schema drift from the model degrade to a plain text
// question.
//
//   - ARITHMETIC     — text-only stem (numbers + operators), text answers.
//   - COUNT          — stem embeds emoji / [icon:NAME] tokens to count or add.
//   - PICK_BY_ICON   — text stem; each answer's content embeds emoji / icons.
//   - IDENTIFY_SHAPE — stem is a single [icon:NAME] token; text answers.
//   - FRACTION       — stem and answers carry LaTeX \frac{a}{b} in $…$, so the
//     client renders them as stacked fractions instead of "3/4".
const (
	TypeArithmetic    = "ARITHMETIC"
	TypeCount         = "COUNT"
	TypePickByIcon    = "PICK_BY_ICON"
	TypeIdentifyShape = "IDENTIFY_SHAPE"
	TypeFraction      = "FRACTION"
)

// Question is one MCQ item as produced by the llm. RightAnswer is the
// label of the correct option and is only included in responses once the
// round has been graded — otherwise it would leak the answer key.
//
// CorrectAnswer, Topic and Difficulty are absent on older rows; the
// deterministic scorer degrades gracefully without them (label match
// only, generic review wording). All are omitempty so the payload of a
// row that lacks them is byte-for-byte unchanged.
//
// Grade is the band this single question targets. It is usually the band
// the whole round was generated for, but an ASSESSMENT deliberately sets
// two of its questions one band higher to probe upward, which is why the
// value lives per question rather than on the round. Quizzes and exercises
// never populate it, so their payloads are unaffected.
//
// There is no per-question level: the exam schema reserves a nullable
// question_level column, but no rule defines it yet, so the prompt does
// not ask for it and nothing stores it. A stored blob that still carries a
// "question_level" key (written before the axis was removed) decodes fine —
// unknown keys are dropped.
//
// Kind, RightAnswerLabels and RightValues are only set by exam sets built
// from the question bank; a question without a Kind is a single-choice one,
// which is every question the model emits and every exercise.
type Question struct {
	QuestionNumber     int            `json:"question_number"`
	QuestionType       string         `json:"question_type,omitempty"`
	QuestionName       string         `json:"question_name"`
	Answers            []AnswerChoice `json:"answers"`
	RightAnswerLabel   string         `json:"right_answer_label,omitempty"`
	RightAnswerContent string         `json:"right_answer_content,omitempty"`
	QuestionTopic      string         `json:"question_topic,omitempty"`
	QuestionGrade      *int           `json:"question_grade,omitempty"`
	QuestionLevel      *int           `json:"question_level,omitempty"`

	// ImageKey is the S3 key of an image shown with the stem (question bank
	// sets only).
	ImageKey string `json:"question_image_key,omitempty"`

	Kind           string   `json:"question_kind,omitempty"`
	RightOptionIDs []string `json:"right_option_ids,omitempty"`
	RightValues    []string `json:"right_values,omitempty"`
}

// Answer kinds. They carry the same strings as the question bank's kinds
// (domain/questionpool) so a bank question converts without a lookup, but
// this package depends on nothing in domain/.
//
//   - SINGLE_CHOICE   — Answers, one correct option id (RightAnswerLabel).
//   - MULTIPLE_CHOICE — Answers, one or more correct option ids (RightOptionIDs).
//   - TRUE_FALSE      — two Answers (ids "true" and "false"), one correct id;
//     graded exactly like SINGLE_CHOICE, the kind only tells the client how to draw it.
//   - NUMERIC_INPUT   — no Answers; the child types a number, right when it
//     equals one of RightValues as a number.
const (
	KindSingleChoice   = "SINGLE_CHOICE"
	KindMultipleChoice = "MULTIPLE_CHOICE"
	KindTrueFalse      = "TRUE_FALSE"
	KindNumericInput   = "NUMERIC_INPUT"
)

// EffectiveKind is the question's kind, SINGLE_CHOICE when none is stored.
func (q Question) EffectiveKind() string {
	if q.Kind == "" {
		return KindSingleChoice
	}
	return q.Kind
}

// TrueOptionID and FalseOptionID are the option ids of a TRUE_FALSE question.
const (
	TrueOptionID  = "true"
	FalseOptionID = "false"
)

// KeyOptionIDs is the answer key as one string, the form the answer log and
// ma_questions store: the correct option id, the correct ids joined for a
// multiple-choice question, "" for a numeric one.
func (q Question) KeyOptionIDs() string {
	switch q.EffectiveKind() {
	case KindMultipleChoice:
		return JoinIDs(q.RightOptionIDs)
	case KindNumericInput:
		return ""
	}
	return q.RightAnswerLabel
}

// KeyText is the displayed text of the answer key: the correct option's
// text, the correct options' texts joined, or the accepted numbers joined.
func (q Question) KeyText() string {
	switch q.EffectiveKind() {
	case KindMultipleChoice:
		if q.RightAnswerContent != "" {
			return q.RightAnswerContent
		}
		var texts []string
		for _, c := range q.Answers {
			if slices.Contains(q.RightOptionIDs, c.Label) {
				texts = append(texts, c.Content)
			}
		}
		return strings.Join(texts, ContentSeparator)
	case KindNumericInput:
		return strings.Join(q.RightValues, ContentSeparator)
	}
	return q.RightAnswerContent
}

// ContentSeparator joins several option texts, or several accepted numbers,
// into the one string a log column holds.
const ContentSeparator = " | "

// JoinIDs is the stored form of a set of option ids: sorted, comma-joined
// ("a,c"; an id never holds a comma). The log keeps one string per answer
// whatever the kind.
func JoinIDs(ids []string) string {
	sorted := slices.Clone(ids)
	slices.Sort(sorted)
	return strings.Join(sorted, ",")
}

// SplitIDs undoes JoinIDs; "" is no ids.
func SplitIDs(joined string) []string {
	if joined == "" {
		return nil
	}
	return strings.Split(joined, ",")
}

// StudentAnswer is the student's chosen label for a single question.
// Classroom exercises use it; an exam answer is ExamAnswer.
type StudentAnswer struct {
	QuestionNumber int    `json:"question_number"`
	Label          string `json:"label"`
}

// ExamAnswer is what a child sent for one exam question: OptionIDs for a
// choice question (one id, or several for MULTIPLE_CHOICE), Value for a
// NUMERIC_INPUT one. Exactly one of the two is set; which one a question
// takes follows its kind. The ids are the ones the question was served
// with: they do not change when the options are shuffled.
type ExamAnswer struct {
	QuestionNumber int      `json:"question_number"`
	OptionIDs      []string `json:"option_ids,omitempty"`
	Value          string   `json:"value,omitempty"`
}

// GradingResult is a graded round's outcome. AssessmentGrade is only
// populated where the flow predicts a grade; practice rounds omit it.
type GradingResult struct {
	TotalQuestions  int     `json:"total_questions"`
	CorrectNumber   int     `json:"correct_number"`
	ScorePercentage int     `json:"score_percentage"`
	Review          string  `json:"review"`
	AssessmentGrade *string `json:"assessment_grade,omitempty"`
}
