package types

import (
	"github.com/graphql-go/graphql"
)

// GradeType represents the Grade GraphQL type
var GradeType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Grade",
	Description: "Educational grade level",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Unique identifier (UUID)",
		},
		"label": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Grade label (e.g., 'Grade 1', 'Grade 2')",
		},
		"description": &graphql.Field{
			Type:        graphql.String,
			Description: "Grade description",
		},
		"icon_url": &graphql.Field{
			Type:        graphql.String,
			Description: "URL to grade icon/image",
		},
		"status": &graphql.Field{
			Type:        graphql.String,
			Description: "Grade status: ACTIVE, INACTIVE",
		},
		"display_order": &graphql.Field{
			Type:        graphql.Int,
			Description: "Display order for sorting",
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

// GradeListType represents a list of grades
var GradeListType = graphql.NewList(GradeType)

// CreateGradeInput represents input for creating a grade
var CreateGradeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "CreateGradeInput",
	Description: "Input for creating a new grade",
	Fields: graphql.InputObjectConfigFieldMap{
		"label": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Grade label (e.g., 'Grade 1')",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Grade description",
		},
		"icon_url": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "URL to grade icon/image",
		},
		"display_order": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Display order for sorting",
		},
	},
})

// UpdateGradeInput represents input for updating a grade
var UpdateGradeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "UpdateGradeInput",
	Description: "Input for updating an existing grade",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Grade ID to update",
		},
		"label": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Grade label (optional)",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Grade description (optional)",
		},
		"icon_url": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "URL to grade icon/image (optional)",
		},
		"status": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Grade status (optional)",
		},
		"display_order": &graphql.InputObjectFieldConfig{
			Type:        graphql.Int,
			Description: "Display order (optional)",
		},
	},
})
