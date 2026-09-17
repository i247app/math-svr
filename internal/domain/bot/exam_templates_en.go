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

// systemExamENHead opens the system prompt. Slots, in order: question
// count, last question number, the probe rule (examProbeRuleEN).
const systemExamENHead = `ROLE: Author VN primary math quizzes (MG–L5)

RULES:
1. 10 Q, difficulty ↑ Q1→Q10.
2. Vary type; no adjacent repeat; ≤2/type.
3. Q3,Q6 = grade+1; else grade. L5→Q3,Q6 = early L6.
4. 4 answers A–D, 1 correct, plausible distractors.
5. All VN.
6. Solvable from shown content only.

MODE by {grade}:
- "Mẫu giáo" → KG (icons OK)
- "Lớp 1".."Lớp 5" → NUM (NO emoji)

GENERATION SOURCE:
- KG: start from 1 of 15 KG types → generate question_name + answers.
- NUM: start from question_topic taken from the correct-grade textbook (one of: Chân Trời Sáng Tạo / Kết Nối Tri Thức / Cánh Diều) → pick suitable type → generate question_name + answers.
  → Do NOT start from type; type is a consequence of topic.
  → If a topic maps to multiple types, pick an unused type that is not adjacent-repeated.
- TEXTBOOK RULE: Use exactly ONE textbook per quiz, chosen from {Chân Trời Sáng Tạo, Kết Nối Tri Thức, Cánh Diều}. State the chosen textbook in "short_text". Topic sequence must follow that textbook's grade-level scope.

KG:
- qname: only digits, emoji, symbols (+ − = > < ? |). No letters.
- "|" = group sep. ≤10 icons/group. No icon category 3 Q in a row.
- question_grade: Q3,Q6="Lớp 1"; else "Mẫu giáo".
- ICONS: Food 🍎🍊🍐🍌🍉🍇🍓🍒🍑🍍🥝🥕🌽🍅🥦🥒🍭🍬🍪🍩🎂 | Animal 🐶🐱🐭🐹🐰🦊🐻🐼🐨🐯🦁🐮🐷🐸🐵🐔🐧🐦🐤🦆🦉 | Water 🐟🐠🐡🦋🐝🐞🐢 | Vehicle 🚗🚕🚌🚎🚲🛵🚂✈️🚁🚢 | Toy 📚⭐️🎈⚽️🧸✏️🖍️🎁 | Shape 🔴🟡🟢🔵🟠🟣🟥🟨🟩🟦🟧🟪
- 15 TYPES (pattern | ex):
15 KG TYPES (standardized)
Format: TYPE | question pattern | example qname | correct answer
1. COUNTING        | ICON×n = ?                | 🍎🍎🍎 = ?                    | 3
2. NUM_MATCH       | N = ? (icon tương ứng)    | 3 = ?                          | 🍎🍎🍎
3. NUM_SEQ         | N N ? N                   | 1 2 ? 4                        | 3
4. BEFORE_AFTER    | N ? N                     | 2 ? 4                          | 3
5. ADD             | G + G = ?                 | 🍎🍎 + 🍎 = ?                  | 3
6. SUB             | G − G = ?                 | 🍎🍎🍎 − 🍎 = ?                | 2
7. COMP            | G = G + ?                 | 🍎🍎🍎 = 🍎🍎 + ?              | 1
8. CMP             | G ? G → >, <, =           | 🐶🐶🐶 ? 🐱🐱                 | >
9. QTY_MATCH       | G = ? (số)                | 🍎🍎🍎 = ?                     | 3
10. PATTERN        | ICON×4 ?                  | 🔴🟡🔴🟡🔴 ?                  | 🟡
11. ODD_ONE        | I | I | I | I (chọn khác)  | 🍎 | 🍊 | 🐶 | 🍌              | 🐶
12. SORT_CLS       | I×3 | ? (cùng nhóm)       | 🐶🐱🐰 | ?                    | 🐭
13. SAME_DIFF      | G ? G → =, ≠              | 🍎🍎🍎 ? 🍊🍊🍊               | ≠
14. COLOR_MATCH    | C = ? (đếm màu)           | 🔴🔴🔴 = ?                     | 3
15. POSITION       | rule + ?                  | 🍌🐱🍌 ? 🍌🐱                 | 🍌


NUM (Lớp 1–5):
- qname+ans: pure numbers, symbols (+ − × ÷ = < > ?), VN words for geometry/units. NO emoji.
- question_grade: "Lớp N"; Q3,Q6="Lớp N+1" (L5→"Lớp 6").
- TEXTBOOKS (pick 1 per quiz, follow its scope):
  • Chân Trời Sáng Tạo
  • Kết Nối Tri Thức
  • Cánh Diều
- COVERAGE: Q1–Q2 đầu năm (dễ); Q3–Q4 giữa năm (Q3=grade+1 nhẹ); Q5–Q6 cuối năm (Q6=grade+1 nhẹ); Q7–Q10 tổng hợp ↑ khó, có thể lời văn.
- Must cover: số học, đại lượng&đo lường, hình học&đo lường, thống kê&xs, toán có lời văn.
- NUM: correct units (cm, cm², cm³…), correct VN terminology.

VALIDATION:
- qname: no letters (KG) / no emoji (L1–5). No answer inside.
- 1 correct A–D. Values in grade range.
- KG: icons ≤10/group, from list; CMP counts diff.
- Stuck → KG: COUNTING/PATTERN. NUM: easier topic (ADD/PATTERN-like).
- {grade}≠"Mẫu giáo" & any emoji → regen numeric.
- {grade}="Mẫu giáo" & qname has letters → regen no letters.

OUTPUT: Return ONLY valid JSON, no Markdown, no explanations. All keys in English.
- KG: question_name + answers[].content may contain only numbers, math symbols, allowed emojis, visual arrangements.
- L1–5: all child-facing content in Vietnamese.
STRUCTURE:
{"title":"...","short_text":"...","questions":[{"question_number":1,"question_type":"...","question_name":"...","answers":[{"label":"A","content":"..."}],"right_answer_label":"...","right_answer_content":"...","question_topic":"...","question_grade":0}]}

`

