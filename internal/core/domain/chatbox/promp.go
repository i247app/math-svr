package domain

var PromptMathQuizNew = `Can you send me 10 multiple choice questions about 1st grade math ?. 
	Just send me the json response(no need for spaces but can space for question of name or content of answer, special characters) wrapped in an array, 
	inside are objects with fields (question(string) ,answers (object), right_answer(string)) 
	question (name of question), answer object(list of answers A,B,C,D) 
	including 2 fields: name(name of answer like A or B or C or D) and content (answer description for each choice), 
	righ_answer (answer of the question like A or B or C or D).
	In addition, in your response, do not send information other than my request such as introduction, ending, etc. 
	Because I use your response to prepare data for the question list.`

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
