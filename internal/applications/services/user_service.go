package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	helper "math-ai.com/math-ai/internal/applications/helpers/user_helper"
	"math-ai.com/math-ai/internal/applications/utils"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	domain "math-ai.com/math-ai/internal/core/domain/profile"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

type UserService struct {
	validator       validators.IUserValidator
	repo            diRepo.IUserRepository
	profileRepo     diRepo.IProfileRepository
	roleRepo        diRepo.IRoleRepository
	responseBuilder *utils.ResponseBuilder
	fileManager     *utils.FileManager
	userCreator     *helper.UserCreator
	userUpdater     *helper.UserUpdater
	userDeleter     *helper.UserDeleter
}

func NewUserService(
	validator validators.IUserValidator,
	repo diRepo.IUserRepository,
	authRepo diRepo.IAuthRepository,
	profileRepo diRepo.IProfileRepository,
	roleRepo diRepo.IRoleRepository,
	userQuizPracticeRepo diRepo.IUserQuizPracticesRepository,
	storageService diSvc.IStorageService,
) diSvc.IUserService {
	responseBuilder := utils.NewResponseBuilder(storageService)
	fileManager := utils.NewFileManager(storageService)
	userCreator := helper.NewUserCreator(repo, authRepo, profileRepo)
	userUpdater := helper.NewUserUpdater(repo, profileRepo)
	userDeleter := helper.NewUserDeleter(repo, authRepo, profileRepo, userQuizPracticeRepo)

	return &UserService{
		validator:       validator,
		repo:            repo,
		profileRepo:     profileRepo,
		roleRepo:        roleRepo,
		responseBuilder: responseBuilder,
		fileManager:     fileManager,
		userCreator:     userCreator,
		userUpdater:     userUpdater,
		userDeleter:     userDeleter,
	}
}

func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUserRequest) (status.Code, []*dto.UserResponse, *pagination.Pagination, error) {
	params := diRepo.ListUsersParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	users, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.FAIL, nil, nil, err
	}

	if len(users) == 0 {
		return status.SUCCESS, []*dto.UserResponse{}, pagination, nil
	}

	// Build responses with presigned URLs using shared utility
	res := s.responseBuilder.BuildUserResponses(ctx, users)

	return status.SUCCESS, res, pagination, nil
}

func (s *UserService) GetUserByLoginName(ctx context.Context, loginName string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.GetUserByLoginName(ctx, loginName)
	if err != nil {
		return status.FAIL, nil, err
	}

	res := s.responseBuilder.BuildUserResponse(ctx, user)

	return status.SUCCESS, res, nil
}

func (s *UserService) GetUserByID(ctx context.Context, uid string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return status.FAIL, nil, err
	}
	if user == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res := s.responseBuilder.BuildUserResponse(ctx, user)

	return status.SUCCESS, res, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return status.FAIL, nil, err
	}
	if user == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res := s.responseBuilder.BuildUserResponse(ctx, user)

	return status.SUCCESS, res, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateUserRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Find role by code to set role ID
	role, err := s.roleRepo.FindByCode(ctx, string(enum.RoleUser))
	if err != nil || role == nil {
		return status.FAIL, nil, err
	}
	req.RoleID = role.ID()

	// Check for duplicate users using helper
	if statusCode, err := s.userCreator.CheckDuplicateUser(ctx, req.Email, req.Phone); err != nil {
		return statusCode, nil, err
	}

	// Handle avatar upload using shared file manager
	avatarKey, statusCode, err := s.fileManager.UploadFile(ctx, req.AvatarFile, req.AvatarFilename, req.AvatarContentType, "user")
	if err != nil {
		return statusCode, nil, err
	}

	// Build user domain
	createUserDomain := dto.BuildUserDomainForCreate(req)

	// Set avatar URL if uploaded
	if avatarKey != nil {
		createUserDomain.SetAvatarKey(avatarKey)
	}

	// Create user with transaction using helper
	err = s.userCreator.CreateWithTransaction(ctx, createUserDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch created user
	user, err := s.repo.FindByID(ctx, createUserDomain.ID())
	if err != nil {
		return status.FAIL, nil, err
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildUserResponse(ctx, user)

	return status.SUCCESS, res, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (status.Code, *dto.UserResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateUserRequest(req); err != nil {
		return statusCode, nil, err
	}

	println("grade_id", req.GradeID)

	// Get existing user
	existingUser, err := s.repo.FindByID(ctx, req.UID)
	if err != nil {
		return status.FAIL, nil, err
	}
	if existingUser == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	// Handle avatar updates using shared file manager
	newAvatarKey, statusCode, err := s.fileManager.UpdateFile(
		ctx,
		existingUser.AvatarKey(),
		req.AvatarFile,
		req.AvatarFilename,
		req.AvatarContentType,
		req.DeleteAvatar,
		"user",
	)
	if err != nil {
		return statusCode, nil, err
	}

	// Build user domain
	updateUserDomain := dto.BuildUserDomainForUpdate(req)

	// Set avatar URL if uploaded
	if newAvatarKey != nil {
		updateUserDomain.SetAvatarKey(newAvatarKey)
	}

	// Build profile domain if profile update is included
	var profileDomain *domain.Profile
	if req.GradeID != nil {
		profileDomain = dto.BuildProfileDomainForUpdate(&dto.UpdateProfileRequest{
			UID:        req.UID,
			GradeID:    req.GradeID,
			SemesterID: req.SemesterID,
		})
	}

	// Update user with transaction using helper
	err = s.userUpdater.UpdateWithTransaction(ctx, updateUserDomain, profileDomain)
	if err != nil {
		return status.FAIL, nil, err
	}

	// Fetch updated user
	user, findErr := s.repo.FindByID(ctx, req.UID)
	if findErr != nil {
		return status.FAIL, nil, findErr
	}

	// Build response using shared utility
	res := s.responseBuilder.BuildUserResponse(ctx, user)

	return status.SUCCESS, res, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (status.Code, error) {
	if statusCode, err := s.validator.ValidateDeleteUserRequest(req); err != nil {
		return statusCode, err
	}

	// Delete user and related data using helper
	err := s.userDeleter.DeleteWithTransaction(ctx, req.UID)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}

func (s *UserService) ForceDeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (status.Code, error) {
	if statusCode, err := s.validator.ValidateDeleteUserRequest(req); err != nil {
		return statusCode, err
	}

	// Get user to retrieve avatar for cleanup
	user, err := s.repo.FindByID(ctx, req.UID)
	if err != nil {
		return status.FAIL, err
	}

	// Force delete user and related data using helper
	err = s.userDeleter.ForceDeleteWithTransaction(ctx, req.UID)
	if err != nil {
		return status.FAIL, err
	}

	// Delete avatar from storage if exists
	if user != nil && user.AvatarKey() != nil {
		s.fileManager.DeleteFile(ctx, user.AvatarKey())
	}

	return status.SUCCESS, nil
}
