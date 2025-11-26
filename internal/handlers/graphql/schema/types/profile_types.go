package types

import (
	"github.com/graphql-go/graphql"
)

// ProfileType represents the Profile GraphQL type
var ProfileType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Profile",
	Description: "User educational profile",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Unique identifier (UUID)",
		},
		"uid": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID",
		},
		"email": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User Email",
		},
		"name": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User Name",
		},
		"phone": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User Phone",
		},
		"grade": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Selected grade level",
		},
		"level": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Selected difficulty level",
		},
		"status": &graphql.Field{
			Type:        graphql.String,
			Description: "Profile status: ACTIVE, INACTIVE",
		},
		"created_at": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Timestamp of creation",
		},
		"modified_at": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Timestamp of last modification",
		},
	},
})

// CreateProfileInput represents input for creating a profile
var CreateProfileInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "CreateProfileInput",
	Description: "Input for creating a user profile",
	Fields: graphql.InputObjectConfigFieldMap{
		"uid": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID",
		},
		"grade": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Grade level",
		},
		"level": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Difficulty level",
		},
	},
})

// UpdateProfileInput represents input for updating a profile
var UpdateProfileInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "UpdateProfileInput",
	Description: "Input for updating a user profile",
	Fields: graphql.InputObjectConfigFieldMap{
		"uid": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID",
		},
		"grade": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Grade level (optional)",
		},
		"level": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Difficulty level (optional)",
		},
	},
})
