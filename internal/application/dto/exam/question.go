package exam

import "math-ai.com/math-ai/internal/application/dto/question"

// ExamQuestion is the exam flow's wire shape for one MCQ item.
//
// It exists because the exam vocabulary is its own: the field names here
// match the columns they end up in (right_answer_label,
// right_answer_content, question_topic, question_grade, question_level),
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
	QuestionLevel      *int                    `json:"question_level,omitempty"`
}

// ToShared converts into the internal shape the scorer and normalisers
// operate on.
func (q ExamQuestion) ToShared() question.Question {
	return question.Question{
		QuestionNumber: q.QuestionNumber,
		QuestionType:   q.QuestionType,
		QuestionName:   q.QuestionName,
		Answers:        q.Answers,
		RightAnswer:    q.RightAnswerLabel,
		CorrectAnswer:  q.RightAnswerContent,
		Topic:          q.QuestionTopic,
		Grade:          q.QuestionGrade,
		Level:          q.QuestionLevel,
	}
}

// ExamQuestionFrom converts back out to the wire shape.
func ExamQuestionFrom(q question.Question) ExamQuestion {
	return ExamQuestion{
		QuestionNumber:     q.QuestionNumber,
		QuestionType:       q.QuestionType,
		QuestionName:       q.QuestionName,
		Answers:            q.Answers,
		RightAnswerLabel:   q.RightAnswer,
		RightAnswerContent: q.CorrectAnswer,
		QuestionTopic:      q.Topic,
		QuestionGrade:      q.Grade,
		QuestionLevel:      q.Level,
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
