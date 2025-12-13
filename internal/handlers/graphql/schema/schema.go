package schema

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/app/services"
	"math-ai.com/math-ai/internal/handlers/graphql/resolvers"
	"math-ai.com/math-ai/internal/handlers/graphql/schema/types"
)

// BuildSchema creates and returns the GraphQL schema
func BuildSchema(serviceContainer *services.ServiceContainer) (graphql.Schema, error) {
	// Initialize resolvers
	userResolver := resolvers.NewUserResolver(serviceContainer.UserService)
	gradeResolver := resolvers.NewGradeResolver(serviceContainer.GradeService)
	loginResolver := resolvers.NewLoginResolver(serviceContainer.AuthService)
	chatboxResolver := resolvers.NewChatBoxResolver(serviceContainer.UserQuizPracticesService)
	profileResolver := resolvers.NewProfileResolver(serviceContainer.ProfileService)

	// Define root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Query",
		Description: "Root query for the Math-AI GraphQL API",
		Fields: graphql.Fields{
			// ============================================================
			// USER QUERIES
			// ============================================================
			"user": &graphql.Field{
				Type:        types.UserType,
				Description: "Get a user by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "User ID (UUID)",
					},
				},
				Resolve: userResolver.GetUser,
			},
			"users": &graphql.Field{
				Type:        types.UserListType,
				Description: "List users with pagination",
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						Description:  "Page number (default: 1)",
						DefaultValue: 1,
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						Description:  "Items per page (default: 10)",
						DefaultValue: 10,
					},
					"search": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Search term for filtering users",
					},
					"order_by": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Field to order by",
					},
					"order_desc": &graphql.ArgumentConfig{
						Type:         graphql.Boolean,
						Description:  "Order descending (default: false)",
						DefaultValue: false,
					},
					"take_all": &graphql.ArgumentConfig{
						Type:         graphql.Boolean,
						Description:  "Retrieve all users without pagination (default: false)",
						DefaultValue: false,
					},
				},
				Resolve: userResolver.ListUsers,
			},

			// ============================================================
			// GRADE QUERIES
			// ============================================================
			"grade": &graphql.Field{
				Type:        types.GradeType,
				Description: "Get a grade by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Grade ID (UUID)",
					},
				},
				Resolve: gradeResolver.GetGrade,
			},
			"gradeByLabel": &graphql.Field{
				Type:        types.GradeType,
				Description: "Get a grade by label",
				Args: graphql.FieldConfigArgument{
					"label": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Grade label (e.g., 'Grade 1')",
					},
				},
				Resolve: gradeResolver.GetGradeByLabel,
			},
			"grades": &graphql.Field{
				Type:        types.GradeListType,
				Description: "List all grades",
				Resolve:     gradeResolver.ListGrades,
			},

			// ============================================================
			// PROFILE QUERIES
			// ============================================================
			"profile": &graphql.Field{
				Type:        types.ProfileType,
				Description: "Fetch user profile by UID",
				Args: graphql.FieldConfigArgument{
					"uid": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "User ID",
					},
				},
				Resolve: profileResolver.FetchProfile,
			},
		},
	})

	// Define root mutation
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Mutation",
		Description: "Root mutation for the Math-AI GraphQL API",
		Fields: graphql.Fields{
			// ============================================================
			// AUTHENTICATION MUTATIONS
			// ============================================================
			"login": &graphql.Field{
				Type:        types.LoginResponseType,
				Description: "User login",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.LoginInput),
						Description: "Login credentials",
					},
				},
				Resolve: loginResolver.Login,
			},

			// ============================================================
			// USER MUTATIONS
			// ============================================================
			"createUser": &graphql.Field{
				Type:        types.UserType,
				Description: "Create a new user",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.CreateUserInput),
						Description: "User creation input",
					},
				},
				Resolve: userResolver.CreateUser,
			},
			"updateUser": &graphql.Field{
				Type:        types.UserType,
				Description: "Update an existing user",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.UpdateUserInput),
						Description: "User update input",
					},
				},
				Resolve: userResolver.UpdateUser,
			},
			"deleteUser": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Soft delete a user",
				Args: graphql.FieldConfigArgument{
					"uid": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "User ID to delete",
					},
				},
				Resolve: userResolver.DeleteUser,
			},
			"forceDeleteUser": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Permanently delete a user",
				Args: graphql.FieldConfigArgument{
					"uid": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "User ID to permanently delete",
					},
				},
				Resolve: userResolver.ForceDeleteUser,
			},

			// ============================================================
			// QUIZ/CHATBOX MUTATIONS
			// ============================================================
			"generateQuiz": &graphql.Field{
				Type:        types.QuizResponseType,
				Description: "Generate a new quiz for a user",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.GenerateQuizInput),
						Description: "Quiz generation input",
					},
				},
				Resolve: chatboxResolver.GenerateQuiz,
			},
			"submitQuiz": &graphql.Field{
				Type:        types.QuizEvaluationResponseType,
				Description: "Submit quiz answers for evaluation",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.SubmitQuizInput),
						Description: "Quiz answers input",
					},
				},
				Resolve: chatboxResolver.SubmitQuiz,
			},
			"generateQuizPractice": &graphql.Field{
				Type:        types.QuizResponseType,
				Description: "Generate a practice quiz based on previous performance",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.GenerateQuizPracticeInput),
						Description: "Practice quiz generation input",
					},
				},
				Resolve: chatboxResolver.GenerateQuizPractice,
			},

			// ============================================================
			// GRADE MUTATIONS
			// ============================================================
			"createGrade": &graphql.Field{
				Type:        types.GradeType,
				Description: "Create a new grade",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.CreateGradeInput),
						Description: "Grade creation input",
					},
				},
				Resolve: gradeResolver.CreateGrade,
			},
			"updateGrade": &graphql.Field{
				Type:        types.GradeType,
				Description: "Update an existing grade",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.UpdateGradeInput),
						Description: "Grade update input",
					},
				},
				Resolve: gradeResolver.UpdateGrade,
			},
			"deleteGrade": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Soft delete a grade",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Grade ID to delete",
					},
				},
				Resolve: gradeResolver.DeleteGrade,
			},
			"forceDeleteGrade": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Permanently delete a grade",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.String),
						Description: "Grade ID to permanently delete",
					},
				},
				Resolve: gradeResolver.ForceDeleteGrade,
			},

			// ============================================================
			// PROFILE MUTATIONS
			// ============================================================
			"createProfile": &graphql.Field{
				Type:        types.ProfileType,
				Description: "Create a user profile",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.CreateProfileInput),
						Description: "Profile creation input",
					},
				},
				Resolve: profileResolver.CreateProfile,
			},
			"updateProfile": &graphql.Field{
				Type:        types.ProfileType,
				Description: "Update a user profile",
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(types.UpdateProfileInput),
						Description: "Profile update input",
					},
				},
				Resolve: profileResolver.UpdateProfile,
			},
		},
	})

	// Build and return the schema
	schemaConfig := graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	}

	return graphql.NewSchema(schemaConfig)
}
