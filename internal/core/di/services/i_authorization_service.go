package di

import (
	"context"

	domain "math-ai.com/math-ai/internal/core/domain/permission"
)

type IAuthorizationService interface {
	// HasPermission checks if a user has permission for a specific endpoint
	HasPermission(ctx context.Context, userID, httpMethod, endpointPath string) (bool, error)

	// GetUserPermissions retrieves all permissions for a user (including role hierarchy)
	GetUserPermissions(ctx context.Context, userID string) ([]*domain.Permission, error)

	// MatchesEndpoint checks if a permission matches a given endpoint
	MatchesEndpoint(permission *domain.Permission, httpMethod, requestPath string) bool
}
