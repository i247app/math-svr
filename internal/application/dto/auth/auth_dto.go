package auth

import "math-ai.com/math-ai/internal/application/dto/user"

type LoginReq struct {
	Phone string `json:"phone"`
}

type LoginRes struct {
	User *user.UserResponse `json:"user"`
}
