package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	"math-ai.com/math-ai/internal/core/di/repositories"
	di "math-ai.com/math-ai/internal/core/di/services"
	domain "math-ai.com/math-ai/internal/core/domain/user"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/utils/pagination"
)

const (
	AvatarPresignedURLExpiration = 24 * time.Hour // 24 hours
)

type UserService struct {
	validator      validators.IUserValidator
	repo           repositories.IUserRepository
	loginRepo      repositories.ILoginRepository
	profileRepo    repositories.IProfileRepository
	storageService di.IStorageService
}

func NewUserService(
	validator validators.IUserValidator,
	repo repositories.IUserRepository,
	loginRepo repositories.ILoginRepository,
	profileRepo repositories.IProfileRepository,
	storageService di.IStorageService,
) di.IUserService {
	return &UserService{
		validator:      validator,
		repo:           repo,
		loginRepo:      loginRepo,
		profileRepo:    profileRepo,
		storageService: storageService,
	}
}

func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUserRequest) (status.Code, []*dto.UserResponse, *pagination.Pagination, error) {
	params := repositories.ListUsersParams{
		Search:    req.Search,
		Page:      req.Page,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
		TakeAll:   req.TakeAll,
	}

	users, pagination, err := s.repo.List(ctx, params)
	if err != nil {
		return status.INTERNAL, nil, nil, err
	}

	if len(users) == 0 {
		return status.SUCCESS, []*dto.UserResponse{}, pagination, nil
	}

	// Build responses with presigned URLs
	res, err := s.buildUserResponsesWithPresignedURLs(ctx, users)
	if err != nil {
		return status.INTERNAL, nil, nil, err
	}

	return status.SUCCESS, res, pagination, nil
}

func (s *UserService) GetUserByLoginName(ctx context.Context, loginName string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.GetUserByLoginName(ctx, loginName)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res, err := s.buildUserResponseWithPresignedURL(ctx, user)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	return status.SUCCESS, res, nil
}

