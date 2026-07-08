package user

import (
	"context"
	"fmt"
	"io"
	"strings"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/user"
	deviceDTO "math-ai.com/math-ai/internal/application/dto/device"
	dto "math-ai.com/math-ai/internal/application/dto/user"
	query "math-ai.com/math-ai/internal/application/query/user"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	domain "math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// avatarFolder is the S3 prefix user (parent) avatars land under.
// Separate from the "profile-avatars" prefix the profile module uses
// for child-profile avatars so the two concerns are bucket-separable.
//
// The presigned-URL lifetime is the imageUrlTTL constant in
// service_helper.go (same package).
const avatarFolder = "user-avatars"

type Service struct {
	deviceSvc            *device.Service
	getUserByUserIdQuery *query.GetUserByUserIdQueryHandler
	getUserByPhoneQuery  *query.GetUserByPhoneQueryHandler
	getUserByEmailQuery  *query.GetUserByEmailQueryHandler
	listUsersQuery       *query.ListUsersQueryHandler
	createUserCmd        *command.CreateUserCommandHandler
	updateUserCmd        *command.UpdateUserCommandHandler
	setAvatarKeyCmd      *command.SetAvatarKeyCommandHandler
	softDeleteUserCmd    *command.SoftDeleteUserCommandHandler
	forceDeleteUserCmd   *command.ForceDeleteUserCommandHandler
	storageProvider      *storage.Adapter
}

func NewService(
	deviceSvc *device.Service,
	repo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		deviceSvc:            deviceSvc,
		getUserByUserIdQuery: query.NewGetUserByUserIdQueryHandler(repo),
		getUserByPhoneQuery:  query.NewGetUserByPhoneQueryHandler(repo),
		getUserByEmailQuery:  query.NewGetUserByEmailQueryHandler(repo),
		listUsersQuery:       query.NewListUsersQueryHandler(repo),
		createUserCmd:        command.NewCreateUserCommandHandler(uow),
		updateUserCmd:        command.NewUpdateUserCommandHandler(uow),
		setAvatarKeyCmd:      command.NewSetAvatarKeyCommandHandler(uow),
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

	userRes := dto.DomainToResponse(user)
	s.populateImageUrl(ctx, userRes)

	return &dto.GetUserByUserIdRes{User: userRes}, nil
}

func (s *Service) ListUsers(ctx context.Context, req *dto.ListUsersReq) (*dto.ListUsersRes, error) {
	users, pg, err := s.listUsersQuery.Handle(ctx, &query.ListUsersQuery{
		Page:  int64(req.Page),
		Limit: int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	responses := dto.DomainListToResponse(users)
	for _, r := range responses {
		s.populateImageUrl(ctx, r)
	}

	return &dto.ListUsersRes{
		Users:      responses,
		Pagination: pg,
	}, nil
}

func (s *Service) GetUserByPhone(ctx context.Context, req *dto.GetUserByPhoneReq) (*dto.GetUserByPhoneRes, error) {
	user, err := s.getUserByPhoneQuery.Handle(ctx, query.GetUserByPhoneQuery{Phone: req.Phone})
	if err != nil {
		return nil, err
	}

	userRes := dto.DomainToResponse(user)
	s.populateImageUrl(ctx, userRes)

	return &dto.GetUserByPhoneRes{User: userRes}, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, req *dto.GetUserByEmailReq) (*dto.GetUserByEmailRes, error) {
	user, err := s.getUserByEmailQuery.Handle(ctx, query.GetUserByEmailQuery{Email: req.Email})
	if err != nil {
		return nil, err
	}

	userRes := dto.DomainToResponse(user)
	s.populateImageUrl(ctx, userRes)

	return &dto.GetUserByEmailRes{User: userRes}, nil
}

func (s *Service) CreateUser(ctx context.Context, sess *session.AppSession, req *dto.CreateUserReq) (*dto.CreateUserRes, error) {
	log := logger.From(ctx)

	if err := ValidateCreateUser(ctx, req); err != nil {
		return nil, err
	}

	// Resolve the avatar reference (if any) BEFORE opening the
	// transaction. Two mutually-exclusive paths — validator already
	// enforced "at most one":
	//
	//   1. req.Avatar != "" — client supplied a URL or bare key for an
	//      object that already lives in our bucket. Just normalize and
	//      persist; nothing to upload.
	//
	//   2. req.AvatarFile != nil — multipart upload; ship to S3 and use
	//      the returned key.
	//
	// Branch one is the new path. The DB write is the cheap, fast step;
	// the S3 round-trip (when it happens) is slow and would hold a tx
	// open if interleaved. If the UoW fails after a fresh upload, the
	// orphan S3 object is best-effort deleted. An object referenced via
	// path 1 is NOT deleted on rollback — the client owns its lifecycle.
	var (
		avatarKey      *string
		uploadedOnThis bool
	)
	switch {
	case strings.TrimSpace(req.Avatar) != "":
		key, err := s.normalizeAvatarKey(ctx, req.Avatar, status.USER_AVATAR_INVALID_REFERENCE)
		if err != nil {
			return nil, err
		}
		avatarKey = &key
	case req.AvatarFile != nil:
		key, err := s.uploadAvatarIfPresent(ctx, req)
		if err != nil {
			return nil, err
		}
		avatarKey = key
		uploadedOnThis = key != nil
	}

	var email *string
	if strings.TrimSpace(req.Email) != "" {
		e := req.Email
		email = &e
	}

	phoneForString, err := utils.NormalizePhone(req.Phone)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("failed to normalize phone: %w", err))
	}
	log.Infof("Phone for string: %s", phoneForString)

	// phoneDecoded, err := url.QueryUnescape(req.Phone)
	// if err != nil {
	// 	return nil, errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("failed to decode phone: %w", err))
	// }

	created, err := s.createUserCmd.Handle(ctx, command.CreateUserCommand{
		Role:      enum.RoleType(req.Role),
		Phone:     phoneForString,
		Email:     email,
		UserName:  req.Name,
		AvatarKey: avatarKey,
	})
	if err != nil {
		// Only delete objects we just uploaded — a client-supplied
		// reference points to storage the client owns (or a prior
		// upload they're reusing); we must not garbage-collect it on
		// our rollback.
		if uploadedOnThis && avatarKey != nil {
			if delErr := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: *avatarKey}); delErr != nil {
				log.Warnf("user.create avatar orphan cleanup failed key=%s err=%v", *avatarKey, delErr)
			}
		}
		return nil, err
	}

	log.Info("Mark device as trusted")
	_, err = s.deviceSvc.VerifyDevice(ctx, &deviceDTO.VerifyDeviceReq{
		UserID:          created.User.UserId(),
		DeviceUUID:      metadata.GetDeviceID(ctx),
		DeviceName:      metadata.GetDeviceName(ctx),
		DevicePushToken: utils.ToStringPtr(metadata.GetDevicePushToken(ctx)),
	})
	if err != nil {
		return nil, err
	}

	userRes := dto.DomainToResponse(created.User)
	s.populateImageUrl(ctx, userRes)

	log.Info("Login successful, updating session data...")
	sessionData := session.InitData{
		Source:    "login",
		IsSecure:  true,
		UID:       userRes.UserID,
		LoginName: userRes.Phone,
	}

	if userRes.Email != nil {
		sessionData.Email = *userRes.Email
	}

	sess.Init(sessionData)

	return &dto.CreateUserRes{
		User: userRes,
	}, nil
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
		UserID: user.UserId(),
	}); err != nil {
		return nil, err
	}

	return &dto.DeleteUserRes{}, nil
}

