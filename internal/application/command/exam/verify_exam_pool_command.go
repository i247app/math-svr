package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// MarkExamPoolVerifyCommand flags a question set as verified (0 -> 1; a
// set already verified keeps its count) or not verified (back to 0).
type MarkExamPoolVerifyCommand struct {
	ExamID   int64
	IsVerify bool
	AdminUID *int64 // nil = no uid (API-key admin)
}

type MarkExamPoolVerifyCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewMarkExamPoolVerifyCommandHandler(uow transaction.UnitOfWork) *MarkExamPoolVerifyCommandHandler {
	return &MarkExamPoolVerifyCommandHandler{uow: uow}
}

func (h *MarkExamPoolVerifyCommandHandler) Handle(ctx context.Context, cmd MarkExamPoolVerifyCommand) (*exam.ExamPool, error) {
	var saved *exam.ExamPool
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if _, err := findLivePool(ctx, repos, cmd.ExamID); err != nil {
			return err
		}
		if err := repos.ExamPool.MarkVerified(ctx, cmd.ExamID, cmd.IsVerify, cmd.AdminUID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		var err error
		saved, err = findLivePool(ctx, repos, cmd.ExamID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// VerifyExamPoolCommand overwrites a question set with the version an
// admin corrected, and counts one more verification.
//
// The corrected set must keep the stored set's SHAPE: the same
// question_numbers, and for each question the same answer labels. Every
// sitting already handed out stores its ordering in those canonical
// terms (question.Shuffle) and every answer log names canonical labels,
// so a renumbered question or a renamed option would break the papers in
// flight and the history. Everything else — wording, option contents,
// which label is right, topic, grade — may change.
type VerifyExamPoolCommand struct {
	ExamID    int64
	Questions []dto.ExamQuestion
	AdminUID  *int64 // nil = no uid (API-key admin)
}

type VerifyExamPoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewVerifyExamPoolCommandHandler(uow transaction.UnitOfWork) *VerifyExamPoolCommandHandler {
	return &VerifyExamPoolCommandHandler{uow: uow}
}

func (h *VerifyExamPoolCommandHandler) Handle(ctx context.Context, cmd VerifyExamPoolCommand) (*exam.ExamPool, error) {
	var saved *exam.ExamPool
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		pool, err := findLivePool(ctx, repos, cmd.ExamID)
		if err != nil {
			return err
		}
		var stored []dto.ExamQuestion
		if err := json.Unmarshal([]byte(pool.AiQuestionsJson()), &stored); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("exam: decode exam_pool %d questions: %w", cmd.ExamID, err))
		}
		if err := SameQuestionShape(stored, cmd.Questions); err != nil {
			return errs.NewError(ctx, status.EXAM_POOL_INVALID_QUESTIONS, nil, err)
		}
		raw, err := json.Marshal(cmd.Questions)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("exam: marshal questions: %w", err))
		}
		if err := repos.ExamPool.ReplaceQuestions(ctx, cmd.ExamID, string(raw), cmd.AdminUID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		saved, err = findLivePool(ctx, repos, cmd.ExamID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// SameQuestionShape reports why fixed cannot replace stored, or nil: same
// question_numbers (any order), the same answer labels per question (any
// order), and a right_answer_label that names one of them.
func SameQuestionShape(stored, fixed []dto.ExamQuestion) error {
	if len(fixed) != len(stored) {
		return fmt.Errorf("%d questions sent, the exam has %d", len(fixed), len(stored))
	}
	want := make(map[int][]string, len(stored))
	for _, q := range stored {
		want[q.QuestionNumber] = labelsOf(q)
	}
	seen := make(map[int]bool, len(fixed))
	for _, q := range fixed {
		labels, ok := want[q.QuestionNumber]
		if !ok || seen[q.QuestionNumber] {
			return fmt.Errorf("question_number %d is not in the exam or is sent twice", q.QuestionNumber)
		}
		seen[q.QuestionNumber] = true
		got := labelsOf(q)
		if !slices.Equal(got, labels) {
			return fmt.Errorf("question %d: answer labels %v, the exam has %v", q.QuestionNumber, got, labels)
		}
		if !slices.Contains(got, q.RightAnswerLabel) {
			return fmt.Errorf("question %d: right_answer_label %q is not one of its answers", q.QuestionNumber, q.RightAnswerLabel)
		}
		if q.QuestionName == "" {
			return fmt.Errorf("question %d: question_name is empty", q.QuestionNumber)
		}
	}
	return nil
}

// labelsOf returns a question's answer labels, sorted.
func labelsOf(q dto.ExamQuestion) []string {
	labels := make([]string, 0, len(q.Answers))
	for _, a := range q.Answers {
		labels = append(labels, a.Label)
	}
	slices.Sort(labels)
	return labels
}

func findLivePool(ctx context.Context, repos transaction.Repositories, examID int64) (*exam.ExamPool, error) {
	pool, err := repos.ExamPool.FindByExamId(ctx, examID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if pool == nil {
		return nil, errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
			errors.New("exam: exam_pool not found"))
	}
	return pool, nil
}
