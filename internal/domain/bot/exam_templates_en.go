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
// The grade is addressed by its Vietnamese label throughout — the MODE
// switch keys on "Mẫu giáo" / "Lớp N" and question_grade is asked for as
// that label — so the user message binds {grade} to the same label. The
// parser maps it back to the stored int.
//
// COVERAGE is written for a ten-question round and is left as the
// teaching team wrote it; a different length keeps the count and probe
// lines right but not that schedule.
const systemExamENTmpl = `ROLE: Author Vietnamese primary math quizzes (Kindergarten–Grade 5)
RULES:
Generate exactly {{N}} questions, with difficulty increasing from Q1 → Q{{N}}.
Vary question types; do not repeat the same type in adjacent questions; use each type no more than 2 times.
{{PROBE_RULE}}
Each question must have exactly 4 answers: A, B, C, D. There must be exactly 1 correct answer and the distractors must be plausible.
All child-facing content must be in Vietnamese.
Every question must be solvable using only the content shown in the question.
MODE by {grade}:
"Mẫu giáo" → KG mode (icons are allowed).
"Lớp 1" through "Lớp 5" → NUM mode (NO emoji).
GENERATION SOURCE:
KG: Start from 1 of the 15 KG question types → generate question_name + answers.
NUM: Start from a question_topic taken from the correct-grade textbook (one of: Chân Trời Sáng Tạo / Kết Nối Tri Thức / Cánh Diều) → select a suitable question type → generate question_name + answers.
→ Do NOT start from the question type. The type must be a consequence of the selected topic.
→ If a topic can map to multiple question types, select an unused type that is not repeated in the adjacent question.
TEXTBOOK RULE: Use exactly ONE textbook for the entire quiz, selected from {Chân Trời Sáng Tạo, Kết Nối Tri Thức, Cánh Diều}. State the selected textbook in short_text. The topic sequence must follow the grade-level scope of that textbook.
KG:
question_name: may contain ONLY digits, emojis, mathematical symbols (+ − = > < ? |). No letters.
"|" is used as a group separator. Maximum 10 icons per group.
Do not use the same icon category for 3 consecutive questions.
{{KG_GRADE}}
ICONS:
Food: 🍎🍊🍐🍌🍉🍇🍓🍒🍑🍍🥝🥕🌽🍅🥦🥒🍭🍬🍪🍩🎂
Animals: 🐶🐱🐭🐹🐰🦊🐻🐼🐨🐯🦁🐮🐷🐸🐵🐔🐧🐦🐤🦆🦉
Water: 🐟🐠🐡🦋🐝🐞🐢
Vehicles: 🚗🚕🚌🚎🚲🛵🚂✈️🚁🚢
Toys: 📚⭐️🎈⚽️🧸✏️🖍️🎁
Shapes: 🔴🟡🟢🔵🟠🟣🟥🟨🟩🟦🟧🟪
15 KG QUESTION TYPES (STANDARDIZED):
Format: TYPE | question pattern | example question_name | correct answer
COUNTING | ICON×n = ? | 🍎🍎🍎 = ? | 3
NUM_MATCH | N = ? (corresponding icons) | 3 = ? | 🍎🍎🍎
NUM_SEQ | N N ? N | 1 2 ? 4 | 3
BEFORE_AFTER | N ? N | 2 ? 4 | 3
ADD | G + G = ? | 🍎🍎 + 🍎 = ? | 3
SUB | G − G = ? | 🍎🍎🍎 − 🍎 = ? | 2
COMP | G = G + ? | 🍎🍎🍎 = 🍎🍎 + ? | 1
CMP | G ? G → >, <, = | 🐶🐶🐶 ? 🐱🐱 | >
QTY_MATCH | G = ? (number) | 🍎🍎🍎 = ? | 3
PATTERN | ICON×4 ? | 🔴🟡🔴🟡🔴 ? | 🟡
ODD_ONE | I | I | I | I (select the different one) | 🍎 | 🍊 | 🐶 | 🍌 | 🐶
SORT_CLS | I×3 | ? (same group) | 🐶🐱🐰 | ? | 🐭
SAME_DIFF | G ? G → =, ≠ | 🍎🍎🍎 ? 🍊🍊🍊 | ≠
COLOR_MATCH | C = ? (count colors) | 🔴🔴🔴 = ? | 3
POSITION | rule + ? | 🍌🐱🍌 ? 🍌🐱 | 🍌
NUM (GRADE 1–5):
{{NUM_GRADE}}
question_name + answers[].content: Vietnamese text, numbers, math symbols; NO decorative emoji.
TEXTBOOKS: Select exactly one textbook for the quiz:
• Chân Trời Sáng Tạo
• Kết Nối Tri Thức
• Cánh Diều
COVERAGE:
• Q1–Q2: beginning-of-year content (easy).
• Q3–Q4: mid-year content (Q3 = light next-grade content).
• Q5–Q6: end-of-year content (Q6 = light next-grade content).
• Q7–Q10: cumulative content with increasing difficulty; word problems may be included.
The quiz must cover all of the following areas:
• Arithmetic / Numbers
• Quantities & Measurement
• Geometry & Geometric Measurement
• Statistics & Probability
• Word Problems
Use correct units (cm, cm², cm³, etc.) and correct Vietnamese mathematical terminology.
VALIDATION:
question_name: no letters in KG / no emoji in Grades 1–5. Do not include the answer in question_name.
Each question must have exactly 1 correct answer among A–D.
Values must be appropriate for the target grade.
KG: maximum 10 icons per group, all icons must come from the allowed list, and CMP questions must compare groups with different quantities.
If generation gets stuck → KG: use COUNTING or PATTERN. NUM: choose an easier topic, similar to ADD/PATTERN.
If {grade} ≠ "Mẫu giáo" and any emoji appears → regenerate using numbers only.
If {grade} = "Mẫu giáo" and question_name contains letters → regenerate without letters.
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
// the rule, the KG question_grade line and the NUM one. The positions
// come from ProbePositions so the prompt asks for exactly what the server
// enforces afterwards.
func examProbeLinesEN(probes []int) (rule, kg, num string) {
	if len(probes) == 0 {
		return "All questions must use content from the current grade.",
			`question_grade: "Mẫu giáo" for all questions.`,
			`question_grade: "Lớp N" for all questions.`
	}
	and := joinPositions(probes, " and ")
	rule = fmt.Sprintf("%s must use content from the next grade; all other questions must use content from the current grade. For Grade %d, %s use early Grade %d content.",
		and, enum.ExamGradeMax, and, GradeProbeCeiling)
	kg = fmt.Sprintf(`question_grade: %s = "Lớp 1"; all other questions = "Mẫu giáo".`, and)
	num = fmt.Sprintf(`question_grade: "Lớp N"; %s = "Lớp N+1" (Grade %d → "Lớp %d").`, and, enum.ExamGradeMax, GradeProbeCeiling)
	return rule, kg, num
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
	rule, kg, num := examProbeLinesEN(ProbePositions(in.ExamType, n))
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
	if avoid := examAvoidBlock(QuizLanguageEnglish, in.Avoid); avoid != "" {
		out += "\n\n" + avoid
	}
	return out
}
