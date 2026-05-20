package user

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/user"
	profileDto "math-ai.com/math-ai/internal/application/dto/profile"
	dto "math-ai.com/math-ai/internal/application/dto/user"
	query "math-ai.com/math-ai/internal/application/query/user"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	domain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

// avatarFolder is the S3 prefix new parent-signup avatars land under. It
// matches the prefix used by /profiles/upload-avatar so an operator looking
// at the bucket sees a single namespace per concern.
const avatarFolder = "profile-avatars"

type Service struct {
	getUserByUserIdQuery *query.GetUserByUserIdQueryHandler
	getUserByPhoneQuery  *query.GetUserByPhoneQueryHandler
	getUserByEmailQuery  *query.GetUserByEmailQueryHandler
	listUsersQuery       *query.ListUsersQueryHandler
	createUserCmd        *command.CreateUserCommandHandler
	updateUserCmd        *command.UpdateUserCommandHandler
	softDeleteUserCmd    *command.SoftDeleteUserCommandHandler
	forceDeleteUserCmd   *command.ForceDeleteUserCommandHandler
	storageProvider      *storage.Adapter
}

func NewService(
	repo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getUserByUserIdQuery: query.NewGetUserByUserIdQueryHandler(repo),
		getUserByPhoneQuery:  query.NewGetUserByPhoneQueryHandler(repo),
		getUserByEmailQuery:  query.NewGetUserByEmailQueryHandler(repo),
		listUsersQuery:       query.NewListUsersQueryHandler(repo),
		createUserCmd:        command.NewCreateUserCommandHandler(uow),
		updateUserCmd:        command.NewUpdateUserCommandHandler(uow),
		softDeleteUserCmd:    command.NewSoftDeleteUserCommandHandler(uow),
		forceDeleteUserCmd:   command.NewForceDeleteUserCommandHandler(uow),
		storageProvider:      storageProvider,
	}
}

func (s *Service) GetUserById(ctx context.Context, req *dto.GetUserByUserIdReq) (*dto.GetUserByUserIdRes, error) {
	user, err := s.getUserByUserIdQuery.Handle(ctx, query.GetUserByUserIdQuery{UserId: req.UserID})
	if err != nil {
		return nil, err
	}

	res := &dto.GetUserByUserIdRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}

func (s *Service) ListUsers(ctx context.Context, req *dto.ListUsersReq) (*dto.ListUsersRes, error) {
	users, pg, err := s.listUsersQuery.Handle(ctx, &query.ListUsersQuery{
		Page:  int64(req.Page),
		Limit: int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListUsersRes{
		Users:      dto.DomainListToResponse(users),
		Pagination: pg,
	}, nil
}

func (s *Service) GetUserByPhone(ctx context.Context, req *dto.GetUserByPhoneReq) (*dto.GetUserByPhoneRes, error) {
	user, err := s.getUserByPhoneQuery.Handle(ctx, query.GetUserByPhoneQuery{Phone: req.Phone})
	if err != nil {
		return nil, err
	}

	res := &dto.GetUserByPhoneRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, req *dto.GetUserByEmailReq) (*dto.GetUserByEmailRes, error) {
	user, err := s.getUserByEmailQuery.Handle(ctx, query.GetUserByEmailQuery{Email: req.Email})
	if err != nil {
		return nil, err
	}

	res := &dto.GetUserByEmailRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}

func (s *Service) CreateUser(ctx context.Context, sess *session.AppSession, req *dto.CreateUserReq) (*dto.CreateUserRes, error) {
	log := logger.From(ctx)

	if err := ValidateCreateUser(ctx, req); err != nil {
		return nil, err
	}

	// Upload the avatar BEFORE opening the transaction. The DB write is
	// the cheap, fast step; the S3 round-trip is slow and would hold a tx
	// open if interleaved. If the UoW fails after upload, the orphan S3
	// object is best-effort deleted.
	avatarKey, err := s.uploadAvatarIfPresent(ctx, req)
	if err != nil {
		return nil, err
	}

	var email *string
	if strings.TrimSpace(req.Email) != "" {
		e := req.Email
		email = &e
	}

	created, err := s.createUserCmd.Handle(ctx, command.CreateUserCommand{
		Phone:     req.Phone,
		Email:     email,
		Name:      req.Name,
		AvatarKey: avatarKey,
	})
	if err != nil {
		if avatarKey != nil {
			if delErr := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: *avatarKey}); delErr != nil {
				log.Warnf("user.create avatar orphan cleanup failed key=%s err=%v", *avatarKey, delErr)
			}
		}
		return nil, err
	}

	log.Info("user.created",
		"user_id", created.User.UserId(),
		"profile_id", created.Profile.ProfileId(),
	)

	userRes := dto.DomainToResponse(created.User)
	profileRes := profileDto.DomainToResponse(created.Profile)
	s.populateImageUrl(ctx, profileRes)

	log.Info("Login successful, updating session data...")
	sessionData := session.InitData{
		Source:    "login",
		IsSecure:  true,
		UID:       userRes.UserID.String(),
		LoginName: userRes.Phone,
	}

	if userRes.Email != nil {
		sessionData.Email = *userRes.Email
	}

	sess.Init(sessionData)

	return &dto.CreateUserRes{
		User:    userRes,
		Profile: profileRes,
	}, nil
}

