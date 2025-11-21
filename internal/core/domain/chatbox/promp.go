package domain

// var PromptMathQuizNew = `Can you send me 10 multiple choice questions about 1st grade math ?.
// 	Just send me the json response(no need for spaces but can space for question of name or content of answer, special characters) wrapped in an array,
// 	inside are objects with fields (question(string) ,answers (object), right_answer(string))
// 	question (name of question), answer object(list of answers A,B,C,D)
// 	including 2 fields: name(name of answer like A or B or C or D) and content (answer description for each choice),
// 	righ_answer (answer of the question like A or B or C or D).
// 	In addition, in your response, do not send information other than my request such as introduction, ending, etc.
// 	Because I use your response to prepare data for the question list.`

var PromptMathQuizNew = `
	Can you send me 10 multiple choice questions about 1st grade math ?.
	The core requirement was that the entire response be delivered as a single, 
	clean JSON array response (no space, no line break for json format, can space for question of name or content of answer, special characters), 
	specifically structured for data preparation. 
	This array must contain objects, each representing a question, 
	and include three mandatory fields: question (for the text), 
	answers (a structured list of choice objects(label, content) labeled A, B, C, D),
	duration (time based on question level - in seconds - max 15 seonds), 
	and right_answer (identifying the correct label). 
	Crucially, the user strictly prohibited any accompanying text, introductions, 
	or explanations outside of the required JSON output itself.`

var PromptMathQuizPractice = `Can you send me 10 multiple choice questions about 1st grade math ?. 
	Just send me the json response(no need for spaces, special characters) wrapped in an array, 
	inside are objects with fields (question(string) ,answers (object), right_answer(string)) 
	question (name of question), answer object(list of answers A,B,C,D) 
	including 2 fields: name(name of answer like A or B or C or D) and content (answer description for each choice), 
	righ_answer (answer of the question like A or B or C or D).
	In addition, in your response, do not send information other than my request such as introduction, ending, etc. 
	Because I use your response to prepare data for the question list.`

var PromptMathQuizExam = `Can you send me 10 multiple choice questions about 1st grade math ?.
	Just send me the json response(no need for spaces, special characters) wrapped in an array, 
	inside are objects with fields (question(string) ,answers (object), right_answer(string)) 
	question (name of question), answer object(list of answers A,B,C,D) 
	including 2 fields: name(name of answer like A or B or C or D) and content (answer description for each choice), 
	righ_answer (answer of the question like A or B or C or D).
	In addition, in your response, do not send information other than my request such as introduction, ending, etc. 
	Because I use your response to prepare data for the question list.`
