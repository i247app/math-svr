package dto

import (
	domain "math-ai.com/math-ai/internal/core/domain/permission"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type PermissionResponse struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  *string       `json:"description"`
	HTTPMethod   string        `json:"http_method"`
	EndpointPath string        `json:"endpoint_path"`
	Resource     *string       `json:"resource"`
	Action       *string       `json:"action"`
	Status       string        `json:"status"`
	CreatedAt    time.MathTime `json:"created_at"`
	ModifiedAt   time.MathTime `json:"modified_at"`
}

type GetPermissionResponse struct {
	Permission *PermissionResponse `json:"permission"`
}

type ListPermissionRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Resource  string `json:"resource,omitempty" form:"resource"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListPermissionResponse struct {
	Items      []PermissionResponse   `json:"permissions"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreatePermissionRequest struct {
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	HTTPMethod   string  `json:"http_method"`
	EndpointPath string  `json:"endpoint_path"`
	Resource     *string `json:"resource"`
	Action       *string `json:"action"`
}

type CreatePermissionResponse struct {
	Permission *PermissionResponse `json:"permission"`
}

type UpdatePermissionRequest struct {
	ID           string        `json:"id"`
	Name         *string       `json:"name,omitempty"`
	Description  *string       `json:"description,omitempty"`
	HTTPMethod   *string       `json:"http_method,omitempty"`
	EndpointPath *string       `json:"endpoint_path,omitempty"`
	Resource     *string       `json:"resource,omitempty"`
	Action       *string       `json:"action,omitempty"`
	Status       *enum.EStatus `json:"status,omitempty"`
}

type UpdatePermissionResponse struct {
	Permission *PermissionResponse `json:"permission"`
}

type DeletePermissionRequest struct {
	ID string `json:"id"`
}

// BuildPermissionDomainForCreate builds a Permission domain from CreatePermissionRequest
func BuildPermissionDomainForCreate(req *CreatePermissionRequest) *domain.Permission {
	permDomain := domain.NewPermissionDomain()
	permDomain.GenerateID()
	permDomain.SetName(req.Name)
	permDomain.SetDescription(req.Description)
	permDomain.SetHTTPMethod(req.HTTPMethod)
	permDomain.SetEndpointPath(req.EndpointPath)
	permDomain.SetResource(req.Resource)
	permDomain.SetAction(req.Action)
	permDomain.SetStatus(string(enum.StatusActive))

	return permDomain
}

// BuildPermissionDomainForUpdate builds a Permission domain from UpdatePermissionRequest
func BuildPermissionDomainForUpdate(req *UpdatePermissionRequest) *domain.Permission {
	permDomain := domain.NewPermissionDomain()
	permDomain.SetID(req.ID)

	if req.Name != nil {
		permDomain.SetName(*req.Name)
	}

	if req.Description != nil {
		permDomain.SetDescription(req.Description)
	}

	if req.HTTPMethod != nil {
		permDomain.SetHTTPMethod(*req.HTTPMethod)
	}

	if req.EndpointPath != nil {
		permDomain.SetEndpointPath(*req.EndpointPath)
	}

	if req.Resource != nil {
		permDomain.SetResource(req.Resource)
	}

	if req.Action != nil {
		permDomain.SetAction(req.Action)
	}

	if req.Status != nil {
		permDomain.SetStatus(string(*req.Status))
	}

	return permDomain
}

// PermissionResponseFromDomain converts Permission domain to PermissionResponse
func PermissionResponseFromDomain(p *domain.Permission) PermissionResponse {
	return PermissionResponse{
		ID:           p.ID(),
		Name:         p.Name(),
		Description:  p.Description(),
		HTTPMethod:   p.HTTPMethod(),
		EndpointPath: p.EndpointPath(),
		Resource:     p.Resource(),
		Action:       p.Action(),
		Status:       p.Status(),
		CreatedAt:    p.CreatedAt(),
		ModifiedAt:   p.ModifiedAt(),
	}
}

// PermissionResponseListFromDomain converts Permission domain list to PermissionResponse list
func PermissionResponseListFromDomain(permissions []*domain.Permission) []PermissionResponse {
	responses := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = PermissionResponseFromDomain(p)
	}
	return responses
}