// uploadAvatarIfPresent ships the multipart avatar (if any) to S3 and returns
// the resulting key. Returns (nil, nil) when no avatar was submitted. Errors
// out with PROFILE_AVATAR_* codes since the eventual destination is a
// ma_profiles row.
func (s *Service) uploadAvatarIfPresent(ctx context.Context, req *dto.CreateUserReq) (*string, error) {
	if req.AvatarFile == nil || req.AvatarFilename == "" {
		return nil, nil
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			errors.New("storage adapter is not configured"))
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    req.AvatarFilename,
		ContentType: req.AvatarContentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        req.AvatarFile,
		Filename:    req.AvatarFilename,
		ContentType: req.AvatarContentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil,
			errors.New("upload returned an empty key"))
	}
	return &uploaded.Key, nil
}

func (s *Service) SoftDeleteUser(ctx context.Context, req *dto.DeleteUserReq) (*dto.DeleteUserRes, error) {
	if err := ValidateDeleteUser(ctx, req); err != nil {
		return nil, err
	}

	user, err := s.getUserByUserIdQuery.Handle(ctx, query.GetUserByUserIdQuery{UserId: req.UserID})
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errs.NewError(ctx, status.NOT_FOUND, nil, nil)
	}

	if err := s.softDeleteUserCmd.Handle(ctx, command.SoftDeleteUserCommand{
		UserID: req.UserID,
	}); err != nil {
		return nil, err
	}

	return &dto.DeleteUserRes{}, nil
}

func (s *Service) ForceDeleteUser(ctx context.Context, req *dto.DeleteUserReq) (*dto.DeleteUserRes, error) {
	if err := ValidateDeleteUser(ctx, req); err != nil {
		return nil, err
	}

	if err := s.forceDeleteUserCmd.Handle(ctx, command.ForceDeleteUserCommand{
		UserID: req.UserID,
	}); err != nil {
		return nil, err
	}

	return &dto.DeleteUserRes{}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *dto.UpdateUserReq) (*dto.UpdateUserRes, error) {
	if err := ValidateUpdateUser(ctx, req); err != nil {
		return nil, err
	}

	user, err := s.updateUserCmd.Handle(ctx, command.UpdateUserCommand{
		ID:     req.ID,
		UserID: req.UserID,
		Email:  req.Email,
		Phone:  req.Phone,
	})
	if err != nil {
		return nil, err
	}

	res := &dto.UpdateUserRes{
		User: dto.DomainToResponse(user),
	}

	return res, nil
}
