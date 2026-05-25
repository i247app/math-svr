package user

import (
	"io"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type UserResponse struct {
	ID     int64   `json:"id"`
	UserID string  `json:"user_id"`
	Name   string  `json:"name"`
	Email  *string `json:"email,omitempty"`
	Phone  string  `json:"phone,omitempty"`
	// AvatarKey is the raw S3 object key persisted on the user row.
	// AvatarUrl is a short-lived presigned URL the module layer fills
	// in on the way out (see populateImageUrl in module/user). Clients
	// should display AvatarUrl and ignore AvatarKey.
	AvatarKey *string `json:"avatar_key,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
	CreateDt  string  `json:"create_dt"`
	ModifyDt  string  `json:"modify_dt"`
}

type GetUserByPhoneReq struct {
	Phone string `json:"phone"`
}

type GetUserByPhoneRes struct {
	User *UserResponse `json:"user"`
}

type GetUserByEmailReq struct {
	Email string `json:"email"`
}

type GetUserByEmailRes struct {
	User *UserResponse `json:"user"`
}

type CreateUserReq struct {
	// UserName is the parent's display name — persisted to
	// ma_users.user_name (NOT NULL). Distinct from Name, which is the
	// child's name and lands in ma_profiles.name.
	// UserName string `json:"user_name"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email,omitempty"`

	AvatarFile        io.Reader `json:"avatar_file"`         // File reader
	AvatarFilename    string    `json:"avatar_file_name"`    // Original filename
	AvatarContentType string    `json:"avatar_content_type"` // MIME type
}

// CreateUserRes carries both the freshly-created parent and their initial
// child profile so the client doesn't need a follow-up /profiles/list round
// trip during onboarding.
type CreateUserRes struct {
	User *UserResponse `json:"user"`
}

type UpdateUserReq struct {
	ID     int64   `json:"id"`
	UserID string  `json:"user_id"`
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Phone  *string `json:"phone,omitempty"`

	AvatarFile        io.Reader `json:"avatar_file"`         // File reader
	AvatarFilename    string    `json:"avatar_file_name"`    // Original filename
	AvatarContentType string    `json:"avatar_content_type"` // MIME type
}

type UpdateUserRes struct {
	User *UserResponse `json:"user"`
}

type GetUserByUserIdReq struct {
	UserID string `json:"user_id"`
}

type GetUserByUserIdRes struct {
	User *UserResponse `json:"user"`
}

type ListUsersReq struct {
	Page int64 `json:"page"`
	Size int64 `json:"size"`
}

type ListUsersRes struct {
	Users      []*UserResponse        `json:"users"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type DeleteUserReq struct {
	UserID string `json:"user_id"`
}

type DeleteUserRes struct {
}

// UploadAvatarRes mirrors profile.UploadAvatarRes — same shape so the
// mobile client can use one renderer for both. AvatarUrl is the
// presigned URL; AvatarKey is exposed so the caller can persist it for
// long-lived references (e.g. re-presigning later via /users/me).
type UploadAvatarRes struct {
	UserID    string `json:"user_id"`
	AvatarKey string `json:"avatar_key"`
	AvatarUrl string `json:"avatar_url"`
}

func DomainToResponse(u *user.User) *UserResponse {
	if u == nil {
		return nil
	}

	return &UserResponse{
		ID:        u.Id(),
		UserID:    u.UserId(),
		Name:      u.UserName(),
		Email:     u.Email(),
		Phone:     u.Phone(),
		AvatarKey: u.AvatarKey(),
		CreateDt:  u.CreateDt().String(),
		ModifyDt:  u.ModifyDt().String(),
	}
}

func DomainListToResponse(users []*user.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, u := range users {
		result[i] = DomainToResponse(u)
	}
	return result
}
