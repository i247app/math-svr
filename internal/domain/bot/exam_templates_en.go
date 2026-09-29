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
// The system prompt is the teaching team's own format (CASE KG / CASE NUM
// and the KG question types stand in for the Vietnamese GRADE PROFILE), so
// the two system prompts are deliberately NOT line-for-line twins. The
// user message is different: it carries what varies per request, and it
// must carry the same facts in both languages — a fact only one template
// renders is a request field that silently stops working the day the
// prompt language flips. It therefore mirrors buildUserExamVN block for
// block: current_grade, the LEVEL PROFILE of a GRADE review, the practice
// brief of a PRACTICE round, and the curriculum context. The GRADE PROFILE
// is the one block left out, on purpose — CASE KG / CASE NUM replace it.

// systemExamENTmpl is the whole system prompt, filled by strings.Replacer:
// {{N}} is the question count; {{EMOJI}} is examAllowedEmoji; {{PROBE_RULE}}, {{KG_GRADE}} and
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
* No repeated question or calculation.
* {{PROBE_RULE}}
* Exactly 4 answers A–D; exactly 1 correct answer; plausible distractors.
* Each question must include all data needed to solve it; never leave the question incomplete.
* Fractions in ASCII (1/2, 3/4); never Unicode fractions or LaTeX.
* All child-facing content must be in Vietnamese; never use English words.

CASE KG — if {grade} = "Mẫu giáo":
* {{KG_GRADE}}
* Start from a KG type, then generate the question.
* Icons allowed ONLY from this list: {{EMOJI}}
* No same icon category 3 questions in a row.
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

### ANSWER VALIDATION — STRICT:

For every question:
1. SOLVE & VERIFY FIRST
* Solve the question and recalculate to confirm the exact correct answer.
* NEVER GUESS.

2. GENERATE 4 OPTIONS A, B, C, D
* All 4 options MUST be different.
* The correct answer MUST appear EXACTLY ONCE.
* The other 3 options MUST be WRONG.
* NEVER duplicate or use the correct answer as a distractor.

3. VALIDATE
* A.content ≠ B.content ≠ C.content ≠ D.content
* The correct answer appears exactly once.
* All 3 distractors are incorrect.

4. ASSIGN THE LABEL LAST
* right_answer_content = the exact content of the correct option.
* right_answer_label = the A/B/C/D label of the option containing "right_answer_content".
* NEVER generate "right_answer_label" independently.
* The label MUST be derived from the actual A–D options.

FINAL CHECK:
* Solve → Verify → Generate A–D → Validate → Identify the correct option → Assign its label.
* If the answer or label does not match, FIX IT before returning the result.

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
		"{{EMOJI}}", examAllowedEmoji,
	).Replace(systemExamENTmpl)
}

// examPracticeBlockEN is examPracticeBlockVN in English — the same facts
// and the same two modes, rule for rule. The child's own stems, answers
// and topics are quoted as stored (Vietnamese): they are data the model
// must recognise, not instructions to translate.
//
// The probe positions are named outright rather than as "the probe rule":
// this system prompt never uses that phrase, it states the positions.
func examPracticeBlockEN(b *PracticeBrief, n int, probes []int) string {
	gradeRule := "keep the question_grade rules above"
	if len(probes) > 0 {
		gradeRule = fmt.Sprintf("keep the question_grade rules above (%s = next grade, every other question = current_grade)",
			joinPositions(probes, ", "))
	}

	var sb strings.Builder
	sb.WriteString("PRACTICE ROUND (built on the child's last sitting):\n")

	switch b.Mode {
	case enum.PracticeModeRetryWeak:
		fmt.Fprintf(&sb, "- The child got %d questions wrong. Weak topics (question_topic, most-missed first): %s.\n",
			len(b.Wrong), joinOr(b.WeakTopics, "(topic unknown)"))
		sb.WriteString("- Questions answered wrong:\n")
		for _, w := range b.Wrong {
			fmt.Fprintf(&sb, "  • %s — correct answer: %s; child chose: %s\n",
				strings.TrimSpace(w.Stem), w.RightAnswer, w.ChildAnswer)
		}
		fmt.Fprintf(&sb, "- REQUIRED: %s; all %d questions focus on the weak topics (more questions for the most-missed ones).\n", gradeRule, n)
		sb.WriteString("- Same skill as the missed questions but NEVER copied verbatim: change the numbers and the context.\n")
		if len(b.StrongTopics) > 0 {
			fmt.Fprintf(&sb, "- You may mix in 1–2 questions on topics the child got right, to consolidate: %s.", joinOr(b.StrongTopics, ""))
		}
	default: // ADVANCE
		fmt.Fprintf(&sb, "- The child got EVERY question right, in these topics: %s.\n", joinOr(b.StrongTopics, "(topic unknown)"))
		fmt.Fprintf(&sb, "- REQUIRED: %s; apart from those next-grade questions, none of the %d questions may go above current_grade.\n", gradeRule, n)
		sb.WriteString("- But make them harder WITHIN current_grade: larger numbers inside the allowed range, more steps, word problems;\n")
		sb.WriteString("  and/or move to other curriculum topics the child has not been tested on yet.")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// examContextEN is examContextVN in English: only the curriculum lines
// that carry a value.
func examContextEN(in ExamPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Semester: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Textbook: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildUserExamEN binds {grade} to the Vietnamese band label the system
// prompt's CASE switch and question_grade vocabulary are written in, then
// adds the per-request blocks in the same order as buildUserExamVN.
func buildUserExamEN(in ExamPromptInput, n int) string {
	var out strings.Builder
	out.WriteString("current_grade: " + ExamTitle(in.Grade) + "\n")

	// The intensity block refines the grade and must not be read as
	// licence to leave it.
	if in.ExamType == enum.ExamTypeGrade && in.Level != nil {
		if block := levelProfileBlock(QuizLanguageEnglish, *in.Level); block != "" {
			out.WriteString("\n" + block + "\n")
		}
	}

	if in.ExamType == enum.ExamTypePractice && in.Practice != nil {
		out.WriteString("\n" + examPracticeBlockEN(in.Practice, n, ProbePositions(in.ExamType, n)) + "\n")
	}

	// CASE NUM tells the model to pick a textbook itself; a stated one
	// must win over that, or the request field does nothing.
	if ctx := examContextEN(in); ctx != "" {
		out.WriteString("\nCURRICULUM (use only to choose topics — NOT to raise or lower difficulty; if a textbook is given, use it instead of choosing one):\n" + ctx + "\n")
	}
	// if avoid := examAvoidBlock(QuizLanguageEnglish, in.Avoid); avoid != "" {
	// 	out.WriteString("\n" + avoid + "\n")
	// }

	return strings.TrimRight(out.String(), "\n")
}
