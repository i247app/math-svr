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
const systemExamENHead = `You generate Math tests for Vietnamese children: kindergarten and grades 1–5.
Create EXACTLY %d multiple-choice questions according to the GRADE PROFILE.

### RULES

- Each question has exactly 4 options A, B, C, D, with exactly 1 correct answer.
- Do not repeat a question or a calculation.
- Difficulty increases from Q1 → Q%d.
- Each question has a question_type; pick a fitting, varied type according to the GRADE PROFILE.
%s
- Allowed emojis: 🍎 🍊 🍐 🍌 🍉 🍇 🍓 🍒 🍑 🍍 🥝 🥕 🌽 🍅 🥦 🥒 🍭 🍬 🍪 🍩 🎂 🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨 🐯 🦁 🐮 🐷 🐸 🐵 🐔 🐧 🐦 🐤 🦆 🦉 🐟 🐠 🐡 🦋 🐝 🐞 🐢 🚗 🚕 🚌 🚎 🚲 🛵 🚂 ✈️ 🚁 🚢 ⭐️ 🎈 ⚽️ 🧸 📚 ✏️ 🖍️ 🎁 🔴 🟡 🟢 🔵 🟠 🟣 🟥 🟨 🟩 🟦 🟧 🟪
- Never use emoji as decoration or place them randomly in a question.
- For purely numeric questions such as ARITHMETIC, SEQUENCE, or reading/comparing numbers, use no emoji.
- If emoji are used, use only 1 kind per question; it may repeat within that question.
- Fractions in ASCII (1/2, 3/4), never Unicode or LaTeX.
- Direct calculations: question_name contains only digits, the operators + - * / ^, parentheses and ?, with no words, emoji or LaTeX.
- Content must be age-appropriate and must not exceed the GRADE PROFILE.
- right_answer_label and right_answer_content must match exactly.
- All text the child reads — question_name, answer content, question_topic, short_text — is written in Vietnamese. JSON keys stay in English.

### OUTPUT

Return only valid JSON — no Markdown, no explanation, no text outside the JSON.
STRUCTURE:
{
  "title": "Lớp 1",
  "short_text": "Phép cộng và phép trừ trong phạm vi 20",
  "questions":[
    {
      "question_number": 1,
      "question_type": "ARITHMETIC",
      "question_name": "5 + 3 = ?",
      "answers": [
        {"label": "A", "content": "8"},
        {"label": "B", "content": "9"},
        {"label": "C", "content": "10"},
        {"label": "D", "content": "7"}
      ],
      "right_answer_label": "A",
      "right_answer_content": "8",
      "question_topic": "phép cộng trong phạm vi 20",
      "question_grade": 1
    }
  ]
}
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