func (s *Service) ForceDeleteUser(ctx context.Context, req *dto.DeleteUserReq) (*dto.DeleteUserRes, error) {
	log := logger.From(ctx)

	if err := ValidateDeleteUser(ctx, req); err != nil {
		return nil, err
	}

	result, err := s.forceDeleteUserCmd.Handle(ctx, command.ForceDeleteUserCommand{
		UserID: req.UserID,
	})
	if err != nil {
		return nil, err
	}

	// DB is the source of truth — once the cascade commits, any avatar
	// objects in S3 are orphans. Delete them best-effort; a failure here is
	// recoverable by a janitor sweep, so we log and move on rather than
	// surfacing the error to the caller.
	if s.storageProvider != nil {
		for _, key := range result.AvatarKeys {
			if key == "" {
				continue
			}
			log.Infof("Delete avatar key: %s", key)
			if delErr := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: key}); delErr != nil {
				log.Warnf("user.force_delete avatar cleanup failed key=%s err=%v", key, delErr)
			}
		}
	}

	return &dto.DeleteUserRes{}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *dto.UpdateUserReq) (*dto.UpdateUserRes, error) {
	if err := ValidateUpdateUser(ctx, req); err != nil {
		return nil, err
	}

	existsUser, err := s.getUserByUserIdQuery.Handle(ctx, query.GetUserByUserIdQuery{UserId: req.UserID})
	if err != nil {
		return nil, err
	}

	if existsUser == nil {
		return nil, errs.NewError(ctx, status.NOT_FOUND, nil, nil)
	}

	// Resolve avatar source — same mutex as create. Nil avatarKey here
	// means "leave avatar_key unchanged"; non-nil means replace.
	var avatarKey *string
	switch {
	case req.Avatar != nil:
		key, err := s.normalizeAvatarKey(ctx, *req.Avatar, status.USER_AVATAR_INVALID_REFERENCE)
		if err != nil {
			return nil, err
		}
		avatarKey = &key
	case req.AvatarFile != nil:
		key, err := s.updateAvatarIfPresent(ctx, req)
		if err != nil {
			return nil, err
		}
		avatarKey = key
	}

	user, err := s.updateUserCmd.Handle(ctx, command.UpdateUserCommand{
		ID:        req.ID,
		UserID:    req.UserID,
		UserName:  req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		AvatarKey: avatarKey,
	})
	if err != nil {
		return nil, err
	}

	// Delete the previous S3 object when the key actually changes. We
	// do this for BOTH source paths: an upload supersedes the old, and
	// a reference change also supersedes the old. If the new reference
	// happens to point at the same key (no-op rewrite), the equality
	// guard below prevents accidental deletion.
	if existsUser.AvatarKey() != nil && *existsUser.AvatarKey() != "" &&
		avatarKey != nil && *existsUser.AvatarKey() != *avatarKey {
		if err := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{
			Key: *existsUser.AvatarKey(),
		}); err != nil {
			return nil, err
		}
	}

	userRes := dto.DomainToResponse(user)
	s.populateImageUrl(ctx, userRes)

	return &dto.UpdateUserRes{User: userRes}, nil
}

