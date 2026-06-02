package bot

import (
	"fmt"
	"strings"
)

// Exercise prompt templates — English. The output schema reuses the
// quiz `{title, questions}` shape so the existing parseGeneration
// helper handles both flows; the teacher's title overrides any title
// the model emits, but we still allow the field so the model has a
// familiar schema to follow.

const systemExerciseGenerateENTmpl = `You are a math exercise generator for Vietnamese primary-school students (Grades 1-5).

A teacher has set a topic by chapter and lesson. Generate EXACTLY %d multiple-choice questions tightly scoped to that lesson; use any grade / curriculum context the user supplies to calibrate difficulty.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D.
- Exactly one answer is correct.
- "question_name" contains ONLY numbers and operators (+, -, *, /, ^, parentheses, "?") — no narrative, no LaTeX, no images, no Vietnamese text.
- Use ASCII fractions like "1/2", never "½".
- Do not repeat questions.
- Stay inside the teacher's chapter + lesson scope; do not drift to other topics even if the grade allows them.

TITLE RULES:
- "title" is a short, specific phrase that names the math skill of the questions (e.g. "Addition within 10", "Equivalent fractions").
- Maximum 80 characters, English, DO NOT include the grade level or the word "exercise".

OUTPUT RULES:
- Return ONLY a JSON object matching the schema below. No prose, no markdown fences, no trailing commentary.
- "questions" MUST be an array with exactly %d items, "question_number" 1..%d in order.

SCHEMA:
{
  "title": "Addition within 10",
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
