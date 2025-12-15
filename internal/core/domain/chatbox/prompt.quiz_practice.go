package domain

var PromptForGenerateQuizPractice = `
Generate 5 multiple-choice math questions for Vietnam %s %s where the grade has 4 term level of difficulty.

CRITICAL: Return ONLY a valid JSON array. Show problem only. NO extra text, explanations, or markdown formatting.
Use this sample multiple choice problem structure:
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
STRICT FORMATTING RULES: NO LATEX. Use simple text only (e.g., "1/2", "x", "^"). "question_name" must be numbers and operators ONLY. Exactly 5 questions with 4 answers (A, B, C, D) each.
`

var PromptForSubmitQuizPracticeAnswer = `
You are a math quiz grading assistant.
Base on the following user's answers below and the question informations belows, analyze and calculate the quiz results.
+ Question Informations: %s
+ User's Answers: %s

Given the user's answers to a quiz, provide the following in a JSON object:
- total_questions: Total number of questions in the quiz
- correct_number: Number of questions the user answered correctly
- score_percentage: The user's score as a percentage (correct_number / total_questions * 100)
- ai_review: A brief review of the user's performance, highlighting strengths and areas for improvement, problem needs to be resolved.

Return ONLY the JSON object with this structure:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Great job! You have a strong understanding of basic math concepts, 
  but your knowledge of subtraction is not good enough. Focus on practicing subtraction problems to improve your skills."
}

Requirements:
- Return ONLY the JSON object (no line break (\n), no need for spaces), nothing else
- Ensure all JSON is properly formatted with correct quotes and commas
- "ai_review" should be less than 200 characters
`

var PromptForGenerateReinforceQuizPractice = `
Generate a new math quiz with 5 multiple-choice questions based on the user's previous performance review and answers below:

+ Question Informations: %s
+ User's Previous Answers: %s
+ AI Review of User's Performance: %s

CRITICAL: Return ONLY a valid JSON array. Show problem only. NO extra text, explanations, or markdown formatting.
Use this sample multiple choice problem structure:
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
STRICT FORMATTING RULES: NO LATEX. Use simple text only (e.g., "1/2", "x", "^"). "question_name" must be numbers and operators ONLY. Exactly 5 questions with 4 answers (A, B, C, D) each
`
