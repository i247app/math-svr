package user

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/pagination"
)

type UserResponse struct {
	ID       int64     `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Email    *string   `json:"email,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	CreateDt string    `json:"create_dt"`
	ModifyDt string    `json:"modify_dt"`
}

type CreateUserReq struct {
	Phone string  `json:"phone"`
	Email *string `json:"email,omitempty"`
}

type CreateUserRes struct {
	User *UserResponse `json:"user"`
}

type UpdateUserReq struct {
	ID     int64     `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Email  *string   `json:"email,omitempty"`
	Phone  *string   `json:"phone,omitempty"`
}

type UpdateUserRes struct {
	User *UserResponse `json:"user"`
}

type GetUserByUserIdReq struct {
	UserID uuid.UUID `json:"user_id"`
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
	UserID uuid.UUID `json:"user_id"`
}

type DeleteUserRes struct {
}

func DomainToResponse(u *user.User) *UserResponse {
	if u == nil {
		return nil
	}

	return &UserResponse{
		ID:       u.Id(),
		UserID:   u.UserId(),
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
