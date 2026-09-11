package exam

import (
	"math-ai.com/math-ai/internal/application/dto/question"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
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

// NormalizeQuestionBands stamps question_grade AND question_level on every
// question from the request, by position, and reports where the model
// disagreed about the grade.
//
// The server decides this, not the model, for the same reason the grade
// profile is code-owned: question_grade is an input to placement — it is
// how "did the child handle material one band up?" gets answered — so a
// model that mislabels a question would quietly move a child up or down a
// grade. The prompt still asks for the field, and the mismatches returned
// here are how we find out whether it is being obeyed.
//
// It runs BEFORE the questions are persisted, so ma_ai_exams.ai_questions_json
// is authoritative and ma_user_exam_details can copy it verbatim at submit
// time without re-deriving anything.
func NormalizeQuestionBands(questions []question.Question, examType enum.ExamType, grade int, level *int) ([]question.Question, []GradeMismatch) {
	if len(questions) == 0 {
		return questions, nil
	}

	probeAt := make(map[int]struct{})
	for _, pos := range domainBot.AssessmentProbePositions(examType, len(questions)) {
		probeAt[pos] = struct{}{}
	}
	probeGrade := domainBot.ProbeGrade(grade)

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

		if q.Grade != nil && *q.Grade != applied {
			mismatches = append(mismatches, GradeMismatch{
				QuestionNumber: number,
				ModelGrade:     q.Grade,
				Applied:        applied,
			})
		}

		stamped := applied
		q.Grade = &stamped

		// Level is uniform across the round — there is no per-question
		// level rule the way Q3/Q6 is a per-question grade rule — but it is
		// stamped per question anyway so ma_user_exam_details carries both
		// bands on every row and a placement query never has to join back
		// to the exam to learn how hard the question was.
		if level != nil {
			stampedLevel := *level
			q.Level = &stampedLevel
		}
	}
	return questions, mismatches
}
