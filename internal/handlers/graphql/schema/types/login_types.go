package types

import (
	"github.com/graphql-go/graphql"
)

// LoginInput represents the login request
var LoginInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "LoginInput",
	Description: "Input for user login",
	Fields: graphql.InputObjectConfigFieldMap{
		"login_name": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Email or phone number",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User password",
		},
		"device_uuid": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Device UUID (optional)",
		},
		"device_name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Device name (optional)",
		},
	},
})

// LoginResponseType represents the login response
var LoginResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "LoginResponse",
	Description: "Login response with user info and session",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type:        UserType,
			Description: "Authenticated user information",
		},
		"needs_2fa": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Whether 2FA is required",
		},
		"is_secure": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Whether the login is secure",
		},
		"login_status": &graphql.Field{
			Type:        graphql.String,
			Description: "Login status",
		},
	},
})
