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

type PermissionController struct {
	appResources      *resources.AppResource
	permissionService diSvc.IPermissionService
}

func NewPermissionController(appResources *resources.AppResource, permissionService diSvc.IPermissionService) *PermissionController {
	return &PermissionController{
		appResources:      appResources,
		permissionService: permissionService,
	}
}

// HandlerListPermissions - GET /permissions/list
func (c *PermissionController) HandlerListPermissions(w http.ResponseWriter, r *http.Request) {
	var req dto.ListPermissionRequest

	// Parse query parameters
	if err := parseQueryParams(r, &req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Convert to service params
	params := diSvc.ListPermissionsServiceParams{
		Search:    req.Search,
		Resource:  req.Resource,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	permissions, pagination, err := c.permissionService.List(r.Context(), params)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	res := &dto.ListPermissionResponse{
		Items:      dto.PermissionResponseListFromDomain(permissions),
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerGetPermission - GET /permissions/{id}
func (c *PermissionController) HandlerGetPermission(w http.ResponseWriter, r *http.Request) {
	permissionID := r.PathValue("id")

	permission, err := c.permissionService.GetByID(r.Context(), permissionID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	permResp := dto.PermissionResponseFromDomain(permission)
	res := &dto.GetPermissionResponse{
		Permission: &permResp,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerCreatePermission - POST /permissions/create
func (c *PermissionController) HandlerCreatePermission(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Build domain object
	permDomain := dto.BuildPermissionDomainForCreate(&req)

	// Create permission
	createdPermission, err := c.permissionService.Create(r.Context(), permDomain)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	permResp := dto.PermissionResponseFromDomain(createdPermission)
	res := &dto.CreatePermissionResponse{
		Permission: &permResp,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerUpdatePermission - POST /permissions/update
func (c *PermissionController) HandlerUpdatePermission(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	// Build domain object
	permDomain := dto.BuildPermissionDomainForUpdate(&req)

	// Update permission
	updatedPermission, err := c.permissionService.Update(r.Context(), permDomain)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	permResp := dto.PermissionResponseFromDomain(updatedPermission)
	res := &dto.UpdatePermissionResponse{
		Permission: &permResp,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

// HandlerDeletePermission - POST /permissions/delete
func (c *PermissionController) HandlerDeletePermission(w http.ResponseWriter, r *http.Request) {
	var req dto.DeletePermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	if err := c.permissionService.Delete(r.Context(), req.ID); err != nil {
		response.WriteJson(w, r.Context(), nil, err, status.FAIL)
		return
	}

	response.WriteJson(w, r.Context(), map[string]string{"message": "Permission deleted successfully"}, nil, status.OK)
}
