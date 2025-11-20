package dto

import "math-ai.com/math-ai/internal/core/domain/user"

func BuildUserDomainFromCreateDTO(dto *CreateUserRequest) *user.User {
	userDomain := user.NewUserDomain()
	userDomain.SetEmail(dto.Email)
	userDomain.SetName(dto.Name)
	userDomain.SetPhone(dto.Phone)
	userDomain.SetPassword(dto.Password)

	return userDomain
}

func BuildUserDomainFromUpdateDTO(dto *UpdateUserRequest) *user.User {
	userDomain := user.NewUserDomain()
	userDomain.SetEmail(*dto.Email)
	userDomain.SetName(*dto.Name)
	userDomain.SetPhone(*dto.Phone)

	return userDomain
}

func UserResponseFromDomain(u *user.User) UserResponse {
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

func UserResponseListFromDomain(users []*user.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = UserResponseFromDomain(u)
	}
	return responses
}