func (s *UserService) GetUserByID(ctx context.Context, uid string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if user == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res, err := s.buildUserResponseWithPresignedURL(ctx, user)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	return status.SUCCESS, res, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (status.Code, *dto.UserResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if user == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	res, err := s.buildUserResponseWithPresignedURL(ctx, user)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	return status.SUCCESS, res, nil
}

// deleteOldAvatar removes old avatar from S3 if it exists
func (s *UserService) deleteOldAvatar(ctx context.Context, avatarKey *string) {
	if avatarKey == nil || *avatarKey == "" {
		return
	}

	deleteReq := &dto.DeleteFileRequest{
		Key: *avatarKey,
	}

	_, err := s.storageService.HandleDelete(ctx, deleteReq)
	if err != nil {
		//logger.Warnf("Failed to delete old avatar (%s): %v", *avatarKey, err)
		// Don't return error - old avatar cleanup is not critical
	}
}

// buildUserResponseWithPresignedURL creates a UserResponse with presigned avatar URL
func (s *UserService) buildUserResponseWithPresignedURL(ctx context.Context, user *domain.User) (*dto.UserResponse, error) {
	res := dto.UserResponseFromDomain(user)

	// Generate presigned URL for avatar if exists
	if user.AvatarURL() != nil && *user.AvatarURL() != "" {
		_, presignedURL, err := s.storageService.CreatePresignedUrl(ctx, &dto.CreatePresignedUrlRequest{
			Key:        *user.AvatarURL(),
			Expiration: AvatarPresignedURLExpiration,
		})
		if err != nil {
			//logger.Warnf("Failed to generate presigned URL for avatar (%s): %v", *user.AvatarURL(), err)
			// Don't fail the request if presigned URL generation fails
			// User data is still valid, just without the presigned URL
		} else {
			res.AvatarPreviewURL = &presignedURL
		}
	}

	return &res, nil
}

// buildUserResponsesWithPresignedURLs creates UserResponses with presigned avatar URLs
func (s *UserService) buildUserResponsesWithPresignedURLs(ctx context.Context, users []*domain.User) ([]*dto.UserResponse, error) {
	responses := make([]*dto.UserResponse, len(users))

	for i, user := range users {
		res, err := s.buildUserResponseWithPresignedURL(ctx, user)
		if err != nil {
			return nil, err
		}
		responses[i] = res
	}

	return responses, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (status.Code, *dto.UserResponse, error) {
	// vaidate request
	if statusCode, err := s.validator.ValidateCreateUserRequest(req); err != nil {
		return statusCode, nil, err
	}

	for _, aka := range []string{req.Email, req.Phone} {
		if aka == "" {
			continue // Skip empty aliases
		}
		existingUser, err := s.repo.GetUserByLoginName(ctx, aka)
		if err != nil {
			return status.INTERNAL, nil, err
		}

		if existingUser != nil {
			if existingUser.Email() == req.Email {
				return status.USER_EMAIL_ALREADY_EXISTS, nil, err_svc.ErrEmailAlreadyExists
			} else if existingUser.Phone() == req.Phone {
				return status.USER_PHONE_ALREADY_EXISTS, nil, err_svc.ErrPhoneAlreadyExists
			}
		}
	}

	// Handle avatar upload before creating user
	var avatarKey *string
	if req.AvatarFile != nil {
		statusCode, res, err := s.storageService.HandleUpload(ctx, &dto.UploadFileRequest{
			File:        req.AvatarFile,
			Filename:    req.AvatarFilename,
			ContentType: req.AvatarContentType,
			Folder:      "user",
		})
		if err != nil {
			return statusCode, nil, fmt.Errorf("failed to upload avatar: %w", err)
		}
		avatarKey = &res.Key
	}

	createUserDomain := dto.BuildUserDomainForCreate(req)

	// Set avatar URL if uploaded
	if avatarKey != nil {
		createUserDomain.SetAvatarURL(avatarKey)
	}

	handler := func(tx *sql.Tx) error {
		// Create the user
		_, err := s.repo.Create(ctx, tx, createUserDomain)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Store aliases
		for _, aka := range []string{createUserDomain.Email(), createUserDomain.Phone()} {
			if aka == "" {
				continue // Skip empty aliases
			}

			createAliasDomain := dto.BuildAliasDomain(createUserDomain.ID(), aka)
			if err := s.repo.StoreUserAlias(ctx, tx, createAliasDomain); err != nil {
				return fmt.Errorf("failed to store user alias in transaction: %v", err)
			}
		}

		// Store login
		createLoginDomain := dto.BuildLoginDomain(createUserDomain.ID(), createUserDomain.Password())
		if err := s.loginRepo.StoreLogin(ctx, tx, createLoginDomain); err != nil {
			return fmt.Errorf("failed to store user login in transaction: %v", err)
		}

		return nil
	}

	// Store login
	err := s.repo.DoTransaction(ctx, handler)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	user, err := s.repo.FindByID(ctx, createUserDomain.ID())
	if err != nil {
		return status.INTERNAL, nil, err
	}

	res, err := s.buildUserResponseWithPresignedURL(ctx, user)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	return status.SUCCESS, res, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (status.Code, *dto.UserResponse, error) {
	// vaidate request
	if statusCode, err := s.validator.ValidateUpdateUserRequest(req); err != nil {
		return statusCode, nil, err
	}

	// Get existing user to check for old avatar
	existingUser, err := s.repo.FindByID(ctx, req.UID)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if existingUser == nil {
		return status.USER_NOT_FOUND, nil, err_svc.ErrUserNotFound
	}

	// Handle avatar updates
	var newAvatarKey *string
	if req.AvatarFile != nil {
		// Upload new avatar
		statusCode, res, err := s.storageService.HandleUpload(ctx, &dto.UploadFileRequest{
			File:        req.AvatarFile,
			Filename:    req.AvatarFilename,
			ContentType: req.AvatarContentType,
			Folder:      "user",
		})
		if err != nil {
			return statusCode, nil, fmt.Errorf("failed to upload avatar: %w", err)
		}
		newAvatarKey = &res.Key

		// Delete old avatar if exists
		if existingUser.AvatarURL() != nil {
			s.deleteOldAvatar(ctx, existingUser.AvatarURL())
		}
	} else if req.DeleteAvatar {
		// Delete avatar if requested
		if existingUser.AvatarURL() != nil {
			s.deleteOldAvatar(ctx, existingUser.AvatarURL())
		}
		emptyString := ""
		newAvatarKey = &emptyString
	}

	handler := func(tx *sql.Tx) error {
		userDomain := dto.BuildUserDomainForUpdate(req)

		// Set avatar URL if changed
		if newAvatarKey != nil {
			userDomain.SetAvatarURL(newAvatarKey)
		}
		_, updateErr := s.repo.Update(ctx, userDomain)
		if updateErr != nil {
			return updateErr
		}

		if req.Level != nil || req.Grade != nil {
			profileDomain := dto.BuildProfileDomainForUpdate(&dto.UpdateProfileRequest{
				UID:   userDomain.ID(),
				Grade: req.Grade,
				Level: req.Level,
			})
			_, profileErr := s.profileRepo.Update(ctx, profileDomain)
			if profileErr != nil {
				return profileErr
			}
		}

		return nil
	}

	txErr := s.repo.DoTransaction(ctx, handler)
	if txErr != nil {
		return status.INTERNAL, nil, txErr
	}

	user, findErr := s.repo.FindByID(ctx, req.UID)
	if findErr != nil {
		return status.INTERNAL, nil, findErr
	}

	res, err := s.buildUserResponseWithPresignedURL(ctx, user)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	return status.SUCCESS, res, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (status.Code, error) {
	if statusCode, err := s.validator.ValidateDeleteUserRequest(req); err != nil {
		return statusCode, err
	}

	handler := func(tx *sql.Tx) error {
		// Delete users
		err := s.repo.Delete(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Delete user aliases
		err = s.repo.DeleteUserAlias(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user aliases in transaction: %v", err)
		}

		// Delete user logins
		err = s.loginRepo.DeleteLogin(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user logins in transaction: %v", err)
		}

		// Delete user profile
		err = s.profileRepo.DeleteByUID(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user profile in transaction: %v", err)
		}

		return nil
	}

	err := s.repo.DoTransaction(ctx, handler)
	if err != nil {
		return status.INTERNAL, err
	}
	return status.SUCCESS, nil
}

func (s *UserService) ForceDeleteUser(ctx context.Context, req *dto.DeleteUserRequest) (status.Code, error) {
	if statusCode, err := s.validator.ValidateDeleteUserRequest(req); err != nil {
		return statusCode, err
	}

	user, err := s.repo.FindByID(ctx, req.UID)
	if err != nil {
		return status.INTERNAL, err
	}

	handler := func(tx *sql.Tx) error {
		// Delete users
		err := s.repo.ForceDelete(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to create user in transaction: %v", err)
		}

		// Delete user aliases
		err = s.repo.ForceDeleteUserAlias(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user aliases in transaction: %v", err)
		}

		// Delete user logins
		err = s.loginRepo.ForceDeleteLogin(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user logins in transaction: %v", err)
		}

		// Delete user profile
		err = s.profileRepo.ForceDeleteByUID(ctx, tx, req.UID)
		if err != nil {
			return fmt.Errorf("failed to delete user profile in transaction: %v", err)
		}

		return nil
	}

	err = s.repo.DoTransaction(ctx, handler)
	if err != nil {
		return status.INTERNAL, err
	}

	// Delete avatar from storage if exists
	if user.AvatarURL() != nil {
		s.deleteOldAvatar(ctx, user.AvatarURL())
	}

	return status.SUCCESS, nil
}
