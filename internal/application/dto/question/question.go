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
package question

// AnswerChoice is one option in a multiple-choice question. Label is the
// alphabetical key ("A".."D") the student selects; Content is the text
// that gets displayed.
type AnswerChoice struct {
	Label   string `json:"label"`
	Content string `json:"content"`
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
const (
	TypeArithmetic    = "ARITHMETIC"
	TypeCount         = "COUNT"
	TypePickByIcon    = "PICK_BY_ICON"
	TypeIdentifyShape = "IDENTIFY_SHAPE"
)

// Question is one MCQ item as produced by the bot. RightAnswer is the
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
}

// StudentAnswer is the student's chosen label for a single question.
type StudentAnswer struct {
	QuestionNumber int    `json:"question_number"`
	Label          string `json:"label"`
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
