package resolvers

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type UserResolver struct {
	userService di.IUserService
}

func NewUserResolver(userService di.IUserService) *UserResolver {
	return &UserResolver{
		userService: userService,
	}
}

// GetUser resolves a single user by ID
func (r *UserResolver) GetUser(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return nil, nil
	}

	statusCode, user, err := r.userService.GetUserByID(params.Context, id)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return user, nil
}

// ListUsers resolves a list of users with pagination
func (r *UserResolver) ListUsers(params graphql.ResolveParams) (interface{}, error) {
	req := &dto.ListUserRequest{
		Page:    1,
		Limit:   10,
		TakeAll: false,
	}

	// Extract pagination parameters
	if page, ok := params.Args["page"].(int); ok {
		req.Page = int64(page)
	}
	if limit, ok := params.Args["limit"].(int); ok {
		req.Limit = int64(limit)
	}
	if search, ok := params.Args["search"].(string); ok {
		req.Search = search
	}
	if orderBy, ok := params.Args["order_by"].(string); ok {
		req.OrderBy = orderBy
	}
	if orderDesc, ok := params.Args["order_desc"].(bool); ok {
		req.OrderDesc = orderDesc
	}
	if takeAll, ok := params.Args["take_all"].(bool); ok {
		req.TakeAll = takeAll
	}

	statusCode, users, pagination, err := r.userService.ListUsers(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return map[string]interface{}{
		"items":      users,
		"pagination": pagination,
	}, nil
}

// CreateUser resolves user creation
func (r *UserResolver) CreateUser(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.CreateUserRequest{
		Name:     input["name"].(string),
		Email:    input["email"].(string),
		Phone:    input["phone"].(string),
		Password: input["password"].(string),
	}

	if role, ok := input["role"].(string); ok {
		req.Role = enum.ERole(role)
	}
	if deviceUUID, ok := input["device_uuid"].(string); ok {
		req.DeviceUUID = deviceUUID
	}
	if deviceName, ok := input["device_name"].(string); ok {
		req.DeviceName = deviceName
	}

	statusCode, user, err := r.userService.CreateUser(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return user, nil
}

// UpdateUser resolves user updates
func (r *UserResolver) UpdateUser(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	uid, ok := input["uid"].(string)
	if !ok {
		return nil, nil
	}

	req := &dto.UpdateUserRequest{
		UID: uid,
	}

	if name, ok := input["name"].(string); ok {
		req.Name = &name
	}
	if email, ok := input["email"].(string); ok {
		req.Email = &email
	}
	if phone, ok := input["phone"].(string); ok {
		req.Phone = &phone
	}
	if role, ok := input["role"].(string); ok {
		r := enum.ERole(role)
		req.Role = &r
	}
	if statusStr, ok := input["status"].(string); ok {
		s := enum.EStatus(statusStr)
		req.Status = &s
	}

	statusCode, user, err := r.userService.UpdateUser(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return user, nil
}

// DeleteUser resolves user soft deletion
func (r *UserResolver) DeleteUser(params graphql.ResolveParams) (interface{}, error) {
	uid, ok := params.Args["uid"].(string)
	if !ok {
		return false, nil
	}

	req := &dto.DeleteUserRequest{
		UID: uid,
	}

	statusCode, err := r.userService.DeleteUser(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return false, err
	}

	return true, nil
}

// ForceDeleteUser resolves user hard deletion
func (r *UserResolver) ForceDeleteUser(params graphql.ResolveParams) (interface{}, error) {
	uid, ok := params.Args["uid"].(string)
	if !ok {
		return false, nil
	}

	req := &dto.DeleteUserRequest{
		UID: uid,
	}

	statusCode, err := r.userService.ForceDeleteUser(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return false, err
	}

	return true, nil
}
