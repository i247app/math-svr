package bot

import "fmt"

// English prompt templates. ai_review is emitted in English. JSON keys
// are always English regardless of QuizLanguage — only ai_review and
// the user-facing parts of system text switch language.

const systemGenerateEN = `You are a math quiz generator for Vietnamese primary-school students (Grades 1-5).

Generate EXACTLY 5 multiple-choice questions calibrated to the requested grade, semester, and curriculum.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D.
- Exactly one answer is correct.
- "question_name" contains ONLY numbers and operators (+, -, *, /, ^, parentheses, "?") — no narrative, no LaTeX, no images, no Vietnamese text.
- Use ASCII fractions like "1/2", never "½".
- Do not repeat questions.

OUTPUT RULES:
- Return ONLY a JSON array matching the schema below. No prose, no markdown fences, no trailing commentary.
- The array MUST have exactly 5 items, "question_number" 1..5 in order.

SCHEMA:
[
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
`

const systemReinforceEN = `You are a math quiz generator for Vietnamese primary-school students (Grades 1-5).

You will be given the student's previous quiz, their answers, and an AI review of their performance. Generate a NEW quiz of EXACTLY 5 multiple-choice questions that reinforces the topics the student got wrong or struggled with, while keeping the difficulty appropriate for their grade, semester, and curriculum.

CONTENT RULES:
- Each question has EXACTLY 4 answers labeled A, B, C, D; exactly one correct.
- "question_name" contains ONLY numbers and operators — no narrative, no LaTeX, no Vietnamese text.
- Use ASCII fractions like "1/2".
- Do not repeat questions verbatim from the previous quiz; create variations that target the same skills.

OUTPUT RULES:
- Return ONLY a JSON array matching the schema below. No prose, no markdown fences.
- The array MUST have exactly 5 items, "question_number" 1..5 in order.

SCHEMA:
[
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
`

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

func userGenerateEN(typ QuizType, in QuizPromptInput) string {
	guidance := "Calibrate difficulty to the grade and semester."
	if typ == QuizTypePractice {
		guidance = "Calibrate difficulty to the semester checkpoint; this is a routine practice round."
	}
	return fmt.Sprintf(`Generate a %s quiz for this student context:
- Grade: %s
- Semester: %s
- Curriculum: %s

%s`, typ, in.Grade, in.Semester, in.Program, guidance)
}

func userReinforceEN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Generate a %s reinforce quiz for this student context:
- Grade: %s
- Semester: %s
- Curriculum: %s

Previous quiz questions (JSON): %s
Student's previous answers (JSON): %s
AI review of previous performance: %s

Target the weak spots from the previous review while keeping difficulty appropriate for the configured grade and semester.`, typ, in.Grade, in.Semester, in.Program,
		in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview)
}

func userGradeEN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Grade this %s quiz.

Quiz questions (JSON): %s
Student's answers (JSON): %s`, typ, in.Questions, in.Answers)
}

func userGradeReinforceEN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Grade this reinforce %s quiz.

Quiz questions (JSON): %s
Student's answers (JSON): %s
Student's currently configured grade: %s`, typ, in.Questions, in.Answers, in.CurrentGrade)
}
