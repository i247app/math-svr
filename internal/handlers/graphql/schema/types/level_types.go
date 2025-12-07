package types

import (
	"github.com/graphql-go/graphql"
)

// LevelType represents the Level GraphQL type
var LevelType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Level",
	Description: "Difficulty/proficiency level",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Unique identifier (UUID)",
		},
		"label": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Level label (e.g., 'Basic', 'Advanced')",
		},
		"description": &graphql.Field{
			Type:        graphql.String,
			Description: "Level description",
		},
		"image_key": &graphql.Field{
			Type:        graphql.String,
			Description: "URL to level icon/image",
		},
		"status": &graphql.Field{
			Type:        graphql.String,
			Description: "Level status: ACTIVE, INACTIVE",
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

// LevelListType represents a list of levels
var LevelListType = graphql.NewList(LevelType)

// CreateLevelInput represents input for creating a level
var CreateLevelInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "CreateLevelInput",
	Description: "Input for creating a new level",
	Fields: graphql.InputObjectConfigFieldMap{
		"label": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Level label (e.g., 'Basic')",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Level description",
		},
		"image_key": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "URL to level icon/image",
		},
		"display_order": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Int),
			Description: "Display order for sorting",
		},
	},
})

// UpdateLevelInput represents input for updating a level
var UpdateLevelInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "UpdateLevelInput",
	Description: "Input for updating an existing level",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Level ID to update",
		},
		"label": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Level label (optional)",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Level description (optional)",
		},
		"image_key": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "URL to level icon/image (optional)",
		},
		"status": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Level status (optional)",
		},
		"display_order": &graphql.InputObjectFieldConfig{
			Type:        graphql.Int,
			Description: "Display order (optional)",
		},
	},
})
