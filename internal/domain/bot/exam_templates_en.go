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
// English prompt will otherwise answer in English. Keep the two files in
// step: a rule added to one belongs in the other.

// systemExamENHead opens the system prompt. Slots, in order: question
// count, last question number, the probe rule (examProbeRuleEN).
const systemExamENHead = `Quyen Vo, [9/17/2026 11:47 PM]
ROLE: Author VN primary math quizzes (MG–L5)

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

// systemExamENTail closes the system prompt. One slot: question count.
const systemExamENTail = `
Self-check before answering: exactly %d questions, 4 options each, 1 correct answer, correct question_grade, no duplicates, correct difficulty, within the GRADE PROFILE.`

// examProbeRuleEN is examProbeRuleVN in English; see there for why the
// positions are computed rather than written out.
func examProbeRuleEN(in ExamPromptInput, n int) string {
	positions := ProbePositions(in.ExamType, n)
	if len(positions) == 0 {
		return `- Every question: question_grade = current_grade.`
	}

	labels := make([]string, 0, len(positions))
	for _, p := range positions {
		labels = append(labels, fmt.Sprintf("Q%d", p))
	}
	and := strings.Join(labels, " and ")
	list := strings.Join(labels, ", ")
	slash := strings.Join(labels, "/")

	return fmt.Sprintf(`- %s are CEILING PROBES, exactly 1 grade harder than current_grade:
  - %s: question_grade = current_grade + 1
  - All other questions: question_grade = current_grade
  - If current_grade = %d, %s stay within grade %d.
- Do not reveal that %s are probes.`,
		and, list, enum.ExamGradeMax, slash, GradeProbeCeiling, slash)
}

func buildSystemExamEN(in ExamPromptInput, n int) string {
	return fmt.Sprintf(systemExamENHead, n, n, examProbeRuleEN(in, n)) +
		fmt.Sprintf(systemExamENTail, n)
}

// examPracticeBlockEN is examPracticeBlockVN in English. The wrong
// answers it lists are quoted as served, so they stay Vietnamese inside
// the English brief — that is the material the round is built from.
func examPracticeBlockEN(b *PracticeBrief, n int) string {
	var sb strings.Builder
	sb.WriteString("PRACTICE ROUND (based on the round the child just took):\n")

	switch b.Mode {
	case enum.PracticeModeRetryWeak:
		fmt.Fprintf(&sb, "- The child got %d questions wrong, in these topics (most wrong first): %s.\n",
			len(b.Wrong), joinOr(b.WeakTopics, "(topic unknown)"))
		sb.WriteString("- The questions answered wrong:\n")
		for _, w := range b.Wrong {
			fmt.Fprintf(&sb, "  • %s — correct: %s; child chose: %s\n",
				strings.TrimSpace(w.Stem), w.RightAnswer, w.ChildAnswer)
		}
		fmt.Fprintf(&sb, "- REQUIRED: keep the RULES on question_grade (probes included); all %d questions focus on the topics answered wrong, weighting the topics with more mistakes.\n", n)
		sb.WriteString("- Same skill as the wrong answers but NEVER a verbatim repeat: change the numbers, change the context.\n")
		if len(b.StrongTopics) > 0 {
			fmt.Fprintf(&sb, "- 1-2 questions may revisit topics the child got right, to reinforce: %s.", joinOr(b.StrongTopics, ""))
		}
	default: // ADVANCE
		fmt.Fprintf(&sb, "- The child answered EVERYTHING correctly, in these topics: %s.\n", joinOr(b.StrongTopics, "(topic unknown)"))
		fmt.Fprintf(&sb, "- REQUIRED: keep the RULES on question_grade (probes included); apart from the probes, none of the %d questions moves to a higher grade.\n", n)
		sb.WriteString("- But they must be harder WITHIN that grade: larger numbers inside the allowed range, more steps, word problems;\n")
		sb.WriteString("  and/or move to other topics in the curriculum the child has not been tested on.")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// examContextEN renders only the curriculum lines that carry a value.
func examContextEN(in ExamPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Semester: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Curriculum: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildUserExamEN(in ExamPromptInput, _ int) string {
	var out strings.Builder

	fmt.Fprintf(&out, "\ncurrent_grade: %d (%s)\n", in.Grade, ExamTitle(in.Grade))

	return strings.TrimRight(out.String(), "\n")
}
