package role

import (
	"io"

	domain "math-ai.com/math-ai/internal/domain/role"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// RoleResponse.RoleImageUrl is a short-lived presigned URL derived from
// RoleImageKey at the service edge; nil when there is no image or storage
// is disabled.
type RoleResponse struct {
	RoleID       int64   `json:"role_id"`
	RoleCode     string  `json:"role_code"`
	RoleName     string  `json:"role_name"`
	Description  *string `json:"description,omitempty"`
	RoleImageKey *string `json:"role_image_key,omitempty"`
	RoleImageUrl *string `json:"role_image_url"`
	Note         *string `json:"note,omitempty"`
	RoleStatus   *string `json:"role_status,omitempty"`
	CreateDt     string  `json:"create_dt"`
	ModifyDt     string  `json:"modify_dt"`
}

type CreateRoleReq struct {
	RoleCode    string  `json:"role_code"`
	RoleName    string  `json:"role_name"`
	Description *string `json:"description,omitempty"`
	Note        *string `json:"note,omitempty"`
	// RoleStatus optionally overrides the default ACTIVE state on create.
	RoleStatus *string `json:"role_status,omitempty"`

	// Multipart only: the "role_image" file part, uploaded to S3 by the
	// service; the resulting key is what lands on ma_roles.role_image_key.
	RoleImageFile        io.Reader `json:"-"`
	RoleImageFilename    string    `json:"-"`
	RoleImageContentType string    `json:"-"`
}

type CreateRoleRes struct {
	Role *RoleResponse `json:"role"`
}

// UpdateRoleReq: a nil field is left unchanged. role_code cannot be changed.
// A new role_image file replaces the current image; RemoveRoleImage deletes
// it (the two are mutually exclusive).
type UpdateRoleReq struct {
	RoleID          int64   `json:"role_id"`
	RoleName        *string `json:"role_name,omitempty"`
	Description     *string `json:"description,omitempty"`
	Note            *string `json:"note,omitempty"`
	RoleStatus      *string `json:"role_status,omitempty"`
	RemoveRoleImage bool    `json:"remove_role_image,omitempty"`
	// Multipart only: the "role_image" file part, uploaded to S3 by the
	// service; the resulting key is what lands on ma_roles.role_image_key.
	RoleImageFile        io.Reader `json:"-"`
	RoleImageFilename    string    `json:"-"`
	RoleImageContentType string    `json:"-"`
}

type UpdateRoleRes struct {
	Role *RoleResponse `json:"role"`
}

type DeleteRoleReq struct {
	RoleID int64 `json:"role_id"`
}

type DeleteRoleRes struct{}

type GetRoleReq struct {
	RoleID int64 `json:"role_id"`
}

type GetRoleRes struct {
	Role *RoleResponse `json:"role"`
}

type ListRolesReq struct {
	Search     *string `json:"search,omitempty"`
	RoleStatus *string `json:"role_status,omitempty"`
	RoleIDs    []int64 `json:"role_ids,omitempty"`
	Page       int64   `json:"page"`
	Size       int64   `json:"size"`
}

type ListRolesRes struct {
	Roles      []*RoleResponse        `json:"roles"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func DomainToResponse(r *domain.Role) *RoleResponse {
	if r == nil {
		return nil
	}
	return &RoleResponse{
		RoleID:       r.RoleId(),
		RoleCode:     r.RoleCode(),
		RoleName:     r.RoleName(),
		Description:  r.Description(),
		RoleImageKey: r.RoleImageKey(),
		Note:         r.Note(),
		RoleStatus:   r.RoleStatus(),
		CreateDt:     r.CreateDt().String(),
		ModifyDt:     r.ModifyDt().String(),
	}
}

func DomainListToResponse(roles []*domain.Role) []*RoleResponse {
	result := make([]*RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = DomainToResponse(r)
	}
	return result
}
