package user

import (
	"io"

	profileDto "math-ai.com/math-ai/internal/application/dto/profile"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/pagination"
	"math-ai.com/math-ai/internal/shared/utils"
)

type UserResponse struct {
	ID       int64   `json:"id"`
	UserID   string  `json:"user_id"`
	Email    *string `json:"email,omitempty"`
	Phone    string  `json:"phone,omitempty"`
	CreateDt string  `json:"create_dt"`
	ModifyDt string  `json:"modify_dt"`
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
	User    *UserResponse               `json:"user"`
	Profile *profileDto.ProfileResponse `json:"profile"`
}

type UpdateUserReq struct {
	ID     int64   `json:"id"`
	UserID string  `json:"user_id"`
	Email  *string `json:"email,omitempty"`
	Phone  *string `json:"phone,omitempty"`
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

func DomainToResponse(u *user.User) *UserResponse {
	if u == nil {
		return nil
	}

	var userId string
	if !utils.IsEmptyUUID(u.UserId()) {
		id := u.UserId().String()
		userId = id
	}

	return &UserResponse{
		ID:       u.Id(),
		UserID:   userId,
		Email:    u.Email(),
		Phone:    u.Phone(),
		CreateDt: u.CreateDt().String(),
		ModifyDt: u.ModifyDt().String(),
	}
}

func DomainListToResponse(users []*user.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, u := range users {
		result[i] = DomainToResponse(u)
	}
	return result
}
