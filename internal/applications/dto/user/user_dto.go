package dto

import (
	"time"

	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserResponse struct {
	ID        int64     `json:"id"`
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
	Name     string     `json:"last_name"`
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
	UID        string        `json:"uid"`
	FirstName  *string       `json:"first_name,omitempty"`
	MiddleName *string       `json:"middle_name,omitempty"`
	LastName   *string       `json:"last_name,omitempty"`
	Phone      *string       `json:"phone,omitempty"`
	Email      *string       `json:"email,omitempty"`
	Role       *enum.ERole   `json:"role,omitempty"`
	Status     *enum.EStatus `json:"status,omitempty"`
}

type UpdateUserResponse struct {
	User *UserResponse `json:"result"`
}

type DeleteUserRequest struct {
	UserID string `json:"uid"`
}
