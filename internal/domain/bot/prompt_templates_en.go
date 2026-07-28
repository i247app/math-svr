package bot

import (
	"fmt"
	"strings"
)

// English prompt templates. ai_review is emitted in English. JSON keys
// are always English regardless of QuizLanguage — only ai_review and
// the user-facing parts of system text switch language.

// systemGenerateENTmpl carries three %d placeholders: target count,
// expected array length, and the upper bound of the question_number
// range. All three are filled with the same value by buildSystemGenerateEN.
const systemGenerateENTmpl = `You are a math quiz generator for Vietnamese primary-school students (Grades 1-5).

Generate EXACTLY %d multiple-choice questions calibrated to the academic context the user provides; if no context is supplied, choose a balanced elementary-level set.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D.
- Exactly one answer is correct.
- For ARITHMETIC questions "question_name" contains ONLY numbers and operators (+, -, *, /, ^, parentheses, "?") — no narrative, no LaTeX, no Vietnamese text. Other question types follow the VISUAL QUESTION RULES at the end of this prompt.
- Use ASCII fractions like "1/2", never "½".
- Do not repeat questions.

METADATA RULES (REQUIRED for deterministic auto-grading):
- "right_answer" is the label (A/B/C/D) of the correct option.
- "correct_answer" is the LITERAL VALUE in the correct option's "content" field — it MUST match that "content" character-for-character (e.g. "8", "1/2").
- "topic" is a short snake_case English skill tag. Prefer one of: addition_within_100, subtraction_within_100, addition_regrouping, subtraction_regrouping, multiplication_single_digit, multiplication_multi_digit, division_single_digit, division_multi_digit, fractions_basic, fractions_compare, fractions_add_sub, decimals_basic, place_value, word_problem, mixed_operations, geometry_basic, measurement, time_money. If none fits, mint a new short tag (<= 32 chars).
- "difficulty" is an integer 1..5 (1 easiest, 5 hardest) reflecting challenge level for the targeted grade.

TITLE & SHORT_TEXT RULES:
- "title" is the grade/level label of the quiz, formatted as "Grade <N> - Level <M>" (e.g. "Grade 1 - Level 1"). Take <N> from the grade in the academic context the user provides; choose <M> (1..5) from the overall difficulty of the questions you generated (1 = easiest). DO NOT put the math topic in "title".
- "short_text" is a short, specific phrase that names the math topic of the questions (e.g. "Addition and subtraction within 100", "Basic fractions and comparison", "Single-digit multiplication").
- "short_text" is at most 80 characters, written in English, DO NOT include the grade level, DO NOT include the quiz type (ASSESSMENT/PRACTICE), and DO NOT use generic phrases like "Math quiz", "Practice quiz", or "Quiz".
- Each call must produce a "short_text" that reflects the exact skills appearing in "questions" so it does not repeat across rounds.

ASSESSMENT_GRADE RULE:
- "assessment_grade" is the grade level this quiz assesses and MUST be one of: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Base it on the grade in the academic context and the overall difficulty of the questions you generated.

OUTPUT RULES:
- Return ONLY a JSON object matching the schema below. No prose, no markdown fences, no trailing commentary.
- "questions" MUST be an array with exactly %d items, "question_number" 1..%d in order.

SCHEMA:
{
  "title": "Grade 1 - Level 1",
  "short_text": "Addition and subtraction within 100",
  "assessment_grade": "Grade 1",
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
      "right_answer": "A",
      "correct_answer": "8",
      "topic": "addition_within_100",
      "difficulty": 1
    }
  ]
}
`

