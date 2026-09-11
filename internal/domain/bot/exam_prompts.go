package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Exam generation prompt. There is exactly one kind — generation.
//
// Grading has no prompt at all any more: the server scores every
// submission deterministically, so the model is never asked to mark a
// child's work. That is why this file has no Grade/Reinforce counterpart
// to the quiz prompts it replaces.
//
// GRADE is the one axis that steers the output today: it decides the
// content (number range, operations, icon policy). Its block is rendered at
// the TOP of the user message so it outranks the schema example, which
// otherwise teaches its own difficulty by imitation.
//
// There is deliberately no LEVEL axis. The schema reserves nullable
// req_level / res_level / question_level columns, but the teaching team
// has not defined what a level IS or how it should move, and a difficulty
// scale the product cannot explain must not be handed to the model as if
// it could. The columns stay NULL until that rule exists; add the axis
// back here when it does, not before.

// ExamPromptInput is everything the generation prompt consumes. It mirrors
// the req_* columns on ma_ai_exams, so what shaped a prompt can always be
// read back off the stored row.
type ExamPromptInput struct {
	ExamType enum.ExamType
	// Grade is the content band, 0..5 (0 = mẫu giáo).
	Grade        int
	NumQuestions int
	Semester     string
	Program      string
}

// examDefaultNumQuestions matches the module validator's default. It is
// repeated here so the domain builder stays usable without going through
// the module layer.
const examDefaultNumQuestions = 10

// assessmentProbeSlots are the 1-based positions of the harder questions
// inside an ASSESSMENT: the child answers mostly at their own grade, and
// these two reach one band higher to find out whether they are ready to
// move up.
//
// The positions are fixed because the exam is fixed at ten questions
// today. A variable-length exam needs a real strategy (a proportion, or
// spread by position) rather than a longer literal — decide that when
// variable length actually ships.
var assessmentProbeSlots = []int{3, 6}

// AssessmentProbePositions returns the probe positions that fit inside an
// exam of numQuestions, or nil when the type does not probe. It is
// exported because the module layer normalises question_grade against
// exactly these positions after generation: the prompt asks, the server
// enforces, and both must read the same list.
func AssessmentProbePositions(examType enum.ExamType, numQuestions int) []int {
	if examType != enum.ExamTypeAssessment {
		return nil
	}
	if numQuestions <= 0 {
		numQuestions = examDefaultNumQuestions
	}
	var out []int
	for _, slot := range assessmentProbeSlots {
		if slot <= numQuestions {
			out = append(out, slot)
		}
	}
	return out
}

// ProbeGrade is the band a probe question targets: one above the child's,
// capped so an ASSESSMENT at Grade 5 reaches Grade 6 and stops there.
func ProbeGrade(grade int) int {
	probe := grade + 1
	if probe > int(GradeProbeCeiling) {
		return int(GradeProbeCeiling)
	}
	return probe
}

// BuildExamPrompt returns the (system, user) contents for one generation
// call. The caller sets JSONMode and a low temperature and forwards them
// through the bot adapter.
func BuildExamPrompt(in ExamPromptInput) (system string, user string, err error) {
	if in.Grade < enum.ExamGradeMin || in.Grade > enum.ExamGradeMax {
		return "", "", fmt.Errorf("bot: exam grade %d out of range [%d,%d]",
			in.Grade, enum.ExamGradeMin, enum.ExamGradeMax)
	}
	if !in.ExamType.IsValid() {
		return "", "", fmt.Errorf("bot: unsupported exam type %q", string(in.ExamType))
	}
	n := in.NumQuestions
	if n <= 0 {
		n = examDefaultNumQuestions
	}
	return buildSystemExamVN(n), buildUserExamVN(in, n), nil
}

// examBandTitle is the exact string the model must put in "title". It is
// fully determined by the grade, so handing the model the finished string
// removes the only reason it had to invent a difficulty label of its own.
func examBandTitle(grade int) string {
	band := gradeBandName(QuizLanguageVietnamese, GradeLevel(grade))
	if band == "" {
		band = fmt.Sprintf("Lớp %d", grade)
	}
	return band
}

// examContextVN renders only the curriculum lines that carry a value, so a
// request with no semester or program still produces a coherent brief.
func examContextVN(in ExamPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Học kỳ: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Chương trình học: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

// examProbeBlockVN spells out the ASSESSMENT probe rule, naming the exact
// positions and the exact band they must reach. The server re-stamps
// question_grade afterwards regardless, so this block is about getting the
// CONTENT right — a probe question that is merely labelled harder, without
// being harder, measures nothing.
func examProbeBlockVN(in ExamPromptInput, n int) string {
	positions := AssessmentProbePositions(in.ExamType, n)
	if len(positions) == 0 {
		return fmt.Sprintf("- Mọi câu đều ở đúng cấp lớp trên, và \"question_grade\" của mọi câu đều là %d.", in.Grade)
	}

	probe := ProbeGrade(in.Grade)
	labels := make([]string, 0, len(positions))
	for _, p := range positions {
		labels = append(labels, fmt.Sprintf("câu %d", p))
	}

	probeName := gradeBandName(QuizLanguageVietnamese, GradeLevel(probe))
	probeRange := gradeBandRange(QuizLanguageVietnamese, GradeLevel(probe))
	line := fmt.Sprintf(`- CÂU DÒ TRẦN (bắt buộc với bài ASSESSMENT): %s phải khó hơn ĐÚNG MỘT BẬC, tức ở mức %s — %s
- Hai câu đó có "question_grade" = %d; TẤT CẢ các câu còn lại có "question_grade" = %d.
- Không đánh dấu, không chú thích, không nói cho học sinh biết câu nào là câu khó hơn.`,
		strings.Join(labels, " và "), probeName, probeRange, probe, in.Grade)
	return line
}
