package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/application/command/shared/seqgen"
	examDto "math-ai.com/math-ai/internal/application/dto/exam"
	sharedQuestion "math-ai.com/math-ai/internal/application/dto/question"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/question"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SaveQuestionsFromPoolCommand copies every question of one ma_exam_pools
// set into the question bank, one ma_questions row each.
//
// The pool row is read inside the transaction and is the only input: its
// ai_questions_json gives the canonical questions (stored order, server-
// stamped question_grade) and its verified_count is copied as is. Taking
// the questions from the caller instead would let the bank drift from
// what the pool actually stores.
type SaveQuestionsFromPoolCommand struct {
	ExamID   int64
	ActorUID *int64 // create_id / modify_id; nil = no uid (API-key admin)
}

type SaveQuestionsFromPoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSaveQuestionsFromPoolCommandHandler(uow transaction.UnitOfWork) *SaveQuestionsFromPoolCommandHandler {
	return &SaveQuestionsFromPoolCommandHandler{uow: uow}
}

// Handle returns how many questions were written.
func (h *SaveQuestionsFromPoolCommandHandler) Handle(ctx context.Context, cmd SaveQuestionsFromPoolCommand) (int, error) {
	if cmd.ExamID <= 0 {
		return 0, errs.NewError(ctx, status.QUESTION_MISSING_EXAM_ID, nil,
			errors.New("question: exam_id is required"))
	}
	var saved int
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		pool, err := repos.ExamPool.FindByExamId(ctx, cmd.ExamID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if pool == nil {
			return errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
				fmt.Errorf("question: exam_pool %d not found", cmd.ExamID))
		}
		saved, err = SaveFromPool(ctx, repos, pool, cmd.ActorUID)
		return err
	})
	if err != nil {
		return 0, err
	}
	return saved, nil
}

// SaveFromPool is the command's body for a caller already inside a UoW
// that holds the pool row it just wrote (/exams/pools/verify), so the bank
// changes in the same transaction as the pool.
//
// Upsert on uk_exam_question (exam_id, question_number), done by hand: a
// number the bank already holds is UPDATEd in place and keeps its
// question_id; only new numbers mint an id and are INSERTed. Two copies of
// the same pool racing both miss the read and the second INSERT fails on
// uk_exam_question (1062) — an error, never a duplicate. A soft-deleted row
// is not read, so copying it again fails the same way instead of
// resurrecting it. The whole set is refused when any question is malformed.
func SaveFromPool(ctx context.Context, repos transaction.Repositories, pool *exam.ExamPool, actorUID *int64) (int, error) {
	questions, err := BankQuestionsFromPool(pool)
	if err != nil {
		return 0, errs.NewError(ctx, status.QUESTION_INVALID_SHAPE, nil, err)
	}
	if len(questions) == 0 {
		return 0, nil
	}

	existing, err := repos.Question.ListByExamId(ctx, pool.ExamId())
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	idByNumber := make(map[int]int64, len(existing))
	for _, q := range existing {
		idByNumber[q.QuestionNumber()] = q.QuestionId()
	}
	var created []*question.Question
	for _, q := range questions {
		if id, ok := idByNumber[q.QuestionNumber()]; ok {
			q.SetQuestionId(id)
			if err := repos.Question.UpdateContent(ctx, q, actorUID); err != nil {
				return 0, errs.NewError(ctx, status.FAIL, nil, err)
			}
			continue
		}
		id, err := seqgen.Next(ctx, repos.Seq, seq.NameQuestion)
		if err != nil {
			return 0, err
		}
		q.SetQuestionId(id)
		created = append(created, q)
	}
	if err := repos.Question.CreateMany(ctx, created, actorUID); err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return len(questions), nil
}

// BankQuestionsFromPool maps a pool's stored questions to bank rows,
// without ids. The type goes through the same question.Normalize as a
// generation (unknown -> ARITHMETIC); question_grade is the one the server
// stamped (the probe band on ASSESSMENT / PRACTICE probes), falling back
// to the pool's grade for a question that carries none.
func BankQuestionsFromPool(pool *exam.ExamPool) ([]*question.Question, error) {
	var stored []examDto.ExamQuestion
	if err := json.Unmarshal([]byte(pool.AiQuestionsJson()), &stored); err != nil {
		return nil, fmt.Errorf("exam_pool %d: decode questions: %w", pool.ExamId(), err)
	}
	shared, _ := sharedQuestion.Normalize(examDto.ToSharedQuestions(stored))

	out := make([]*question.Question, 0, len(shared))
	seen := make(map[int]bool, len(shared))
	for _, sq := range shared {
		if seen[sq.QuestionNumber] {
			return nil, fmt.Errorf("exam_pool %d: question_number %d appears twice", pool.ExamId(), sq.QuestionNumber)
		}
		seen[sq.QuestionNumber] = true
		q, err := bankQuestion(pool, sq)
		if err != nil {
			return nil, fmt.Errorf("exam_pool %d question %d: %w", pool.ExamId(), sq.QuestionNumber, err)
		}
		out = append(out, q)
	}
	return out, nil
}

func bankQuestion(pool *exam.ExamPool, sq sharedQuestion.Question) (*question.Question, error) {
	if sq.QuestionNumber <= 0 {
		return nil, errors.New("question_number must be positive")
	}
	name := strings.TrimSpace(sq.QuestionName)
	if name == "" {
		return nil, errors.New("question_name is empty")
	}
	answers, err := bankAnswers(sq)
	if err != nil {
		return nil, err
	}
	grade := pool.ReqGrade()
	if sq.QuestionGrade != nil {
		grade = *sq.QuestionGrade
	}
	if grade < 0 {
		return nil, fmt.Errorf("question_grade %d is negative", grade)
	}

	q := question.NewQuestion()
	q.SetExamId(pool.ExamId())
	q.SetQuestionNumber(sq.QuestionNumber)
	q.SetQuestionType(sq.QuestionType)
	q.SetQuestionName(name)
	q.SetAnswers(answers)
	if topic := strings.TrimSpace(sq.QuestionTopic); topic != "" {
		q.SetQuestionTopic(&topic)
	}
	q.SetQuestionGrade(grade)
	q.SetQuestionTags([]string{})
	q.SetVerifiedCount(pool.VerifiedCount())
	return q, nil
}

// bankAnswers drops the labels and keeps the stored (display) order; the
// option named by right_answer_label is the one is_correct answer.
func bankAnswers(sq sharedQuestion.Question) ([]question.Answer, error) {
	if len(sq.Answers) < 2 {
		return nil, fmt.Errorf("%d answers, need at least 2", len(sq.Answers))
	}
	labels := make(map[string]bool, len(sq.Answers))
	answers := make([]question.Answer, 0, len(sq.Answers))
	correct := 0
	for _, a := range sq.Answers {
		label := strings.TrimSpace(a.Label)
		if label == "" || labels[label] {
			return nil, fmt.Errorf("answer label %q is blank or repeated", a.Label)
		}
		labels[label] = true
		isCorrect := label == strings.TrimSpace(sq.RightAnswerLabel)
		if isCorrect {
			correct++
		}
		answers = append(answers, question.Answer{Content: a.Content, IsCorrect: isCorrect})
	}
	if correct != 1 {
		return nil, fmt.Errorf("right_answer_label %q names no answer", sq.RightAnswerLabel)
	}
	return answers, nil
}
