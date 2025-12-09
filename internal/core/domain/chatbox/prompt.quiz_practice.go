package domain

var PromptForGenerateQuizPractice = `
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

CRITICAL FORMATTING RULES - DO NOT USE LATEX:
- DO NOT use LaTeX format like \frac{1}{2}, \times, \div, \sqrt, etc.
- DO NOT use curly braces {} for mathematical expressions
- DO NOT use backslashes \ in any form

USE SIMPLE TEXT FORMAT INSTEAD:
- Fractions: Use "/" (e.g., "1/2" not "\frac{1}{2}" or "frac{1}{2}")
- Multiplication: Use "x" or "*" (e.g., "2 x 3" not "2 \times 3" or "2 \\times 3")
- Division: Use "/" or "÷" (e.g., "6 / 2" not "6 \div 2")
- Square root: Use "√" or write "square root of" (e.g., "√16" not "\sqrt{16}")
- Exponents: Use "^" (e.g., "2^3" not "2^{3}")
- Parentheses: Use only ( ) not { } or [ ]

EXAMPLES OF CORRECT FORMAT:
✓ "1/2 + 1/4 = ?"
✓ "2 x 3 = ?"
✓ "10 / 2 = ?"
✓ "√16 = ?"
✓ "2^3 = ?"
✓ "(3 + 2) x 4 = ?"

EXAMPLES OF WRONG FORMAT (DO NOT USE):
✗ "\frac{1}{2} + \frac{1}{4} = ?"
✗ "frac{1}{2} + frac{1}{4} = ?"
✗ "2 \times 3 = ?"
✗ "10 \div 2 = ?"
✗ "\sqrt{16} = ?"
✗ "2^{3} = ?"

Requirements:
- Return ONLY the JSON array (no line break (\n), no need for spaces but can space for question of name or content of answer), nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "question_name" must be included only number and mathematical operators without additional text
- "right_answer" must be one of: A, B, C, or D
- Use simple text format for ALL mathematical expressions (no LaTeX)
- Ensure all JSON is properly formatted with correct quotes and commas
- Questions should be appropriate for level and focus on concepts I give you above
- Use %s for all questions and answers
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
- Use %s for the ai_review (note: should be less than 200 characters)
`

var PromptForGenerateReinforceQuizPractice = `
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

CRITICAL FORMATTING RULES - DO NOT USE LATEX:
- DO NOT use LaTeX format like \frac{1}{2}, \times, \div, \sqrt, etc.
- DO NOT use curly braces {} for mathematical expressions
- DO NOT use backslashes \ in any form

USE SIMPLE TEXT FORMAT INSTEAD:
- Fractions: Use "/" (e.g., "1/2" not "\frac{1}{2}" or "frac{1}{2}")
- Multiplication: Use "x" or "*" (e.g., "2 x 3" not "2 \times 3" or "2 \\times 3")
- Division: Use "/" or "÷" (e.g., "6 / 2" not "6 \div 2")
- Square root: Use "√" or write "square root of" (e.g., "√16" not "\sqrt{16}")
- Exponents: Use "^" (e.g., "2^3" not "2^{3}")
- Parentheses: Use only ( ) not { } or [ ]

EXAMPLES OF CORRECT FORMAT:
✓ "1/2 + 1/4 = ?"
✓ "2 x 3 = ?"
✓ "10 / 2 = ?"
✓ "√16 = ?"
✓ "2^3 = ?"
✓ "(3 + 2) x 4 = ?"

EXAMPLES OF WRONG FORMAT (DO NOT USE):
✗ "\frac{1}{2} + \frac{1}{4} = ?"
✗ "frac{1}{2} + frac{1}{4} = ?"
✗ "2 \times 3 = ?"
✗ "10 \div 2 = ?"
✗ "\sqrt{16} = ?"
✗ "2^{3} = ?"

Requirements:
- Return ONLY the JSON array (no line break (\n), no need for spaces but can space for question of name or content of answer), nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "question_name" must be included only number and mathematical operators without additional text
- "right_answer" must be one of: A, B, C, or D
- Use simple text format for ALL mathematical expressions (no LaTeX)
- Ensure all JSON is properly formatted with correct quotes and commas
- Questions should focus on areas for improvement mentioned in the review
- Use %s for all questions and answers
`
