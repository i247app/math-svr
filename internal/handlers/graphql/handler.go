package graphql

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/app/services"
	"math-ai.com/math-ai/internal/handlers/graphql/schema"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type GraphQLHandler struct {
	schema graphql.Schema
}

// NewGraphQLHandler creates a new GraphQL handler with the provided service container
func NewGraphQLHandler(serviceContainer *services.ServiceContainer) (*GraphQLHandler, error) {
	gqlSchema, err := schema.BuildSchema(serviceContainer)
	if err != nil {
		return nil, err
	}

	return &GraphQLHandler{
		schema: gqlSchema,
	}, nil
}

// ServeHTTP handles GraphQL HTTP requests
func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var requestBody struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Errorf("Failed to decode GraphQL request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute the GraphQL query
	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  requestBody.Query,
		VariableValues: requestBody.Variables,
		OperationName:  requestBody.OperationName,
		Context:        r.Context(),
	})

	// Log errors if any
	if len(result.Errors) > 0 {
		logger.Errorf("GraphQL errors: %v", result.Errors)
	}

	// // Set response headers
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)

	// // Write the response
	// if err := json.NewEncoder(w).Encode(result); err != nil {
	// 	logger.Errorf("Failed to encode GraphQL response: %v", err)
	// }

	response.WriteJson(w, r.Context(), result, nil, status.OK)
}
