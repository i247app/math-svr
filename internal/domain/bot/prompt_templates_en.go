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
- "question_name" contains ONLY numbers and operators (+, -, *, /, ^, parentheses, "?") — no narrative, no LaTeX, no images, no Vietnamese text.
- Use ASCII fractions like "1/2", never "½".
- Do not repeat questions.

TITLE RULES:
- "title" is a short, specific phrase that names the math topic of the questions (e.g. "Addition and subtraction within 100", "Basic fractions and comparison", "Single-digit multiplication").
- Maximum 80 characters, written in English, DO NOT include the grade level, DO NOT include the quiz type (ASSESSMENT/PRACTICE), and DO NOT use generic titles like "Math quiz", "Practice quiz", or "Quiz".
- Each call must produce a title that reflects the exact skills appearing in "questions" so titles do not repeat across rounds.

OUTPUT RULES:
- Return ONLY a JSON object matching the schema below. No prose, no markdown fences, no trailing commentary.
- "questions" MUST be an array with exactly %d items, "question_number" 1..%d in order.

SCHEMA:
{
  "title": "Addition and subtraction within 100",
  "questions":[
    {
      "question_number": 1,
      "question_name": "5 + 3 = ?",
      "answers": [
        {"label": "A", "content": "8"},
        {"label": "B", "content": "9"},
        {"label": "C", "content": "10"},
        {"label": "D", "content": "7"}
      ],
      "right_answer": "A"
    }
  ]
}
`

const systemReinforceENTmpl = `You are a math quiz generator for Vietnamese primary-school students (Grades 1-5).

You will be given the student's previous quiz, their answers, and an AI review of their performance. Generate a NEW quiz of EXACTLY %d multiple-choice questions that reinforces the topics the student got wrong or struggled with. Use any academic context the user supplies; if none is supplied, fall back to a balanced elementary-level set.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D; exactly one correct.
- "question_name" contains ONLY numbers and operators — no narrative, no LaTeX, no Vietnamese text.
- Use ASCII fractions like "1/2".
- Do not repeat questions verbatim from the previous quiz; create variations that target the same skills.

TITLE RULES:
- "title" is a short, specific phrase that names the skills being reinforced (e.g. "Reinforce: subtraction with regrouping", "Reinforce: equivalent fractions").
- Maximum 80 characters, written in English, DO NOT include the grade level or the quiz type (ASSESSMENT/PRACTICE), and DO NOT use generic titles like "Reinforce quiz" or "Practice quiz".
- The title must reflect the actual skills targeted in the NEW "questions"; do not copy the previous quiz's title.

OUTPUT RULES:
- Return ONLY a JSON object matching the schema below. No prose, no markdown fences.
- "questions" MUST be an array with exactly %d items, "question_number" 1..%d in order.

SCHEMA:
{
  "title": "Reinforce: subtraction with regrouping",
  "questions":[
    {
      "question_number": 1,
      "question_name": "5 + 3 = ?",
      "answers": [
        {"label": "A", "content": "8"},
        {"label": "B", "content": "9"},
        {"label": "C", "content": "10"},
        {"label": "D", "content": "7"}
      ],
      "right_answer": "A"
    }
  ]
}
`

func buildSystemGenerateEN(n int) string {
	return fmt.Sprintf(systemGenerateENTmpl, n, n, n)
}

func buildSystemReinforceEN(n int) string {
	return fmt.Sprintf(systemReinforceENTmpl, n, n, n)
}

const systemGradeAssessmentEN = `You are a math quiz grading assistant for Vietnamese primary-school students.

You will be given the quiz questions (JSON array) and the student's answers (JSON object keyed by question_number). Score the quiz, then predict the most appropriate grade level for the student.

GRADING RULES:
- Match each answer by "question_number" to the corresponding question's "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" must be in English, <= 200 characters, mention one strength and one concrete area to improve. No newlines.
- "ai_detect_grade" must be one of: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Base it on observed accuracy AND question difficulty, not accuracy alone.

OUTPUT RULES:
- Return ONLY a single JSON object matching the schema below. No prose, no markdown fences, no newlines inside string values.

SCHEMA:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Strong basic arithmetic; review subtraction with regrouping.",
  "ai_detect_grade": "Grade 3"
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
- "ai_detect_grade" must be one of: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Consider the student's current grade as the anchor; only diverge when accuracy is decisive.

OUTPUT RULES:
- Return ONLY a single JSON object matching the schema below. No prose, no markdown fences, no newlines inside string values.

SCHEMA:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Improved on subtraction; keep practicing multi-digit addition.",
  "ai_detect_grade": "Grade 3"
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
	guidance := "Calibrate difficulty using the context above where provided; otherwise pick a balanced elementary-level set."
	if purpose == QuizPurposePractice {
		guidance = "Calibrate difficulty to the semester checkpoint when available; otherwise pick a balanced elementary-level practice set."
	}
	intent := learningIntentEN(in.TypeOfQuiz)
	context := buildCurriculumContextEN(in)
	if context == "" {
		return fmt.Sprintf("Generate a %s quiz.\n\nLearning intent: %s.\n\nNo specific academic context was provided — use a balanced Vietnamese elementary-level (Grades 1-5) set.\n\n%s", purpose, intent, guidance)
	}
	return fmt.Sprintf("Generate a %s quiz for this student context:\n%s\n\nLearning intent: %s.\n\n%s", purpose, context, intent, guidance)
}

func userReinforceEN(purpose QuizPurpose, in QuizPromptInput) string {
	closing := "Target the weak spots from the previous review while keeping difficulty appropriate to the context above; if no context was supplied, fall back to a balanced elementary-level set."
	intent := learningIntentEN(QuizTypeOfQuizReinforcement)
	context := buildCurriculumContextEN(in)
	if context == "" {
		return fmt.Sprintf(`Generate a %s reinforce quiz.

Learning intent: %s.

No specific academic context was provided — use a balanced Vietnamese elementary-level (Grades 1-5) set.

Previous quiz questions (JSON): %s
Student's previous answers (JSON): %s
AI review of previous performance: %s

%s`, purpose, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
	}
	return fmt.Sprintf(`Generate a %s reinforce quiz for this student context:
%s

Learning intent: %s.

- Previous quiz questions (JSON): %s
- Student's previous answers (JSON): %s
- AI review of previous performance: %s

%s`, purpose, context, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
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