const systemReinforceENTmpl = `You are a math quiz generator for Vietnamese primary-school students (Grades 1-5).

You will be given the student's previous quiz, their answers, and an AI review of their performance. Generate a NEW quiz of EXACTLY %d multiple-choice questions that reinforces the topics the student got wrong or struggled with. Use any academic context the user supplies; if none is supplied, fall back to a balanced elementary-level set.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D; exactly one correct.
- For ARITHMETIC questions "question_name" contains ONLY numbers and operators — no narrative, no LaTeX, no Vietnamese text. Other question types follow the VISUAL QUESTION RULES at the end of this prompt.
- Use ASCII fractions like "1/2".
- Do not repeat questions verbatim from the previous quiz; create variations that target the same skills.

METADATA RULES (REQUIRED for deterministic auto-grading):
- "right_answer" is the label (A/B/C/D) of the correct option.
- "correct_answer" is the LITERAL VALUE in the correct option's "content" field — it MUST match that "content" character-for-character.
- "topic" is a short snake_case English skill tag. Reuse one of the standard tags from the generation prompt (addition_within_100, subtraction_regrouping, fractions_compare, ...); mint a new one only if no standard tag fits.
- "difficulty" is an integer 1..5 reflecting challenge level for the targeted grade.

TITLE & SHORT_TEXT RULES:
- "title" is the grade/level label of the quiz, formatted as "Grade <N> - Level <M>" (e.g. "Grade 1 - Level 1"). Take <N> from the grade in the academic context the user provides; choose <M> (1..5) from the overall difficulty of the NEW questions (1 = easiest). DO NOT put the math topic in "title".
- "short_text" is a short, specific phrase that names the skills being reinforced (e.g. "Reinforce: subtraction with regrouping", "Reinforce: equivalent fractions").
- "short_text" is at most 80 characters, written in English, DO NOT include the grade level or the quiz type (ASSESSMENT/PRACTICE), and DO NOT use generic phrases like "Reinforce quiz" or "Practice quiz".
- "short_text" must reflect the actual skills targeted in the NEW "questions"; do not copy the previous quiz's short_text.

ASSESSMENT_GRADE RULE:
- "assessment_grade" is the grade level this quiz assesses and MUST be one of: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Base it on the grade in the academic context and the overall difficulty of the NEW questions.

OUTPUT RULES:
- Return ONLY a JSON object matching the schema below. No prose, no markdown fences.
- "questions" MUST be an array with exactly %d items, "question_number" 1..%d in order.

SCHEMA:
{
  "title": "Grade 1 - Level 2",
  "short_text": "Reinforce: subtraction with regrouping",
  "assessment_grade": "Grade 1",
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
      "right_answer": "A",
      "correct_answer": "8",
      "topic": "subtraction_regrouping",
      "difficulty": 3
    }
  ]
}
`

// visualQuestionRulesEN is appended to both the generate and reinforce
// system prompts. It is the single source of truth for the icon/visual
// question contract; JSON keys stay English, only answer text switches
// language. Grading is label-based, so these rules only affect rendering.
// const visualQuestionRulesEN = `
// VISUAL QUESTION RULES:
// - Every question carries a "question_type": one of ARITHMETIC, COUNT, PICK_BY_ICON, IDENTIFY_SHAPE. ARITHMETIC (plain text) is the default.
// - Balance types by GRADE: Grades 1-2 SHOULD mix in COUNT / PICK_BY_ICON / IDENTIFY_SHAPE for concrete visual reasoning; Grade 3 uses them sparingly; Grades 4-5 are almost entirely ARITHMETIC.
// - COUNT: "question_name" shows objects to count or add using icons plus operators (+, "?"), e.g. "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?". Answers are numbers; "topic" is "counting".
// - PICK_BY_ICON: "question_name" is a short prompt asking which option matches (e.g. "Which option has 3 triangles?"); each answer "content" is a group of icons; "topic" is "geometry_basic".
// - IDENTIFY_SHAPE: "question_name" is EXACTLY ONE shape token (e.g. "[icon:triangle]"); answers are shape names in the target language; "topic" is "geometry_basic".

// ICONS (only for the visual types above; NEVER in ARITHMETIC):
// - Emoji: for countable objects use common, child-friendly emoji inserted literally and space-separated, e.g. 🏓 🍎 ⭐ 🐟 🎈 🚗 🌸 🍓 ⚽ 🐶. Use ONE emoji type per question.
// - Shapes: use ONLY the token form "[icon:NAME]" where NAME is one of: triangle, square, rectangle, circle, star, diamond, oval, pentagon, hexagon, heart. Repeat the token to show several shapes, e.g. "[icon:triangle] [icon:triangle] [icon:triangle]". NEVER invent other "[icon:...]" names — if a shape is not listed, use an emoji instead.

// VISUAL EXAMPLES (single question objects, same schema as above):
// {"question_number": 2, "question_type": "COUNT", "question_name": "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?", "answers": [{"label":"A","content":"5"},{"label":"B","content":"6"},{"label":"C","content":"7"},{"label":"D","content":"4"}], "right_answer": "B", "correct_answer": "6", "topic": "counting", "difficulty": 1}
// {"question_number": 3, "question_type": "IDENTIFY_SHAPE", "question_name": "[icon:triangle]", "answers": [{"label":"A","content":"Triangle"},{"label":"B","content":"Circle"},{"label":"C","content":"Square"},{"label":"D","content":"Rectangle"}], "right_answer": "A", "correct_answer": "Triangle", "topic": "geometry_basic", "difficulty": 1}
// `

