package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// The USER message of an exam generation: what varies per request.
//
// The system prompt is not here. It lives in ma_exam_prompts, one row per
// grade, written and rewritten by admins (/exams/prompts/*), and is sent
// verbatim. This file builds what goes beside it: current_grade, the LEVEL
// PROFILE of a GRADE review, the brief of a PRACTICE round, and the
// curriculum context. The instructions are English; the child's own stems,
// answers and topics are quoted as stored (Vietnamese).
//
// Every block here talks about the round in general terms ("the
// question_grade rules above", "the weak topics") so it fits whatever an
// admin writes into the system prompt, as long as that prompt keeps the
// probe questions at Q3 and Q6 (ProbePositions, which the server also uses
// to stamp question_grade).

// examPracticeBlockEN tells the model what the child just did and how the
// round must respond to it. The two modes are spelled out as separate rules
// rather than left to the model's judgement: "harder" without a bound
// drifts into the next grade, and the grade is the one thing a practice
// round must not move. The child's own stems, answers and topics are data
// the model must recognise, not instructions to translate.
//
// The probe positions are named outright, so the block does not depend on
// how the stored system prompt phrases its probe rule.
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

// examContextEN renders only the curriculum lines that carry a value, so a
// request with no semester or program still produces a coherent brief.
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

// buildUserExamEN names the grade by its Vietnamese band label ("Lớp 3"),
// then adds the per-request blocks that apply.
func buildUserExamEN(in ExamPromptInput, n int) string {
	var out strings.Builder
	out.WriteString("current_grade: " + GradeLabel(in.Grade) + "\n")

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

	// The system prompt lets the model follow any of the textbooks; a
	// stated one must win over that choice, or the request field does
	// nothing.
	if ctx := examContextEN(in); ctx != "" {
		out.WriteString("\nCURRICULUM (use only to choose topics — NOT to raise or lower difficulty; if a textbook is given, use it instead of choosing one):\n" + ctx + "\n")
	}
	// if avoid := examAvoidBlock(in.Avoid); avoid != "" {
	// 	out.WriteString("\n" + avoid + "\n")
	// }

	return strings.TrimRight(out.String(), "\n")
}

// examAvoidBlock lists the stems the child has already met, one per line.
// Blank stems are skipped; an empty list renders nothing so the caller can
// append it unconditionally.
func examAvoidBlock(stems []string) string {
	var sb strings.Builder
	for _, stem := range stems {
		if stem = strings.TrimSpace(stem); stem != "" {
			sb.WriteString("- " + stem + "\n")
		}
	}
	if sb.Len() == 0 {
		return ""
	}
	return "AVOID REPEATS\nThe child has already seen these questions; do not reuse or lightly reword them (change the numbers, objects, or structure):\n" +
		strings.TrimRight(sb.String(), "\n")
}

// joinOr renders a topic list, or the fallback when there is none.
func joinOr(items []string, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return strings.Join(items, ", ")
}
