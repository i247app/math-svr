package types

import (
	"github.com/graphql-go/graphql"
)

// QuestionAnswerType represents a single answer option
var QuestionAnswerType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "QuestionAnswer",
	Description: "Answer option for a quiz question",
	Fields: graphql.Fields{
		"label": &graphql.Field{
			Type:        graphql.String,
			Description: "Answer label (A, B, C, D)",
		},
		"content": &graphql.Field{
			Type:        graphql.String,
			Description: "Answer content",
		},
	},
})

// QuestionType represents a quiz question
var QuestionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Question",
	Description: "Quiz question with multiple choice answers",
	Fields: graphql.Fields{
		"question_number": &graphql.Field{
			Type:        graphql.Int,
			Description: "Question number",
		},
		"question_name": &graphql.Field{
			Type:        graphql.String,
			Description: "Question text",
		},
		"answers": &graphql.Field{
			Type:        graphql.NewList(QuestionAnswerType),
			Description: "List of answer options",
		},
		"right_answer": &graphql.Field{
			Type:        graphql.String,
			Description: "Correct answer label",
		},
	},
})

// QuizAnswerType represents the quiz evaluation result
var QuizAnswerType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "QuizAnswer",
	Description: "Quiz evaluation result with score and AI review",
	Fields: graphql.Fields{
		"total_questions": &graphql.Field{
			Type:        graphql.Int,
			Description: "Total number of questions",
		},
		"correct_number": &graphql.Field{
			Type:        graphql.Int,
			Description: "Number of correct answers",
		},
		"score_percentage": &graphql.Field{
			Type:        graphql.Int,
			Description: "Score as percentage",
		},
		"ai_review": &graphql.Field{
			Type:        graphql.String,
			Description: "AI-generated review and feedback",
		},
	},
})

// GenerateQuizInput represents input for generating a quiz
var GenerateQuizInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "GenerateQuizInput",
	Description: "Input for generating a quiz",
	Fields: graphql.InputObjectConfigFieldMap{
		"uid": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID",
		},
		"type_of_task": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Type of quiz task",
		},
		"type_of_purpose": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Purpose of the quiz",
		},
		"message": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Custom message/prompt",
		},
	},
})

// SubmitQuizAnswerInput represents an answer to a single question
var SubmitQuizAnswerInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "SubmitQuizAnswerInput",
	Description: "Answer to a single quiz question",
	Fields: graphql.InputObjectConfigFieldMap{
		"question_number": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Question number",
		},
		"answer": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Selected answer",
		},
	},
})

// SubmitQuizInput represents input for submitting quiz answers
var SubmitQuizInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "SubmitQuizInput",
	Description: "Input for submitting quiz answers",
	Fields: graphql.InputObjectConfigFieldMap{
		"uid": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID",
		},
		"answers": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(SubmitQuizAnswerInputType)),
			Description: "List of answers",
		},
	},
})

// GenerateQuizPracticeInput represents input for generating practice quiz
var GenerateQuizPracticeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "GenerateQuizPracticeInput",
	Description: "Input for generating a practice quiz",
	Fields: graphql.InputObjectConfigFieldMap{
		"uid": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID",
		},
		"message": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Custom message/prompt",
		},
	},
})

// QuizResponseType represents the quiz generation response
var QuizResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "QuizResponse",
	Description: "Response with generated quiz questions",
	Fields: graphql.Fields{
		"user_latest_quiz_id": &graphql.Field{
			Type:        graphql.String,
			Description: "ID of the user's latest quiz",
		},
		"response": &graphql.Field{
			Type:        graphql.String,
			Description: "Response message",
		},
		"data": &graphql.Field{
			Type:        graphql.NewList(QuestionType),
			Description: "List of quiz questions",
		},
		"role": &graphql.Field{
			Type:        graphql.String,
			Description: "Response role (assistant)",
		},
		"model": &graphql.Field{
			Type:        graphql.String,
			Description: "AI model used",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Response timestamp",
		},
	},
})

// QuizEvaluationResponseType represents the quiz evaluation response
var QuizEvaluationResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "QuizEvaluationResponse",
	Description: "Response with quiz evaluation and AI review",
	Fields: graphql.Fields{
		"user_latest_quiz_id": &graphql.Field{
			Type:        graphql.String,
			Description: "ID of the user's latest quiz",
		},
		"response": &graphql.Field{
			Type:        graphql.String,
			Description: "Response message",
		},
		"data": &graphql.Field{
			Type:        QuizAnswerType,
			Description: "Quiz evaluation data",
		},
		"role": &graphql.Field{
			Type:        graphql.String,
			Description: "Response role (assistant)",
		},
		"model": &graphql.Field{
			Type:        graphql.String,
			Description: "AI model used",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Response timestamp",
		},
	},
})