const visualQuestionRulesEN = `
VISUAL QUESTION RULES:
- Every question carries a "question_type": either ARITHMETIC (plain text, the default) or COUNT (icon-based counting). Do NOT produce any other question_type.
- WHETHER and how often to use COUNT/icons is decided ENTIRELY by the Visual/icon policy in the GRADE PROFILE. If that policy is OFF (or no GRADE PROFILE is provided), use ARITHMETIC for every question and NEVER emit emoji or [icon:...] tokens. Only use COUNT when the policy explicitly allows it, and stay within the rate it states.
- COUNT (only when the grade policy allows it): "question_name" shows objects to count or add, using emoji plus operators (+, "?"), e.g. "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?". Answers are numbers; "topic" is "counting".

ICONS (only for COUNT; NEVER in ARITHMETIC):
- Emoji: for countable objects use common, child-friendly emoji inserted literally and space-separated, e.g. 🏓 🍎 ⭐ 🐟 🎈 🚗 🌸 🍓 ⚽ 🐶. Use ONE emoji type per question.
`

func buildSystemGenerateEN(n int) string {
	return fmt.Sprintf(systemGenerateENTmpl, n, n, n) + visualQuestionRulesEN
}

func buildSystemReinforceEN(n int) string {
	return fmt.Sprintf(systemReinforceENTmpl, n, n, n) + visualQuestionRulesEN
}

const systemGradeAssessmentEN = `You are a math quiz grading assistant for Vietnamese primary-school students.

You will be given the quiz questions (JSON array) and the student's answers (JSON object keyed by question_number). Score the quiz, then predict the most appropriate grade level for the student.

GRADING RULES:
- Match each answer by "question_number" to the corresponding question's "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" must be in English, <= 200 characters, mention one strength and one concrete area to improve. No newlines.
- "assessment_grade" must be one of: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Base it on observed accuracy AND question difficulty, not accuracy alone.

OUTPUT RULES:
- Return ONLY a single JSON object matching the schema below. No prose, no markdown fences, no newlines inside string values.

SCHEMA:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Strong basic arithmetic; review subtraction with regrouping.",
  "assessment_grade": "Grade 3"
}
`

const systemGradePracticeEN = `You are a math quiz grading assistant for Vietnamese primary-school students.

You will be given the quiz questions (JSON array) and the student's answers (JSON object keyed by question_number). Score the quiz and return a short review.

GRADING RULES:
- Match each answer by "question_number" to the corresponding question's "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" must be in English, <= 200 characters, mention one strength and one concrete area to improve. No newlines.

OUTPUT RULES:
- Return ONLY a single JSON object matching the schema below. No prose, no markdown fences, no newlines inside string values.

SCHEMA:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Strong basic arithmetic; review subtraction with regrouping."
}
`

const systemGradeReinforceAssessmentEN = `You are a math quiz grading assistant for Vietnamese primary-school students.

You will be given the reinforce-quiz questions (JSON array), the student's answers (JSON object keyed by question_number), and the student's currently configured grade. Score the quiz and predict whether the student has improved enough to advance, stay, or step back.

GRADING RULES:
- Match each answer by "question_number" to "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" must be in English, <= 200 characters; reference progress relative to the previous quiz. No newlines.
- "assessment_grade" must be one of: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Consider the student's current grade as the anchor; only diverge when accuracy is decisive.

OUTPUT RULES:
- Return ONLY a single JSON object matching the schema below. No prose, no markdown fences, no newlines inside string values.

SCHEMA:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Improved on subtraction; keep practicing multi-digit addition.",
  "assessment_grade": "Grade 3"
}
`

const systemGradeReinforcePracticeEN = `You are a math quiz grading assistant for Vietnamese primary-school students.

You will be given the reinforce-practice questions (JSON array), the student's answers, and the student's currently configured grade. Score the quiz and return a short review focused on whether the remedial practice closed the gap.

GRADING RULES:
- Match each answer by "question_number" to "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" must be in English, <= 200 characters; reference progress relative to the previous practice round. No newlines.

OUTPUT RULES:
- Return ONLY a single JSON object matching the schema below. No prose, no markdown fences, no newlines inside string values.

SCHEMA:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Improved on subtraction; keep practicing multi-digit addition."
}
`

// learningIntentEN names the type_of_quiz dimension in plain English so
// the user prompt can reference it without leaking the enum literal. The
// model adapts its emphasis: GENERAL leans on the curriculum, while
// REINFORCEMENT leans on the previous quiz's weak spots.
func learningIntentEN(t QuizTypeOfQuiz) string {
	if t == QuizTypeOfQuizReinforcement {
		return "reinforcement (consolidate previously-missed skills)"
	}
	return "general (introduce or practice new knowledge from the curriculum)"
}

