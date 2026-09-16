package bot

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Exam generation prompt. There is exactly one kind — generation.
//
// Grading has no prompt at all any more: the server scores every
// submission deterministically, so the model is never asked to mark a
// child's work. That is why this file has no Grade/Reinforce counterpart
// to the quiz prompts it replaces.
//
// GRADE is the axis that decides the content (number range, operations,
// icon policy). Its block opens the user message so it outranks the schema
// example in the system prompt, which otherwise teaches its own difficulty
// by imitation. LEVEL refines intensity inside that grade and is rendered
// for a GRADE review only — see level_profile.go.
//
// The system prompt is the terse rule sheet in exam_templates_vn.go; the
// user message carries only what varies per request.

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
	// Level is the client-stated intensity, 1..10, already clamped to the
	// grade's ceiling by the caller. Rendered for a GRADE review only;
	// nil, or any other type, means no LEVEL PROFILE block.
	Level *int
	// Practice aims a PRACTICE round at what the child just did. Required
	// when ExamType is PRACTICE, ignored otherwise.
	Practice *PracticeBrief
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
	if in.ExamType == enum.ExamTypePractice && in.Practice == nil {
		return "", "", fmt.Errorf("bot: a PRACTICE prompt needs a practice brief")
	}
	n := in.NumQuestions
	if n <= 0 {
		n = examDefaultNumQuestions
	}
	return buildSystemExamVN(in, n), buildUserExamVN_V2(in, n), nil
}

// ExamTitle is the stored ai_title for a round at this grade: the band
// name, and nothing else. The schema shows the model a title slot, but the
// server overwrites it with this — the title was always a pure function of
// the grade, and stamping it is cheaper and steadier than a prompt rule
// telling the model not to decorate it.
func ExamTitle(grade int) string {
	band := gradeBandName(QuizLanguageVietnamese, GradeLevel(grade))
	if band == "" {
		band = fmt.Sprintf("Lớp %d", grade)
	}
	return band
}
