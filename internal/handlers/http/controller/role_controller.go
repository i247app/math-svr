package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type RoleController struct {
	appResources *resources.AppResource
	roleService  diSvc.IRoleService
}

func NewRoleController(appResources *resources.AppResource, roleService diSvc.IRoleService) *RoleController {
	return &RoleController{
		appResources: appResources,
		roleService:  roleService,
	}
}

// HandlerListRoles - GET /roles/list
func (c *RoleController) HandlerListRoles(w http.ResponseWriter, r *http.Request) {
	var req dto.ListRoleRequest

	// Parse query parameters
	if err := parseQueryParams(r, &req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Convert to service params
	params := diSvc.ListRolesServiceParams{
		Search:        req.Search,
		Page:          req.Page,
		Limit:         req.Limit,
		OrderBy:       req.OrderBy,
		OrderDesc:     req.OrderDesc,
		TakeAll:       req.TakeAll,
		IncludeSystem: req.IncludeSystem,
	}

	roles, pagination, err := c.roleService.List(r.Context(), params)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	res := &dto.ListRoleResponse{
		Items:      dto.RoleResponseListFromDomain(roles),
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerGetRole - GET /roles/{id}
func (c *RoleController) HandlerGetRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")

	role, err := c.roleService.GetByID(r.Context(), roleID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	roleResp := dto.RoleResponseFromDomain(role)
	res := &dto.GetRoleResponse{
		Role: &roleResp,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerGetRolePermissions - GET /roles/{id}/permissions
func (c *RoleController) HandlerGetRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")

	permissions, err := c.roleService.GetPermissions(r.Context(), roleID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	res := &dto.GetRolePermissionsResponse{
		RoleID:      roleID,
		Permissions: dto.PermissionResponseListFromDomain(permissions),
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerCreateRole - POST /roles/create
func (c *RoleController) HandlerCreateRole(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Build domain object
	roleDomain := dto.BuildRoleDomainForCreate(&req)

	// Create role
	createdRole, err := c.roleService.Create(r.Context(), roleDomain)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	roleResp := dto.RoleResponseFromDomain(createdRole)
	res := &dto.CreateRoleResponse{
		Role: &roleResp,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerUpdateRole - POST /roles/update
func (c *RoleController) HandlerUpdateRole(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Build domain object
	roleDomain := dto.BuildRoleDomainForUpdate(&req)

	// Update role
	updatedRole, err := c.roleService.Update(r.Context(), roleDomain)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	roleResp := dto.RoleResponseFromDomain(updatedRole)
	res := &dto.UpdateRoleResponse{
		Role: &roleResp,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerDeleteRole - POST /roles/delete
func (c *RoleController) HandlerDeleteRole(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	if err := c.roleService.Delete(r.Context(), req.ID); err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	response.WriteJson(w, r.Context(), map[string]string{"message": "Role deleted successfully"}, nil, status.OK)
}

// HandlerAssignPermissions - POST /roles/assign-permissions
func (c *RoleController) HandlerAssignPermissions(w http.ResponseWriter, r *http.Request) {
	var req dto.AssignPermissionsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	if err := c.roleService.AssignPermissions(r.Context(), req.RoleID, req.PermissionIDs); err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	response.WriteJson(w, r.Context(), map[string]string{"message": "Permissions assigned successfully"}, nil, status.OK)
}

// HandlerRevokePermissions - POST /roles/revoke-permissions
func (c *RoleController) HandlerRevokePermissions(w http.ResponseWriter, r *http.Request) {
	var req dto.RevokePermissionsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	if err := c.roleService.RevokePermissions(r.Context(), req.RoleID, req.PermissionIDs); err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	response.WriteJson(w, r.Context(), map[string]string{"message": "Permissions revoked successfully"}, nil, status.OK)
}

// Helper function to parse query parameters
func parseQueryParams(r *http.Request, req interface{}) error {
	// This is a simplified version - you may want to use a proper query parser
	// For now, return nil as the struct will use default values
	return nil
}