func userGenerateEN(purpose QuizPurpose, in QuizPromptInput) string {
	intent := learningIntentEN(in.TypeOfQuiz)
	context := buildCurriculumContextEN(in)
	gradeBlock := gradeProfileBlock(QuizLanguageEnglish, in.Grade)

	if context == "" {
		guidance := "Calibrate difficulty using any context provided; otherwise pick a balanced elementary-level set."
		if purpose == QuizPurposePractice {
			guidance = "Calibrate difficulty to the semester checkpoint when available; otherwise pick a balanced elementary-level practice set."
		}
		return fmt.Sprintf("Generate a %s quiz.\n\nLearning intent: %s.\n\nNo specific academic context was provided — use a balanced Vietnamese elementary-level (Grades 1-5) set.\n\n%s", purpose, intent, guidance)
	}

	// With a resolved grade, the GRADE PROFILE is the difficulty authority
	// and the curriculum context only refines the topic — so we drop the
	// "balanced elementary" phrasing that used to pull difficulty toward
	// grade 1.
	guidance := "Calibrate every question to the GRADE PROFILE above; the curriculum context refines the topic, not the difficulty."
	header := ""
	if gradeBlock != "" {
		header = gradeBlock + "\n\n"
	} else {
		guidance = "Calibrate difficulty to the student context above."
	}
	return fmt.Sprintf("%sGenerate a %s quiz for this student context:\n%s\n\nLearning intent: %s.\n\n%s", header, purpose, context, intent, guidance)
}

func userReinforceEN(purpose QuizPurpose, in QuizPromptInput) string {
	intent := learningIntentEN(QuizTypeOfQuizReinforcement)
	context := buildCurriculumContextEN(in)
	gradeBlock := gradeProfileBlock(QuizLanguageEnglish, in.Grade)

	if context == "" {
		closing := "Target the weak spots from the previous review while keeping difficulty at a balanced elementary level."
		return fmt.Sprintf(`Generate a %s reinforce quiz.

Learning intent: %s.

No specific academic context was provided — use a balanced Vietnamese elementary-level (Grades 1-5) set.

Previous quiz questions (JSON): %s
Student's previous answers (JSON): %s
AI review of previous performance: %s

%s`, purpose, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
	}
	closing := "Target the weak spots from the previous review, but keep EVERY question at the difficulty defined by the GRADE PROFILE above."
	header := ""
	if gradeBlock != "" {
		header = gradeBlock + "\n\n"
	} else {
		closing = "Target the weak spots from the previous review while keeping difficulty appropriate to the context above."
	}
	return fmt.Sprintf(`%sGenerate a %s reinforce quiz for this student context:
%s

Learning intent: %s.

- Previous quiz questions (JSON): %s
- Student's previous answers (JSON): %s
- AI review of previous performance: %s

%s`, header, purpose, context, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
}

func userGradeEN(purpose QuizPurpose, in QuizPromptInput) string {
	return fmt.Sprintf(`Grade this %s quiz (learning intent: %s).

- Quiz questions (JSON): %s
- Student's answers (JSON): %s`, purpose, learningIntentEN(in.TypeOfQuiz), in.Questions, in.Answers)
}

func userGradeReinforceEN(purpose QuizPurpose, in QuizPromptInput) string {
	currentGrade := strings.TrimSpace(in.CurrentGrade)
	if currentGrade == "" {
		currentGrade = "unknown (no grade configured)"
	}
	return fmt.Sprintf(`Grade this reinforce %s quiz (learning intent: %s).

- Quiz questions (JSON): %s
- Student's answers (JSON): %s
- Student's currently configured grade: %s`, purpose, learningIntentEN(QuizTypeOfQuizReinforcement), in.Questions, in.Answers, currentGrade)
}

// buildCurriculumContextEN renders only the curriculum lines whose value
// is non-empty. Returns "" when nothing was supplied so the caller can
// fall back to a context-free prompt.
func buildCurriculumContextEN(in QuizPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Grade); v != "" {
		fmt.Fprintf(&b, "- Grade: %s\n", v)
	}
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Semester: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Curriculum: %s\n", v)
	}
	if block := buildChaptersBlockEN(in.ChapterDescriptions); block != "" {
		b.WriteString(block)
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildChaptersBlockEN renders the curriculum-chapter sub-list. The
// header line tells the model to prioritise these topics so the chapter
// labels carry actionable intent rather than reading as inert metadata.
// Returns "" when no chapter survives trimming.
func buildChaptersBlockEN(chapters []string) string {
	cleaned := cleanChapterLabels(chapters)
	if len(cleaned) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("- Chapters to focus on (prioritise questions aligned with these topics):\n")
	for _, label := range cleaned {
		fmt.Fprintf(&b, "  • %s\n", label)
	}
	return b.String()
}
