package bot

import (
	"fmt"
	"strings"
)

// Exercise prompt templates — English. The output schema reuses the
// quiz `{short_text, questions}` shape so the existing parseGeneration
// helper handles both flows. The teacher supplies the exercise title
// separately, so the model only emits "short_text" (the auto-generated
// topic description) here.

const systemExerciseGenerateENTmpl = `You are a math exercise generator for Vietnamese primary-school students (Grades 1-5).

A teacher has set a topic by chapter and lesson. Generate EXACTLY %d multiple-choice questions tightly scoped to that lesson; use any grade / curriculum context the user supplies to calibrate difficulty.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D.
- Exactly one answer is correct.
- "question_name" contains ONLY numbers and operators (+, -, *, /, ^, parentheses, "?") — no narrative, no LaTeX, no images, no Vietnamese text.
- Use ASCII fractions like "1/2", never "½".
- Do not repeat questions.
- Stay inside the teacher's chapter + lesson scope; do not drift to other topics even if the grade allows them.

METADATA RULES (REQUIRED for deterministic auto-grading):
- "right_answer" is the label (A/B/C/D) of the correct option.
- "correct_answer" is the LITERAL VALUE in the correct option's "content" field — it MUST match that "content" character-for-character (e.g. "8", "1/2").
- "topic" is a short snake_case English skill tag. Prefer one of: addition_within_100, subtraction_within_100, addition_regrouping, subtraction_regrouping, multiplication_single_digit, multiplication_multi_digit, division_single_digit, division_multi_digit, fractions_basic, fractions_compare, fractions_add_sub, decimals_basic, place_value, word_problem, mixed_operations, geometry_basic, measurement, time_money. If none fits, mint a new short tag (<= 32 chars).
- "difficulty" is an integer 1..5 (1 easiest, 5 hardest) reflecting challenge level for the targeted lesson.

SHORT_TEXT RULES:
- "short_text" is a short, specific phrase that names the math skill of the questions (e.g. "Addition within 10", "Equivalent fractions").
- Maximum 80 characters, English, DO NOT include the grade level or the word "exercise".
- "short_text" must reflect the skills appearing in "questions"; it is the auto-generated topic description (the teacher's own title is set separately).

OUTPUT RULES:
- Return ONLY a JSON object matching the schema below. No prose, no markdown fences, no trailing commentary.
- "questions" MUST be an array with exactly %d items, "question_number" 1..%d in order.

SCHEMA:
{
  "short_text": "Addition within 10",
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
      "right_answer": "A",
      "correct_answer": "8",
      "topic": "addition_within_100",
      "difficulty": 1
    }
  ]
}
`

func buildSystemExerciseGenerateEN(n int) string {
	return fmt.Sprintf(systemExerciseGenerateENTmpl, n, n, n)
}

func userExerciseGenerateEN(in ExercisePromptInput) string {
	scope := fmt.Sprintf("- Chapter: %s\n- Lesson: %s",
		strings.TrimSpace(in.ChapterName), strings.TrimSpace(in.LessonName))
	context := buildExerciseContextEN(in)
	if context == "" {
		return fmt.Sprintf(`Generate an in-class exercise scoped to the following teacher-set topic:
%s

No additional grade / curriculum context was supplied — calibrate to a balanced elementary level (Grades 1-5).

Keep every question inside the chapter + lesson scope above.`, scope)
	}
	return fmt.Sprintf(`Generate an in-class exercise scoped to the following teacher-set topic:
%s

Additional context:
%s

Use the additional context above to refine question difficulty, focus areas, and phrasing — without overriding the chapter + lesson scope, which is binding. The teacher's guidance, when present, is a hint, not a license to drift to other topics.`, scope, context)
}

// Grading prompt — English. Output schema matches QuizGradingResult so
// the existing parseGradedQuiz helper handles the response.
const systemExerciseGradeEN = `You grade multiple-choice math exercises for Vietnamese primary-school students.

You receive:
- "questions": the original items, each with a "right_answer" label (A/B/C/D).
- "answers": the student's chosen labels keyed by "question_number".

TASK:
- Match each answer against the corresponding "right_answer" to decide correctness.
- Any question without a matching answer is counted WRONG.
- Compute total_questions, correct_number, and score_percentage (floor to integer).
- Write a short "ai_review" (2-4 sentences, English): note strengths, common mistakes, and one specific revision tip tied to the chapter + lesson scope.

OUTPUT RULES:
- Return ONLY the JSON object below — no prose, no markdown fences.
- ai_review must be <= 200 characters, mention one strength and one concrete area to improve. No newlines.

SCHEMA:
{
  "total_questions": 10,
  "correct_number": 8,
  "score_percentage": 80,
  "ai_review": "..."
}
`

func userExerciseGradeEN(in ExercisePromptInput) string {
	return fmt.Sprintf(`Original questions (JSON):
%s

Student answers (JSON):
%s

Grade the submission and respond using the schema above.`, strings.TrimSpace(in.Questions), strings.TrimSpace(in.Answers))
}
