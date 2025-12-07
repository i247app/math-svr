package types

import (
	"github.com/graphql-go/graphql"
)

// UserType represents the User GraphQL type
var UserType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "User",
	Description: "User account information",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Unique identifier (UUID)",
		},
		"name": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User full name",
		},
		"email": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User email address",
		},
		"phone": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User phone number",
		},
		"dob": &graphql.Field{
			Type:        graphql.String,
			Description: "User dob",
		},
		"avatar_key": &graphql.Field{
			Type:        graphql.String,
			Description: "URL to user avatar image",
		},
		"role": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User role: admin, user, guest",
		},
		"created_at": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Account creation timestamp",
		},
		"modified_at": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Account last modification timestamp",
		},
	},
})

// PaginationType represents pagination metadata
var PaginationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Pagination",
	Description: "Pagination metadata for list queries",
	Fields: graphql.Fields{
		"has_next": &graphql.Field{
			Type:        graphql.Int,
			Description: "Current page number",
		},
		"has_previous": &graphql.Field{
			Type:        graphql.Int,
			Description: "Current page number",
		},
		"page": &graphql.Field{
			Type:        graphql.Int,
			Description: "Current page number",
		},
		"size": &graphql.Field{
			Type:        graphql.Int,
			Description: "Number of items per page",
		},
		"skip": &graphql.Field{
			Type:        graphql.Int,
			Description: "Number of items to skip",
		},
		"take_all": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Whether to take all items without pagination",
		},
		"total_count": &graphql.Field{
			Type:        graphql.Int,
			Description: "Total number of items",
		},
		"total_pages": &graphql.Field{
			Type:        graphql.Int,
			Description: "Total number of pages",
		},
	},
})

// UserListType represents a paginated list of users
var UserListType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "UserList",
	Description: "Paginated list of users",
	Fields: graphql.Fields{
		"items": &graphql.Field{
			Type:        graphql.NewList(UserType),
			Description: "List of users",
		},
		"pagination": &graphql.Field{
			Type:        PaginationType,
			Description: "Pagination metadata",
		},
	},
})

// CreateUserInput represents input for creating a user
var CreateUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "CreateUserInput",
	Description: "Input for creating a new user",
	Fields: graphql.InputObjectConfigFieldMap{
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User full name",
		},
		"email": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User email address",
		},
		"phone": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User phone number",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User password",
		},
		"role": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "User role: admin, user, guest",
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

// UpdateUserInput represents input for updating a user
var UpdateUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "UpdateUserInput",
	Description: "Input for updating an existing user",
	Fields: graphql.InputObjectConfigFieldMap{
		"uid": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "User ID to update",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "User full name (optional)",
		},
		"email": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "User email address (optional)",
		},
		"phone": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "User phone number (optional)",
		},
		"role": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "User role (optional)",
		},
		"status": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "User status (optional)",
		},
	},
})
