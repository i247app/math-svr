package command

import (
	"context"
	"encoding/json"
	"fmt"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/question"
	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// QuestionRows converts the questions JSON an exam pool stores
// (questions_json, []dto.ExamQuestion) into its ma_questions rows, one
// per element, in stored order. It reads the very string the pool row
// holds, so every row matches its JSON element field for field (the key
// is its correct_answer, in the bank's shape); an empty question_topic is
// stored as NULL. Ids are left to the caller.
func QuestionRows(examID int64, questionsJSON string) ([]*question.Question, error) {
	var stored []dto.ExamQuestion
	if err := json.Unmarshal([]byte(questionsJSON), &stored); err != nil {
		return nil, fmt.Errorf("exam: decode exam_pool %d questions: %w", examID, err)
	}
	active := string(enum.StatusActive)
	rows := make([]*question.Question, 0, len(stored))
	for _, s := range stored {
		options := make([]question.Option, 0, len(s.Options))
		for _, o := range s.Options {
			options = append(options, question.Option{ID: o.ID, Text: o.Text})
		}
		var key questionpool.CorrectAnswer
		if ca := s.CorrectAnswer; ca != nil {
			key = questionpool.CorrectAnswer{OptionIds: ca.OptionIDs, Value: ca.Value, Values: ca.Values}
		}
		q := question.NewQuestion()
		q.SetExamId(examID)
		q.SetQuestionNumber(s.QuestionNumber)
		q.SetRenderType(s.RenderType)
		q.SetQuestionKind(nonEmpty(s.QuestionKind))
		q.SetQuestionText(s.QuestionText)
		q.SetOptions(options)
		q.SetCorrectAnswer(key)
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
