package dto

import (
	"time"

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
