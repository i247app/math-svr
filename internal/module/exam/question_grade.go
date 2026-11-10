package exam

import (
	"math-ai.com/math-ai/internal/application/dto/aiquestion"
	"math-ai.com/math-ai/internal/domain/prompt/input"
	"math-ai.com/math-ai/internal/shared/enum"
)

// GradeMismatch records that the model's own question_grade disagreed with
// the one the server stamped. It is never an error — the server's value
// wins unconditionally — but a run of these means the prompt is not
// landing, which is worth seeing in the logs before it shows up as
// mis-placed children.
type GradeMismatch struct {
	QuestionNumber int
	ModelGrade     *int
	Applied        int
}

// NormalizeQuestionBands stamps question_grade on every question from the
// request, by position, and reports where the model disagreed.
//
// The server decides this, not the model, for the same reason the grade
// profile is code-owned: question_grade is an input to placement — it is
// how "did the child handle material one band up?" gets answered — so a
// model that mislabels a question would quietly move a child up or down a
// grade. The prompt still asks for the field, and the mismatches returned
// here are how we find out whether it is being obeyed.
//
// It runs BEFORE the questions are persisted, so ma_exam_pools.questions_json
// is authoritative and ma_exam_session_lines can copy it verbatim at submit
// time without re-deriving anything.
func NormalizeQuestionBands(questions []aiquestion.Question, examType enum.ExamType, grade int) ([]aiquestion.Question, []GradeMismatch) {
	if len(questions) == 0 {
		return questions, nil
	}

	probeAt := make(map[int]struct{})
	for _, pos := range input.ProbePositions(examType, len(questions)) {
		probeAt[pos] = struct{}{}
	}
	probeGrade := input.ProbeGrade(grade)

	var mismatches []GradeMismatch
	for i := range questions {
		q := &questions[i]

		// Trust question_number when the model numbered the item sanely;
		// fall back to the slice position when it did not, so a payload
		// with missing or duplicated numbers still gets every question
		// stamped instead of silently defaulting them all to the base grade.
		number := q.QuestionNumber
		if number <= 0 || number > len(questions) {
			number = i + 1
		}

		applied := grade
		if _, isProbe := probeAt[number]; isProbe {
			applied = probeGrade
		}

		if q.QuestionGrade != nil && *q.QuestionGrade != applied {
			mismatches = append(mismatches, GradeMismatch{
				QuestionNumber: number,
				ModelGrade:     q.QuestionGrade,
				Applied:        applied,
			})
		}

		stamped := applied
		q.QuestionGrade = &stamped
	}
	return questions, mismatches
}

// LevelMismatch records that the model's question_level or
// question_sub_level disagreed with the stamped one; logged, never fatal.
type LevelMismatch struct {
	QuestionNumber                int
	ModelLevel, ModelSubLevel     *int
	AppliedLevel, AppliedSubLevel int
}

// StampQuestionLevels writes question_level and question_sub_level on every
// question of a model-written round, by position, from the pairs drawn for
// the round (input.QuestionLevels — the very values the user message asked
// for), and reports where the model disagreed. Server-stamped for the
// same reason question_grade is: the columns feed analytics, and a value
// the model invented would be a difficulty nobody asked for.
func StampQuestionLevels(questions []aiquestion.Question, levels, subLevels []int) ([]aiquestion.Question, []LevelMismatch) {
	var mismatches []LevelMismatch
	for i := range questions {
		q := &questions[i]
		// Same position rule as NormalizeQuestionBands.
		number := q.QuestionNumber
		if number <= 0 || number > len(questions) {
			number = i + 1
		}
		// A model that wrote more questions than asked gets the last pair.
		idx := min(number, len(levels)) - 1
		if idx < 0 {
			continue
		}
		l, s := levels[idx], subLevels[idx]
		if (q.QuestionLevel != nil && *q.QuestionLevel != l) || (q.QuestionSubLevel != nil && *q.QuestionSubLevel != s) {
			mismatches = append(mismatches, LevelMismatch{
				QuestionNumber: number,
				ModelLevel:     q.QuestionLevel, ModelSubLevel: q.QuestionSubLevel,
				AppliedLevel: l, AppliedSubLevel: s,
			})
		}
		q.QuestionLevel, q.QuestionSubLevel = &l, &s
	}
	return questions, mismatches
}
