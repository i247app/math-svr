package services

import (
	"context"
	"fmt"
	"strings"

	repointerface "math-ai.com/math-ai/internal/core/di/repositories"
	svcinterface "math-ai.com/math-ai/internal/core/di/services"
	domain "math-ai.com/math-ai/internal/core/domain/permission"
)

type authorizationService struct {
	userRepo       repointerface.IUserRepository
	roleRepo       repointerface.IRoleRepository
	permissionRepo repointerface.IPermissionRepository
}

func NewAuthorizationService(
	userRepo repointerface.IUserRepository,
	roleRepo repointerface.IRoleRepository,
	permissionRepo repointerface.IPermissionRepository,
) svcinterface.IAuthorizationService {
	return &authorizationService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// HasPermission checks if a user has permission for a specific endpoint
func (s *authorizationService) HasPermission(ctx context.Context, userID, httpMethod, endpointPath string) (bool, error) {
	// Get user permissions (including role hierarchy)
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %v", err)
	}

	// Check if any permission matches the endpoint
	for _, permission := range permissions {
		if s.MatchesEndpoint(permission, httpMethod, endpointPath) {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions retrieves all permissions for a user (including role hierarchy)
func (s *authorizationService) GetUserPermissions(ctx context.Context, userID string) ([]*domain.Permission, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get user's role ID (try role_id first, fallback to role string)
	roleID := user.RoleID()
	if roleID == nil || *roleID == "" {
		// Fallback: map role string to role ID
		roleCode := user.Role()
		if roleCode == "" {
			return []*domain.Permission{}, nil
		}

		role, err := s.roleRepo.FindByCode(ctx, roleCode)
		if err != nil {
			return nil, fmt.Errorf("failed to find role by code: %v", err)
		}
		if role == nil {
			return []*domain.Permission{}, nil
		}
		roleIDStr := role.ID()
		roleID = &roleIDStr
	}

	// Get role hierarchy (current role + all parent roles)
	roles, err := s.roleRepo.GetHierarchy(ctx, *roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role hierarchy: %v", err)
	}

	// Collect all permissions from all roles in hierarchy
	permissionMap := make(map[string]*domain.Permission) // Use map to deduplicate
	for _, role := range roles {
		rolePermissions, err := s.permissionRepo.FindByRoleID(ctx, role.ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %v", role.ID(), err)
		}

		for _, perm := range rolePermissions {
			permissionMap[perm.ID()] = perm
		}
	}

	// Convert map to slice
	permissions := make([]*domain.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// MatchesEndpoint checks if a permission matches a given endpoint
func (s *authorizationService) MatchesEndpoint(permission *domain.Permission, httpMethod, requestPath string) bool {
	// Check HTTP method
	if !s.matchesMethod(permission.HTTPMethod(), httpMethod) {
		return false
	}

	// Check endpoint path
	return s.matchesPath(permission.EndpointPath(), requestPath)
}

// matchesMethod checks if permission method matches request method
func (s *authorizationService) matchesMethod(permMethod, reqMethod string) bool {
	// Wildcard matches all methods
	if permMethod == "*" {
		return true
	}

	// Exact match (case insensitive)
	return strings.EqualFold(permMethod, reqMethod)
}

// matchesPath checks if permission path matches request path
func (s *authorizationService) matchesPath(permPath, reqPath string) bool {
	// Exact match
	if permPath == reqPath {
		return true
	}

	// Wildcard match (e.g., /users/* matches /users/123)
	if strings.HasSuffix(permPath, "/*") {
		prefix := strings.TrimSuffix(permPath, "/*")
		// Check if request path starts with the prefix
		if strings.HasPrefix(reqPath, prefix+"/") {
			return true
		}
		// Also match exact prefix (e.g., /users/* matches /users)
		if reqPath == prefix {
			return true
		}
	}

	return false
}
