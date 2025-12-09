package domain

var PromptMathQuizNew = `
Generate 5 multiple-choice math questions for %s of the %s program.

CRITICAL: Your response MUST be ONLY valid JSON. Do not include any text, explanations, markdown formatting, or code blocks before or after the JSON.

Return a JSON array with exactly this structure:
[
  {
    "question_number": 1,
    "question_name": "1 + 1 = ?",
    "answers": [
      {"label": "A", "content": "1"},
      {"label": "B", "content": "2"},
      {"label": "C", "content": "3"},
      {"label": "D", "content": "4"}
    ],
    "right_answer": "B"
  }
]

NOTICE: In the example above, LaTeX backslashes are properly escaped as double backslash (\\frac not \frac).

Requirements:
- Return ONLY the JSON array (no line break (\n), no need for spaces but can space for question of name or content of answer), nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "question_name" should be included only number and mathematical operators without additional text
- "right_answer" must be one of: A, B, C, or D
- CRITICAL: All backslashes in JSON strings MUST be escaped with double backslash (\\)
  Example: For LaTeX \frac use \\frac, for \{ use \\{, for \sqrt use \\sqrt
  Don't response format like: {{frac{1}{2}}}, response should be: 1/2
- Ensure all JSON is properly formatted with correct quotes and commas
- Questions should be appropriate for level and focus on concepts I give you above
- Use %s for all questions and answers
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
- Use %s for the ai_review (note: should be less than 200 characters)
`

var PromptMathQuizPractice = `
Generate a new math quiz with 5 multiple-choice questions based on the user's previous performance review and answers below:

+ Question Informations: %s
+ User's Previous Answers: %s
+ AI Review of User's Performance: %s

CRITICAL: Your response MUST be ONLY valid JSON. Do not include any text, explanations, markdown formatting, or code blocks before or after the JSON.

Return a JSON array with exactly this structure:
[
  {
    "question_number": 1,
    "question_name": "1 + 1 = ?",
    "answers": [
      {"label": "A", "content": "1"},
      {"label": "B", "content": "2"},
      {"label": "C", "content": "3"},
      {"label": "D", "content": "4"}
    ],
    "right_answer": "B"
  }
]

NOTICE: In the example above, LaTeX backslashes are properly escaped as double backslash (\\frac not \frac).

Requirements:
- Return ONLY the JSON array (no line break (\n), no need for spaces but can space for question of name or content of answer), nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "question_name" should be included only number and mathematical operators without additional text
- "right_answer" must be one of: A, B, C, or D
- CRITICAL: All backslashes in JSON strings MUST be escaped with double backslash (\\)
  Example: For LaTeX \frac use \\frac, for \{ use \\{, for \sqrt use \\sqrt
  Don't response format like: {{frac{1}{2}}}, response should be: 1/2
- Ensure all JSON is properly formatted with correct quotes and commas
- Questions should focus on areas for improvement mentioned in the review
- Use %s for all questions and answers
`

var SubmitQuizAnswerAssessmentPrompt = `
You are a math quiz grading assistant with grade assessment capabilities.
Based on the user's answers and question information below, analyze and calculate the quiz results, and predict their appropriate grade level.

+ Question Information: %s
+ User's Answers: %s
+ User's Current Grade: %s

Given the user's answers to a quiz, provide the following in a JSON object:
- total_questions: Total number of questions in the quiz
- correct_number: Number of questions the user answered correctly
- score_percentage: The user's score as a percentage (correct_number / total_questions * 100)
- ai_review: A brief review of the user's performance, highlighting strengths and areas for improvement
- ai_detect_grade: Based on the difficulty of questions and user's performance, predict the most appropriate grade level for this user (e.g., "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", etc.)

Return ONLY the JSON object with this structure:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Great job! You have a strong understanding of basic math concepts, but your knowledge of subtraction needs improvement. Focus on practicing subtraction problems to improve your skills.",
  "ai_detect_grade": "Grade 3"
}

Requirements:
- Return ONLY the JSON object (no line break (\n), no need for spaces), nothing else
- Ensure all JSON is properly formatted with correct quotes and commas
- Use %s for the ai_review (note: should be less than 200 characters)
- ai_detect_grade should be in the format "Grade X" where X is a number (1-12) or "Kindergarten"
- Consider the question difficulty, user's performance, and current grade when predicting the appropriate grade level
`
