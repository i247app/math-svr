package services

import (
	"context"
	"fmt"

	repointerface "math-ai.com/math-ai/internal/core/di/repositories"
	svcinterface "math-ai.com/math-ai/internal/core/di/services"
	permDomain "math-ai.com/math-ai/internal/core/domain/permission"
	roleDomain "math-ai.com/math-ai/internal/core/domain/role"
	rpDomain "math-ai.com/math-ai/internal/core/domain/role_permission"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type roleService struct {
	roleRepo           repointerface.IRoleRepository
	permissionRepo     repointerface.IPermissionRepository
	rolePermissionRepo repointerface.IRolePermissionRepository
}

func NewRoleService(
	roleRepo repointerface.IRoleRepository,
	permissionRepo repointerface.IPermissionRepository,
	rolePermissionRepo repointerface.IRolePermissionRepository,
) svcinterface.IRoleService {
	return &roleService{
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
	}
}

// List retrieves a paginated list of roles
func (s *roleService) List(ctx context.Context, params svcinterface.ListRolesServiceParams) ([]*roleDomain.Role, *pagination.Pagination, error) {
	repoParams := repointerface.ListRolesParams{
		Search:        params.Search,
		Page:          params.Page,
		Limit:         params.Limit,
		OrderBy:       params.OrderBy,
		OrderDesc:     params.OrderDesc,
		TakeAll:       params.TakeAll,
		IncludeSystem: true,
	}

	return s.roleRepo.List(ctx, repoParams)
}

// GetByID retrieves a role by ID
func (s *roleService) GetByID(ctx context.Context, id string) (*roleDomain.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %v", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// GetByCode retrieves a role by code
func (s *roleService) GetByCode(ctx context.Context, code string) (*roleDomain.Role, error) {
	role, err := s.roleRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by code: %v", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

// GetPermissions retrieves all permissions for a role (including inherited)
func (s *roleService) GetPermissions(ctx context.Context, roleID string) ([]*permDomain.Permission, error) {
	// Get role hierarchy
	roles, err := s.roleRepo.GetHierarchy(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role hierarchy: %v", err)
	}

	// Collect permissions from all roles in hierarchy
	permissionMap := make(map[string]*permDomain.Permission)
	for _, role := range roles {
		permissions, err := s.permissionRepo.FindByRoleID(ctx, role.ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %v", role.ID(), err)
		}

		for _, perm := range permissions {
			permissionMap[perm.ID()] = perm
		}
	}

	// Convert map to slice
	result := make([]*permDomain.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		result = append(result, perm)
	}

	return result, nil
}

// Create creates a new role
func (s *roleService) Create(ctx context.Context, role *roleDomain.Role) (*roleDomain.Role, error) {
	// Generate ID if not set
	if role.ID() == "" {
		role.GenerateID()
	}

	// Set default status if not set
	if role.Status() == "" {
		role.SetStatus("ACTIVE")
	}

	// Validate parent role exists if specified
	if role.ParentRoleID() != nil && *role.ParentRoleID() != "" {
		parentRole, err := s.roleRepo.FindByID(ctx, *role.ParentRoleID())
		if err != nil {
			return nil, fmt.Errorf("failed to validate parent role: %v", err)
		}
		if parentRole == nil {
			return nil, fmt.Errorf("parent role not found")
		}
	}

	// Create role
	_, err := s.roleRepo.Create(ctx, nil, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %v", err)
	}

	// Return created role
	return s.roleRepo.FindByID(ctx, role.ID())
}

// Update updates an existing role
func (s *roleService) Update(ctx context.Context, role *roleDomain.Role) (*roleDomain.Role, error) {
	// Check if role exists
	existingRole, err := s.roleRepo.FindByID(ctx, role.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to find role: %v", err)
	}
	if existingRole == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Prevent updating system roles
	if existingRole.IsSystemRole() {
		return nil, fmt.Errorf("cannot update system role")
	}

	// Validate parent role if being updated
	if role.ParentRoleID() != nil && *role.ParentRoleID() != "" {
		parentRole, err := s.roleRepo.FindByID(ctx, *role.ParentRoleID())
		if err != nil {
			return nil, fmt.Errorf("failed to validate parent role: %v", err)
		}
		if parentRole == nil {
			return nil, fmt.Errorf("parent role not found")
		}

		// Prevent circular references
		if *role.ParentRoleID() == role.ID() {
			return nil, fmt.Errorf("role cannot be its own parent")
		}
	}

	// Update role
	_, err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %v", err)
	}

	// Return updated role
	return s.roleRepo.FindByID(ctx, role.ID())
}

// Delete soft deletes a role
func (s *roleService) Delete(ctx context.Context, id string) error {
	// Check if role exists
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find role: %v", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Prevent deleting system roles
	if role.IsSystemRole() {
		return fmt.Errorf("cannot delete system role")
	}

	return s.roleRepo.Delete(ctx, id)
}

// ForceDelete permanently deletes a role
func (s *roleService) ForceDelete(ctx context.Context, id string) error {
	// Check if role exists
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find role: %v", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Prevent deleting system roles
	if role.IsSystemRole() {
		return fmt.Errorf("cannot force delete system role")
	}

	return s.roleRepo.ForceDelete(ctx, nil, id)
}

// AssignPermissions assigns multiple permissions to a role
func (s *roleService) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Verify role exists
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to find role: %v", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Create role-permission mappings
	rolePermissions := make([]*rpDomain.RolePermission, 0, len(permissionIDs))
	for _, permID := range permissionIDs {
		// Verify permission exists
		perm, err := s.permissionRepo.FindByID(ctx, permID)
		if err != nil {
			return fmt.Errorf("failed to find permission %s: %v", permID, err)
		}
		if perm == nil {
			return fmt.Errorf("permission %s not found", permID)
		}

		rp := rpDomain.NewRolePermissionDomain()
		rp.GenerateID()
		rp.SetRoleID(roleID)
		rp.SetPermissionID(permID)
		rolePermissions = append(rolePermissions, rp)
	}

	// Bulk assign permissions
	return s.rolePermissionRepo.BulkAssignPermissions(ctx, nil, rolePermissions)
}

// RevokePermissions removes multiple permissions from a role
func (s *roleService) RevokePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Verify role exists
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to find role: %v", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Revoke each permission
	for _, permID := range permissionIDs {
		err := s.rolePermissionRepo.RevokePermission(ctx, nil, roleID, permID)
		if err != nil {
			return fmt.Errorf("failed to revoke permission %s: %v", permID, err)
		}
	}

	return nil
}