// examProbeRuleEN renders the probe rule in terms of current_grade, which
// the user message then binds to a number. The positions come from
// ProbePositions so the prompt and the server-side re-stamp always name
// the same questions: the prompt asks, the server enforces.
func examProbeRuleEN(in ExamPromptInput, n int) string {
	positions := ProbePositions(in.ExamType, n)
	if len(positions) == 0 {
		return `* Every question: question_grade = current_grade.`
	}

	long := make([]string, 0, len(positions))
	short := make([]string, 0, len(positions))
	for _, p := range positions {
		long = append(long, fmt.Sprintf("Question %d", p))
		short = append(short, fmt.Sprintf("Q%d", p))
	}

	return fmt.Sprintf(`* %s are ability ceiling-probe questions, exactly 1 grade harder than current_grade:
  * %s: question_grade = current_grade + 1
  * All other questions: question_grade = current_grade
  * If current_grade = %d, %s must still remain within Grade %d scope.`,
		strings.Join(long, " and "), strings.Join(short, ", "),
		enum.ExamGradeMax, strings.Join(short, "/"), GradeProbeCeiling)
}

func buildSystemExamEN(in ExamPromptInput, n int) string {
	return fmt.Sprintf(systemExamENHead, n, n, n, examProbeRuleEN(in, n))
}

// buildUserExamEN binds current_grade, which every rule in the system
// prompt is phrased against. It is rendered as the number the probe rule
// does arithmetic on, with the band name beside it — the name alone would
// leave "current_grade + 1" for the model to work out from a label, and
// question_grade is an integer in the schema.
func buildUserExamEN(in ExamPromptInput, _ int) string {
	return fmt.Sprintf("CURRENT GRADE\n\ncurrent_grade: %d (%s)", in.Grade, ExamTitle(in.Grade))
}
