package dto

import (
	domain "math-ai.com/math-ai/internal/core/domain/role"
	permDomain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type RoleResponse struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Code         string        `json:"code"`
	Description  *string       `json:"description"`
	ParentRoleID *string       `json:"parent_role_id"`
	IsSystemRole bool          `json:"is_system_role"`
	Status       string        `json:"status"`
	DisplayOrder int8          `json:"display_order"`
	CreatedAt    time.MathTime `json:"created_at"`
	ModifiedAt   time.MathTime `json:"modified_at"`
}

type RoleWithPermissionsResponse struct {
	Role        RoleResponse        `json:"role"`
	Permissions []PermissionResponse `json:"permissions"`
}

type GetRoleResponse struct {
	Role *RoleResponse `json:"role"`
}

type ListRoleRequest struct {
	Search        string `json:"search,omitempty" form:"search"`
	Page          int64  `json:"page" form:"page"`
	Limit         int64  `json:"size" form:"size"`
	OrderBy       string `json:"order_by" form:"order_by"`
	OrderDesc     bool   `json:"order_desc" form:"order_desc"`
	TakeAll       bool   `json:"take_all" form:"take_all"`
	IncludeSystem bool   `json:"include_system" form:"include_system"`
}

type ListRoleResponse struct {
	Items      []*RoleResponse        `json:"roles"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateRoleRequest struct {
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	Description  *string `json:"description"`
	ParentRoleID *string `json:"parent_role_id"`
	DisplayOrder int8    `json:"display_order"`
}

type CreateRoleResponse struct {
	Role *RoleResponse `json:"role"`
}

type UpdateRoleRequest struct {
	ID           string        `json:"id"`
	Name         *string       `json:"name,omitempty"`
	Code         *string       `json:"code,omitempty"`
	Description  *string       `json:"description,omitempty"`
	ParentRoleID *string       `json:"parent_role_id,omitempty"`
	Status       *enum.EStatus `json:"status,omitempty"`
	DisplayOrder *int8         `json:"display_order,omitempty"`
}

type UpdateRoleResponse struct {
	Role *RoleResponse `json:"role"`
}

type DeleteRoleRequest struct {
	ID string `json:"id"`
}

type AssignPermissionsRequest struct {
	RoleID        string   `json:"role_id"`
	PermissionIDs []string `json:"permission_ids"`
}

type RevokePermissionsRequest struct {
	RoleID        string   `json:"role_id"`
	PermissionIDs []string `json:"permission_ids"`
}

type GetRolePermissionsResponse struct {
	RoleID      string               `json:"role_id"`
	Permissions []PermissionResponse `json:"permissions"`
}

// BuildRoleDomainForCreate builds a Role domain from CreateRoleRequest
func BuildRoleDomainForCreate(req *CreateRoleRequest) *domain.Role {
	roleDomain := domain.NewRoleDomain()
	roleDomain.GenerateID()
	roleDomain.SetName(req.Name)
	roleDomain.SetCode(req.Code)
	roleDomain.SetDescription(req.Description)
	roleDomain.SetParentRoleID(req.ParentRoleID)
	roleDomain.SetIsSystemRole(false) // Custom roles are never system roles
	roleDomain.SetStatus(string(enum.StatusActive))
	roleDomain.SetDisplayOrder(req.DisplayOrder)

	return roleDomain
}

// BuildRoleDomainForUpdate builds a Role domain from UpdateRoleRequest
func BuildRoleDomainForUpdate(req *UpdateRoleRequest) *domain.Role {
	roleDomain := domain.NewRoleDomain()
	roleDomain.SetID(req.ID)

	if req.Name != nil {
		roleDomain.SetName(*req.Name)
	}

	if req.Code != nil {
		roleDomain.SetCode(*req.Code)
	}

	if req.Description != nil {
		roleDomain.SetDescription(req.Description)
	}

	if req.ParentRoleID != nil {
		roleDomain.SetParentRoleID(req.ParentRoleID)
	}

	if req.Status != nil {
		roleDomain.SetStatus(string(*req.Status))
	}

	if req.DisplayOrder != nil {
		roleDomain.SetDisplayOrder(*req.DisplayOrder)
	}

	return roleDomain
}

// RoleResponseFromDomain converts Role domain to RoleResponse
func RoleResponseFromDomain(r *domain.Role) RoleResponse {
	return RoleResponse{
		ID:           r.ID(),
		Name:         r.Name(),
		Code:         r.Code(),
		Description:  r.Description(),
		ParentRoleID: r.ParentRoleID(),
		IsSystemRole: r.IsSystemRole(),
		Status:       r.Status(),
		DisplayOrder: r.DisplayOrder(),
		CreatedAt:    r.CreatedAt(),
		ModifiedAt:   r.ModifiedAt(),
	}
}

// RoleResponseListFromDomain converts Role domain list to RoleResponse list
func RoleResponseListFromDomain(roles []*domain.Role) []*RoleResponse {
	responses := make([]*RoleResponse, len(roles))
	for i, r := range roles {
		resp := RoleResponseFromDomain(r)
		responses[i] = &resp
	}
	return responses
}

// RoleWithPermissionsFromDomain converts Role and Permissions to RoleWithPermissionsResponse
func RoleWithPermissionsFromDomain(role *domain.Role, permissions []*permDomain.Permission) RoleWithPermissionsResponse {
	return RoleWithPermissionsResponse{
		Role:        RoleResponseFromDomain(role),
		Permissions: PermissionResponseListFromDomain(permissions),
	}
}
