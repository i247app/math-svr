package domain

// var PromptMathQuizNew = `Can you send me 10 multiple choice questions about 1st grade math ?.
// 	Just send me the json response(no need for spaces but can space for question of name or content of answer, special characters) wrapped in an array,
// 	inside are objects with fields (question(string) ,answers (object), right_answer(string))
// 	question (name of question), answer object(list of answers A,B,C,D)
// 	including 2 fields: name(name of answer like A or B or C or D) and content (answer description for each choice),
// 	righ_answer (answer of the question like A or B or C or D).
// 	In addition, in your response, do not send information other than my request such as introduction, ending, etc.
// 	Because I use your response to prepare data for the question list.`

var PromptMathQuizNew = `Generate 5 multiple-choice math questions for %s level (%s).

CRITICAL: Your response MUST be ONLY valid JSON. Do not include any text, explanations, markdown formatting, or code blocks before or after the JSON.

Return a JSON array with exactly this structure:
[
  {
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
- Use Vietnamese for all questions and answers`

var PromptMathQuizPractice = `Generate 10 practice math questions for reinforcing concepts.

CRITICAL: Your response MUST be ONLY valid JSON. Do not include any text, explanations, markdown formatting, or code blocks before or after the JSON.

Return a JSON array with exactly this structure:
[
  {
    "question": "Solve: 5 × 3 = ?",
    "answers": [
      {"label": "A", "content": "12"},
      {"label": "B", "content": "15"},
      {"label": "C", "content": "18"},
      {"label": "D", "content": "20"}
    ],
    "right_answer": "B",
  }
]

Requirements:
- In json response format (no line break, no need for spaces but can space for question of name or content of answer)
- Return ONLY the JSON array, nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "right_answer" must be one of: A, B, C, or D
- "duration" is time in seconds (8-15 for practice level)
- Ensure all JSON is properly formatted with correct quotes and commas`

var PromptMathQuizExam = `Generate 10 exam-level math questions with higher difficulty for comprehensive assessment.

CRITICAL: Your response MUST be ONLY valid JSON. Do not include any text, explanations, markdown formatting, or code blocks before or after the JSON.

Return a JSON array with exactly this structure:
[
  {
    "question": "If x + 7 = 15, what is x?",
    "answers": [
      {"label": "A", "content": "6"},
      {"label": "B", "content": "7"},
      {"label": "C", "content": "8"},
      {"label": "D", "content": "9"}
    ],
    "right_answer": "C",
  }
]

Requirements:
- Return ONLY the JSON array, nothing else
- Each question must have exactly 4 answers with labels A, B, C, D
- "right_answer" must be one of: A, B, C, or D
- Questions should be more challenging and test deeper understanding
- Ensure all JSON is properly formatted with correct quotes and commas`
