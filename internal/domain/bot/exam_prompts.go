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
// GRADE is the axis that decides the content (number range, operations,
// icon policy). The system prompt carries it as the CASE KG / CASE NUM
// switch plus the 11 kindergarten question types, and the user message
// binds {grade} to the band label that switch keys on. LEVEL refines
// intensity inside that grade and is rendered for a GRADE review only —
// see level_profile.go.
//
// The system prompt is a rule sheet written once per prompt language
// (exam_templates_vn.go, exam_templates_en.go) in the same format; the
// user message carries only what varies per request.
//
// The LANGUAGE OF THE PROMPT is a cost decision — English instructions
// tokenise shorter than Vietnamese ones — and is separate from the
// language of the round: the questions,
// topics and short_text the model writes are Vietnamese in both, because
// the product serves Vietnamese children and ma_exam_pools cannot record a
// row as being anything else.

// ExamPromptInput is everything the generation prompt consumes. It mirrors
// the req_* columns on ma_exam_pools, so what shaped a prompt can always be
// read back off the stored row.
type ExamPromptInput struct {
	// Language is the language the INSTRUCTIONS are written in. Empty
	// means Vietnamese. It does not change the language of the generated
	// round, which is always Vietnamese — see the package note.
	Language QuizLanguage
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

// examAllowedEmoji is the only icon set a generated round may use. It is
// ONE list for both prompt languages: the two templates are written
// separately, and a list copied into each is a list that drifts.
const examAllowedEmoji = "🍎 🍊 🍐 🍌 🍉 🍇 🍓 🍒 🍑 🍍 🥝 🥕 🌽 🍅 🥦 🥒 🍭 🍬 🍪 🍩 🎂 🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨 🐯 🦁 🐮 🐷 🐸 🐵 🐔 🐧 🐦 🐤 🦆 🦉 🐟 🐠 🐡 🦋 🐝 🐞 🐢 🚗 🚕 🚌 🚎 🚲 🛵 🚂 ✈️ 🚁 🚢 ⭐️ 🎈 ⚽️ 🧸 📚 ✏️ 🖍️ 🎁 🔴 🟡 🟢 🔵 🟠 🟣 🟥 🟨 🟩 🟦 🟧 🟪"

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

// compactPositions renders 1-based positions the way the teaching team
// writes them — "Q1,2,4,5" — one Q prefix, the rest bare.
func compactPositions(positions []int) string {
	if len(positions) == 0 {
		return ""
	}
	parts := make([]string, 0, len(positions))
	for _, p := range positions {
		parts = append(parts, fmt.Sprint(p))
	}
	return "Q" + strings.Join(parts, ",")
}

// probeAndRestLists renders the two lists every CASE block names: the
// probe positions, and every other question in a round of n. The
// complement is computed here, once for both prompt languages, so a
// prompt can never name a question twice or leave one unassigned.
func probeAndRestLists(probes []int, n int) (probeList, restList string) {
	isProbe := make(map[int]bool, len(probes))
	for _, p := range probes {
		isProbe[p] = true
	}
	var rest []int
	for q := 1; q <= n; q++ {
		if !isProbe[q] {
			rest = append(rest, q)
		}
	}
	return compactPositions(probes), compactPositions(rest)
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
	lang := QuizLanguageVietnamese
	if in.Language != "" {
		if lang, err = normalizeLanguage(in.Language); err != nil {
			return "", "", err
		}
	}
	n := in.NumQuestions
	if n <= 0 {
		n = examDefaultNumQuestions
	}
	if lang == QuizLanguageEnglish {
		return buildSystemExamEN(in, n), buildUserExamEN(in, n), nil
	}
	return buildSystemExamVN(in, n), buildUserExamVN(in, n), nil
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
