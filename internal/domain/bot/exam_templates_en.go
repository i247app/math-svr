package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// English exam-generation templates: the same rule sheet as
// exam_templates_vn.go, written in English because English instructions
// tokenise shorter, and a generation call pays for its prompt every time.
//
// Only the INSTRUCTIONS are English. The round itself — question text,
// answer content, question_topic, short_text — is Vietnamese in both
// templates, and this one says so explicitly, since a model reading an
// English prompt will otherwise answer in English.
//
// The user message here is only the grade. The GRADE PROFILE the rules
// refer to, the LEVEL PROFILE and the practice brief are not sent in this
// language — that is the shape the teaching team asked to trial, and the
// Vietnamese template still renders all of them for comparison.

// systemExamENTmpl is the whole system prompt, filled by strings.Replacer:
// {{N}} is the question count; {{PROBE_RULE}}, {{KG_GRADE}} and
// {{NUM_GRADE}} are the three lines that name the probe positions
// (examProbeLinesEN), so the prompt and the server-side re-stamp always
// agree on which questions reach up a grade.
//
// The grade is addressed by its Vietnamese label throughout — the CASE
// switch keys on "Mẫu giáo" / "Lớp N" and question_grade is asked for as
// that label — so the user message binds {grade} to the same label. The
// parser maps it back to the stored int.
const systemExamENTmpl = `ROLE: Author Vietnamese math quizzes for Kindergarten–Grade 5.
Generate EXACTLY {{N}} multiple-choice questions for {grade}.
GENERAL:
* Difficulty increases Q1→Q{{N}}.
* No same type in adjacent questions; each type max 2 times.
* {{PROBE_RULE}}
* Exactly 4 answers A–D; exactly 1 correct answer; plausible distractors.
* Every question must be solvable from its content alone.
* All child-facing content is Vietnamese.
* JSON keys are English.

CASE KG — if {grade} = "Mẫu giáo":
* {{KG_GRADE}}
* Start from a KG type, then generate the question.
* Icons allowed; no same icon category 3 questions in a row.
* Use short Vietnamese text only when symbols/icons are insufficient.
* "|" separates icon groups; max 10 icons/group.

11 KG QUESTION TYPES:
Format: TYPE | question pattern | example question_name | correct answer
1. COUNTING | ICON×n = ? | 🍎🍎🍎 = ? | 3
2. NUM_MATCH | N = ? (corresponding icons) | 3 = ? | 🥝🥝🥝
3. ADD | G + G = ? | 🦉🦉 + 🦉 = ? | 3
4. SUB | G − G = ? | 🐝🐝🐝 − 🐝 = ? | 2
5. COMP | G = G + ? | 🚁🚁🚁 = 🚁🚁 + ? | 1
6. CMP | G ? G → >, <, = | 🎁🎁🎁 ? ⚽️⚽️ | >
7. PATTERN | ICON×4 ? | 🔴🟡🔴🟡🔴 ? | 🟡
8. ODD_ONE | select the different item | Chọn khác loại:  🍒 | 🍑 | 🍍| 🦋
9. NUM_SEQ | N N ? N | 1 2 ? 4 | 3
10. BEFORE_AFTER | N ? N | 2 ? 4 | 3
11.ORDER_NUM | N N N → sắp xếp tăng/giảm dần | Sắp xếp tăng dần: 321 | 123
CASE NUM — if {grade} = Lớp 1–5:
* NO emoji.
* Select exactly ONE textbook: Chân Trời Sáng Tạo / Kết Nối Tri Thức / Cánh Diều.
* State textbook in short_text.
* Follow that textbook's grade-level topic scope.
* Select question_topic FIRST, then derive a suitable question_type.
* Do NOT start from question type.
* {{NUM_GRADE}}
* Keep questions clear, age-appropriate, and progressively harder.


OUTPUT:
Return ONLY valid JSON. Do not return Markdown or explanations. All JSON keys must be in English.
KG: question_name and answers[].content may contain only numbers, mathematical symbols, allowed emojis, and visual arrangements.
Grades 1–5: all child-facing content must be in Vietnamese.
JSON keys must always remain in English.
STRUCTURE:
{
  "title": "...",
  "short_text": "...",
  "questions": [
    {
      "question_number": 1,
      "question_type": "...",
      "question_name": "...",
      "answers": [
        {
          "label": "A",
          "content": "..."
        }
      ],
      "right_answer_label": "...",
      "right_answer_content": "...",
      "question_topic": "...",
      "question_grade": "..."
    }
  ]
}
IMPORTANT LANGUAGE RULE:
Instructions are in English, but all generated quiz content must be in Vietnamese. JSON keys must be in English.`

// examProbeLinesEN renders the three lines that name the probe positions:
// the GENERAL rule, the KG question_grade line and the NUM one. The
// positions come from ProbePositions so the prompt asks for exactly what
// the server enforces afterwards; the "all other questions" list is the
// complement of the probes within the round.
func examProbeLinesEN(probes []int, n int) (rule, kg, num string) {
	if len(probes) == 0 {
		return "All questions = current-grade content.",
			"All questions = Mẫu giáo.",
			"All questions = current grade."
	}
	isProbe := map[int]bool{}
	for _, p := range probes {
		isProbe[p] = true
	}
	var rest []int
	for q := 1; q <= n; q++ {
		if !isProbe[q] {
			rest = append(rest, q)
		}
	}
	probeList := compactPositions(probes)
	restList := compactPositions(rest)
	rule = fmt.Sprintf("%s = next-grade content. Grade %d → early Grade %d.",
		joinPositions(probes, ", "), enum.ExamGradeMax, GradeProbeCeiling)
	kg = fmt.Sprintf("%s = Mẫu giáo; %s = Lớp 1.", restList, probeList)
	num = fmt.Sprintf("%s = current grade; %s = next grade.", restList, probeList)
	return rule, kg, num
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

// joinPositions renders 1-based positions as Q3, Q6, … with the given
// separator.
func joinPositions(positions []int, sep string) string {
	labels := make([]string, 0, len(positions))
	for _, p := range positions {
		labels = append(labels, fmt.Sprintf("Q%d", p))
	}
	return strings.Join(labels, sep)
}

func buildSystemExamEN(in ExamPromptInput, n int) string {
	rule, kg, num := examProbeLinesEN(ProbePositions(in.ExamType, n), n)
	return strings.NewReplacer(
		"{{N}}", fmt.Sprint(n),
		"{{PROBE_RULE}}", rule,
		"{{KG_GRADE}}", kg,
		"{{NUM_GRADE}}", num,
	).Replace(systemExamENTmpl)
}

// buildUserExamEN binds {grade} to the Vietnamese band label the system
// prompt's MODE switch and question_grade vocabulary are written in, then
// lists the stems the child has recently met so the round does not
// repeat them.
func buildUserExamEN(in ExamPromptInput, _ int) string {
	out := "current_grade: " + ExamTitle(in.Grade)
	// if avoid := examAvoidBlock(QuizLanguageEnglish, in.Avoid); avoid != "" {
	// 	out += "\n\n" + avoid
	// }
	return out
}
