package services

import (
	"context"
	"fmt"

	repointerface "math-ai.com/math-ai/internal/core/di/repositories"
	svcinterface "math-ai.com/math-ai/internal/core/di/services"
	domain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type permissionService struct {
	permissionRepo repointerface.IPermissionRepository
}

func NewPermissionService(
	permissionRepo repointerface.IPermissionRepository,
) svcinterface.IPermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
	}
}

// List retrieves a paginated list of permissions
func (s *permissionService) List(ctx context.Context, params svcinterface.ListPermissionsServiceParams) ([]*domain.Permission, *pagination.Pagination, error) {
	repoParams := repointerface.ListPermissionsParams{
		Search:    params.Search,
		Resource:  params.Resource,
		Page:      params.Page,
		Limit:     params.Limit,
		OrderBy:   params.OrderBy,
		OrderDesc: params.OrderDesc,
		TakeAll:   params.TakeAll,
	}

	return s.permissionRepo.List(ctx, repoParams)
}

// GetByID retrieves a permission by ID
func (s *permissionService) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %v", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("permission not found")
	}
	return permission, nil
}

// GetByName retrieves a permission by name
func (s *permissionService) GetByName(ctx context.Context, name string) (*domain.Permission, error) {
	permission, err := s.permissionRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by name: %v", err)
	}
	if permission == nil {
		return nil, fmt.Errorf("permission not found")
	}
	return permission, nil
}

// GetByResource retrieves all permissions for a resource
func (s *permissionService) GetByResource(ctx context.Context, resource string) ([]*domain.Permission, error) {
	permissions, err := s.permissionRepo.FindByResource(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by resource: %v", err)
	}
	return permissions, nil
}

// Create creates a new permission
func (s *permissionService) Create(ctx context.Context, permission *domain.Permission) (*domain.Permission, error) {
	// Generate ID if not set
	if permission.ID() == "" {
		permission.GenerateID()
	}

	// Set default status if not set
	if permission.Status() == "" {
		permission.SetStatus("ACTIVE")
	}

	// Validate that method and endpoint are provided
	if permission.HTTPMethod() == "" {
		return nil, fmt.Errorf("http_method is required")
	}
	if permission.EndpointPath() == "" {
		return nil, fmt.Errorf("endpoint_path is required")
	}

	// Check if permission already exists for this endpoint
	existing, err := s.permissionRepo.FindByEndpoint(ctx, permission.HTTPMethod(), permission.EndpointPath())
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permission: %v", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("permission already exists for %s %s", permission.HTTPMethod(), permission.EndpointPath())
	}

	// Create permission
	_, err = s.permissionRepo.Create(ctx, nil, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %v", err)
	}

	// Return created permission
	return s.permissionRepo.FindByID(ctx, permission.ID())
}

// Update updates an existing permission
func (s *permissionService) Update(ctx context.Context, permission *domain.Permission) (*domain.Permission, error) {
	// Check if permission exists
	existingPermission, err := s.permissionRepo.FindByID(ctx, permission.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to find permission: %v", err)
	}
	if existingPermission == nil {
		return nil, fmt.Errorf("permission not found")
	}

	// Update permission
	_, err = s.permissionRepo.Update(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission: %v", err)
	}

	// Return updated permission
	return s.permissionRepo.FindByID(ctx, permission.ID())
}

// Delete soft deletes a permission
func (s *permissionService) Delete(ctx context.Context, id string) error {
	// Check if permission exists
	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find permission: %v", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	return s.permissionRepo.Delete(ctx, id)
}

// ForceDelete permanently deletes a permission
func (s *permissionService) ForceDelete(ctx context.Context, id string) error {
	// Check if permission exists
	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find permission: %v", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	return s.permissionRepo.ForceDelete(ctx, nil, id)
}
