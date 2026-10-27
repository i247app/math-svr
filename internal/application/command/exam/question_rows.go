package command

import (
	"context"
	"encoding/json"
	"fmt"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/question"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// QuestionRows converts the questions JSON an exam pool stores
// (ai_questions_json, []dto.ExamQuestion) into its ma_questions rows, one
// per element, in stored order. It reads the very string the pool row
// holds, so every row matches its JSON element field for field; an empty
// right_answer_content or question_topic is stored as NULL. Ids are left
// to the caller.
func QuestionRows(examID int64, questionsJSON string) ([]*question.Question, error) {
	var stored []dto.ExamQuestion
	if err := json.Unmarshal([]byte(questionsJSON), &stored); err != nil {
		return nil, fmt.Errorf("exam: decode exam_pool %d questions: %w", examID, err)
	}
	active := string(enum.StatusActive)
	rows := make([]*question.Question, 0, len(stored))
	for _, s := range stored {
		answers := make([]question.Answer, 0, len(s.Answers))
		for _, a := range s.Answers {
			answers = append(answers, question.Answer{Label: a.Label, Content: a.Content})
		}
		q := question.NewQuestion()
		q.SetExamId(examID)
		q.SetQuestionNumber(s.QuestionNumber)
		q.SetQuestionType(s.QuestionType)
		q.SetQuestionName(s.QuestionName)
		q.SetAnswers(answers)
		q.SetRightAnswerLabel(s.RightAnswerLabel)
		q.SetRightAnswerContent(nonEmpty(s.RightAnswerContent))
		q.SetQuestionTopic(nonEmpty(s.QuestionTopic))
		q.SetQuestionGrade(s.QuestionGrade)
		q.SetQuestionStatus(&active)
		rows = append(rows, q)
	}
	return rows, nil
}

// mintQuestionRows is QuestionRows with a question_id minted for every
// row, inside the caller's transaction.
func mintQuestionRows(ctx context.Context, repos transaction.Repositories, examID int64, questionsJSON string) ([]*question.Question, error) {
	rows, err := QuestionRows(examID, questionsJSON)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	for _, q := range rows {
		id, err := seqgen.Next(ctx, repos.Seq, seq.NameQuestion)
		if err != nil {
			return nil, err
		}
		q.SetQuestionId(id)
	}
	return rows, nil
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
