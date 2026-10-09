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
// child's work.
//
// The SYSTEM prompt is data, not code: ma_exam_prompts holds one per grade,
// written by admins, sent verbatim. Everything in it — question count,
// probe questions, curriculum scope, output structure — is fixed text. This
// package builds the USER message beside it (exam_user_prompt.go), and owns
// the rules the server enforces on the reply whatever the prompt says:
// which questions are probes (ProbePositions), which band they reach
// (ProbeGrade), and the title (ExamTitle).
//
// The instructions are English, because English tokenises shorter and a
// generation pays for its prompt every time. The round the model writes —
// questions, topics, short_text — is Vietnamese: the product serves
// Vietnamese children and ma_exam_pools cannot record a row as anything
// else.

// ExamPromptInput is everything the generation prompt consumes. It mirrors
// the req_* columns on ma_exam_pools, so what shaped a prompt can always be
// read back off the stored row.
type ExamPromptInput struct {
	ExamType enum.ExamType
	// Grade is the content band, 0..5 (0 = mẫu giáo).
	Grade        int
	NumQuestions int
	Semester     string
	Program      string
	// Level is the client-stated intensity, 0..9, already clamped to the
	// grade's ceiling by the caller. Rendered for a GRADE review only;
	// nil, or any other type, means no LEVEL PROFILE block.
	Level *int
	// Practice aims a PRACTICE round at what the child just did. Required
	// when ExamType is PRACTICE, ignored otherwise.
	Practice *PracticeBrief
	// Avoid is the stems of questions this child has recently seen at
	// this grade. The model is told not to reuse or lightly reword them;
	// empty renders nothing.
	Avoid []string
}

// PracticeBrief is what the model is told about the sitting a PRACTICE
// round follows. It is built by the application layer from the child's
// answer log; the prompt renders it and nothing else reads it.
//
// Wrong is every question the child got wrong, as shown. WeakTopics are
// the topics those came from, most-wrong first, so the prompt can weight
// them. StrongTopics are the topics the child got entirely right — in
// ADVANCE mode the round moves past them.
type PracticeBrief struct {
	Mode         enum.PracticeMode
	Wrong        []PracticeItem
	WeakTopics   []string
	StrongTopics []string
	// Tally is the per-topic count behind the two lists — how many
	// answers the topic had and how many went wrong. The prompt does not
	// read it; the journey view does, to preview the drill.
	Tally map[string]TopicTally
}

// TopicTally is one topic's score inside a sitting.
type TopicTally struct {
	Wrong    int
	Answered int
}

// PracticeItem is one wrong answer: the stem as served, its topic, what
// was right, and what the child picked.
type PracticeItem struct {
	Stem        string
	Topic       string
	RightAnswer string
	ChildAnswer string
}

// examDefaultNumQuestions matches the module validator's default. It is
// repeated here so the domain builder stays usable without going through
// the module layer.
const examDefaultNumQuestions = 10

// probeSlots are the 1-based positions of the harder questions inside a
// round that probes (see enum.ExamType.HasProbes): the child answers
// mostly at their own grade, and these two reach one band higher to find
// out whether they are ready to move up.
//
// The positions are fixed because the exam is fixed at ten questions
// today. A variable-length exam needs a real strategy (a proportion, or
// spread by position) rather than a longer literal — decide that when
// variable length actually ships.
var probeSlots = []int{3, 6}

// ProbePositions returns the probe positions that fit inside an exam of
// numQuestions, or nil when the type does not probe. It is exported
// because the module layer normalises question_grade against exactly
// these positions after generation: the prompt asks, the server enforces,
// and both must read the same list.
func ProbePositions(examType enum.ExamType, numQuestions int) []int {
	if !examType.HasProbes() {
		return nil
	}
	if numQuestions <= 0 {
		numQuestions = examDefaultNumQuestions
	}
	var out []int
	for _, slot := range probeSlots {
		if slot <= numQuestions {
			out = append(out, slot)
		}
	}
	return out
}

// joinPositions renders 1-based positions as Q3, Q6, … with the given
// separator.
func joinPositions(positions []int, sep string) string {
	labels := make([]string, 0, len(positions))
	for _, p := range positions {
		labels = append(labels, fmt.Sprintf("Q%d", p))
	}
	return strings.Join(labels, sep)
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

// BuildExamUserPrompt returns the user message of one generation call: the
// per-request brief that goes beside the grade's stored system prompt.
func BuildExamUserPrompt(in ExamPromptInput) (string, error) {
	if in.Grade < enum.ExamGradeMin || in.Grade > enum.ExamGradeMax {
		return "", fmt.Errorf("bot: exam grade %d out of range [%d,%d]",
			in.Grade, enum.ExamGradeMin, enum.ExamGradeMax)
	}
	if !in.ExamType.IsValid() {
		return "", fmt.Errorf("bot: unsupported exam type %q", string(in.ExamType))
	}
	if in.ExamType == enum.ExamTypePractice && in.Practice == nil {
		return "", fmt.Errorf("bot: a PRACTICE prompt needs a practice brief")
	}
	n := in.NumQuestions
	if n <= 0 {
		n = examDefaultNumQuestions
	}
	return buildUserExamEN(in, n), nil
}

// examTypeTitleVN is how each exam type reads in a title.
var examTypeTitleVN = map[enum.ExamType]string{
	enum.ExamTypeAssessment: "Đánh Giá",
	enum.ExamTypeGrade:      "Theo Lớp",
	enum.ExamTypePractice:   "Luyện Tập",
}

// ExamTitle is the stored ai_title of a round: "Toán <type> - <band>",
// e.g. "Toán Đánh Giá - Lớp 2". It is a pure function of the type and the
// grade, so the server stamps it and the model is not asked for one.
func ExamTitle(examType enum.ExamType, grade int) string {
	if label, ok := examTypeTitleVN[examType]; ok {
		return "Toán " + label + " - " + GradeLabel(grade)
	}
	return "Toán - " + GradeLabel(grade)
}

// GradeLabel is the Vietnamese band name of a grade ("Mẫu giáo",
// "Lớp 1", …), used to name the grade in prompts and in the exam title.
func GradeLabel(grade int) string {
	band := gradeBandName(QuizLanguageVietnamese, GradeLevel(grade))
	if band == "" {
		band = fmt.Sprintf("Grade %d", grade)
	}
	return band
}
