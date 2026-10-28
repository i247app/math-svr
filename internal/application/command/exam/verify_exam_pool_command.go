package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"math-ai.com/math-ai/internal/application/dto/aiquestion"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
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
// terms (aiquestion.Shuffle) and every answer log names canonical labels,
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
		// A set drawn from the question bank is corrected in the bank.
		if pool.ExamSource() == string(enum.ExamSourcePool) {
			return errs.NewError(ctx, status.EXAM_POOL_BANK_READONLY, nil,
				fmt.Errorf("exam: exam_pool %d was drawn from the question bank", cmd.ExamID))
		}
		var stored []dto.ExamQuestion
		if err := json.Unmarshal([]byte(pool.QuestionsJson()), &stored); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("exam: decode exam_pool %d questions: %w", cmd.ExamID, err))
		}
		if err := SameQuestionShape(stored, cmd.Questions); err != nil {
			return errs.NewError(ctx, status.EXAM_POOL_INVALID_QUESTIONS, nil, err)
		}
		// Only a set the model wrote reaches here (a bank set is refused
		// above) and the model writes no images: whatever image a correction
		// carries is dropped, so a typed-in key or URL is never stored.
		for i := range cmd.Questions {
			q := &cmd.Questions[i]
			q.QuestionImageKey, q.QuestionImageURL = "", ""
			q.RightAnswerText = "" // derived from the key, never stored
			q.Options = slices.Clone(q.Options)
			for j := range q.Options {
				q.Options[j].ImageKey, q.Options[j].ImageURL = "", ""
			}
		}
		raw, err := json.Marshal(cmd.Questions)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("exam: marshal questions: %w", err))
		}
		if err := repos.ExamPool.ReplaceQuestions(ctx, cmd.ExamID, string(raw), cmd.AdminUID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		// ma_questions follows the overwritten JSON in the same tx.
		rows, err := mintQuestionRows(ctx, repos, cmd.ExamID, string(raw))
		if err != nil {
			return err
		}
		if err := repos.Question.ReplaceByExamId(ctx, cmd.ExamID, rows, cmd.AdminUID); err != nil {
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
// question_numbers (any order), the same question_kind and option ids per
// question (any order), and a usable question — a question_text and a
// correct_answer that fits its kind and names its options (dto.Usable).
func SameQuestionShape(stored, fixed []dto.ExamQuestion) error {
	if len(fixed) != len(stored) {
		return fmt.Errorf("%d questions sent, the exam has %d", len(fixed), len(stored))
	}
	want := make(map[int]dto.ExamQuestion, len(stored))
	for _, q := range stored {
		want[q.QuestionNumber] = q
	}
	seen := make(map[int]bool, len(fixed))
	for _, q := range fixed {
		old, ok := want[q.QuestionNumber]
		if !ok || seen[q.QuestionNumber] {
			return fmt.Errorf("question_number %d is not in the exam or is sent twice", q.QuestionNumber)
		}
		seen[q.QuestionNumber] = true
		if kindOf(q) != kindOf(old) {
			return fmt.Errorf("question %d: question_kind %q, the exam has %q", q.QuestionNumber, q.QuestionKind, old.QuestionKind)
		}
		if got, labels := labelsOf(q), labelsOf(old); !slices.Equal(got, labels) {
			return fmt.Errorf("question %d: option ids %v, the exam has %v", q.QuestionNumber, got, labels)
		}
		if !q.Usable() {
			return fmt.Errorf("question %d: needs a question_text and a correct_answer that fits its kind and names its options", q.QuestionNumber)
		}
	}
	return nil
}

// kindOf is a question's kind, SINGLE_CHOICE when it states none.
func kindOf(q dto.ExamQuestion) string {
	if q.QuestionKind == "" {
		return aiquestion.KindSingleChoice
	}
	return q.QuestionKind
}

// labelsOf returns a question's option ids, sorted.
func labelsOf(q dto.ExamQuestion) []string {
	ids := make([]string, 0, len(q.Options))
	for _, o := range q.Options {
		ids = append(ids, o.ID)
	}
	slices.Sort(ids)
	return ids
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
