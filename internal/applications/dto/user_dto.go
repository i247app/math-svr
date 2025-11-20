package dto

import (
	"time"

	domain "math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	AvatarURL *string   `json:"avatar_url"`
	Role      string    `json:"role"`
	CreateAt  time.Time `json:"created_at"`
	ModifyAt  time.Time `json:"modified_at"`
}

type GetUserResponse struct {
	User *UserResponse `json:"result"`
}

type ListUserRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	Page      int64  `json:"page" form:"page"`
	Limit     int64  `json:"size" form:"size"`
	OrderBy   string `json:"order_by" form:"order_by"`
	OrderDesc bool   `json:"order_desc" form:"order_desc"`
	TakeAll   bool   `json:"take_all" form:"take_all"`
}

type ListUserResponse struct {
	Items      []*UserResponse        `json:"result"`
	Pagination *pagination.Pagination `json:"metadata"`
}

type CreateUserRequest struct {
	Name     string     `json:"name"`
	Phone    string     `json:"phone"`
	Email    string     `json:"email"`
	Role     enum.ERole `json:"role,omitempty"`
	Password string     `json:"password"`

	DeviceUUID string `json:"device_uuid,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

type CreateUserResponse struct {
	User *UserResponse `json:"result"`
}

type UpdateUserRequest struct {
	UID    string        `json:"uid"`
	Name   *string       `json:"name,omitempty"`
	Phone  *string       `json:"phone,omitempty"`
	Email  *string       `json:"email,omitempty"`
	Role   *enum.ERole   `json:"role,omitempty"`
	Status *enum.EStatus `json:"status,omitempty"`
}

type UpdateUserResponse struct {
	User *UserResponse `json:"result"`
}

type DeleteUserRequest struct {
	UserID string `json:"uid"`
}

func BuildUserDomainForCreate(dto *CreateUserRequest) *domain.User {
	userDomain := domain.NewUserDomain()
	userDomain.GenerateID()
	userDomain.SetEmail(dto.Email)
	userDomain.SetName(dto.Name)
	userDomain.SetPhone(dto.Phone)
	userDomain.SetPassword(dto.Password)
	userDomain.SetRole(string(dto.Role))

	return userDomain
}

func BuildUserDomainForUpdate(dto *UpdateUserRequest) *domain.User {
	userDomain := domain.NewUserDomain()
	userDomain.SetID(dto.UID)

	if dto.Email != nil {
		userDomain.SetEmail(*dto.Email)
	}

	if dto.Name != nil {
		userDomain.SetName(*dto.Name)
	}

	if dto.Phone != nil {
		userDomain.SetPhone(*dto.Phone)
	}

	if dto.Role != nil {
		userDomain.SetRole(string(*dto.Role))
	}

	return userDomain
}

func BuildAliasDomain(uid string, aka string) *domain.Alias {
	aliasDomain := domain.NewAliasDomain()
	aliasDomain.GenerateID()
	aliasDomain.SetUID(uid)
	aliasDomain.SetAka(aka)

	return aliasDomain
}

func UserResponseFromDomain(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID(),
		Email:     u.Email(),
		Name:      u.Name(),
		Phone:     u.Phone(),
		AvatarURL: u.AvatarURL(),
		Role:      u.Role(),
		CreateAt:  u.CreatedAt(),
		ModifyAt:  u.ModifyAt(),
	}
}

func UserResponseListFromDomain(users []*domain.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = UserResponseFromDomain(u)
	}
	return responses
}