// UploadAvatar streams the multipart body to S3 then persists the
// resulting key against the user inside a UnitOfWork. Returns the key
// plus a short-lived presigned URL for immediate display. Mirrors the
// profile module's UploadAvatar so the mobile client can reuse its
// uploader for both endpoints.
func (s *Service) UploadAvatar(ctx context.Context, userID int64, filename, contentType string, file io.Reader) (*dto.UploadAvatarRes, error) {
	if userID == 0 {
		return nil, errs.NewError(ctx, status.USER_NOT_FOUND, nil,
			ErrUserIDRequired)
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageAdapterNotConfigured)
	}
	if file == nil || filename == "" {
		return nil, errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil,
			ErrAvatarFileRequired)
	}

	// Verify the user exists BEFORE uploading so we don't leave orphan
	// S3 objects when the caller passes a bogus user_id.
	existing, err := s.getUserByUserIdQuery.Handle(ctx, query.GetUserByUserIdQuery{UserId: userID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.USER_NOT_FOUND, nil,
			ErrUserNotFound)
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    filename,
		ContentType: contentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        file,
		Filename:    filename,
		ContentType: contentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.USER_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.USER_AVATAR_UPLOAD_FAILED, nil,
			ErrUploadReturnedEmptyKey)
	}

	if err := s.setAvatarKeyCmd.Handle(ctx, command.SetAvatarKeyCommand{
		UserID:    userID,
		AvatarKey: uploaded.Key,
	}); err != nil {
		// Best-effort cleanup of the orphaned S3 object — we ignore
		// errors here since the original DB write failure is what the
		// caller cares about.
		_ = s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: uploaded.Key})
		return nil, err
	}

	signed, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        uploaded.Key,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		// Key is persisted; failing here just means we can't return a
		// preview URL now. Log and return the key without it — the
		// client can re-fetch via /users/me to get a fresh presigned URL.
		logger.From(ctx).Warnf("user.avatar presign failed user_id=%d err=%v", userID, err)
		signed = ""
	}

	logger.From(ctx).Info("user.avatar_uploaded",
		"user_id", userID,
		"avatar_key", uploaded.Key,
	)

	return &dto.UploadAvatarRes{
		UserID:    userID,
		AvatarKey: uploaded.Key,
		AvatarUrl: signed,
	}, nil
}
