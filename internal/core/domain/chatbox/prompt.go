package domain

var PromptMathQuizNew = `
Generate 5 multiple-choice math questions for %s level (%s).

CRITICAL: Your response MUST be ONLY valid JSON. Do not include any text, explanations, markdown formatting, or code blocks before or after the JSON.

Return a JSON array with exactly this structure:
[
  {
    "question_number": 1,
    "question_name": "What is 2 + 2?",
    "answers": [
      {"label": "A", "content": "3"},
      {"label": "B", "content": "4"},
      {"label": "C", "content": "5"},
      {"label": "D", "content": "6"}
    ],
    "right_answer": "B"
  }
]

Requirements:
- Return ONLY the JSON array (no line break (\n), no need for spaces but can space for question of name or content of answer), nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "right_answer" must be one of: A, B, C, or D
- Ensure all JSON is properly formatted with correct quotes and commas
- Questions should be appropriate for level and focus on concepts I give you above
- Use Vietnamese for all questions and answers
`

var SubmitQuizAnswerPrompt = `
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
- Use Vietnamese for the ai_review
`

var PromptMathQuizPractice = ``

var PromptMathQuizExam = ``
