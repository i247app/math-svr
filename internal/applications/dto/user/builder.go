package dto

import "math-ai.com/math-ai/internal/core/domain/user"

func UserResponseFromDomain(u *user.User) UserResponse {
	return UserResponse{
		ID:        u.ID(),
		Email:     u.Email(),
		Name:      u.Name(),
		Phone:     u.Phone(),
		AvatarURL: u.AvatarURL(),
		CreateAt:  u.CreatedAt(),
		ModifyAt:  u.ModifyAt(),
	}
}

func UserResponseListFromDomain(users []*user.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = UserResponseFromDomain(u)
	}
	return responses
}
