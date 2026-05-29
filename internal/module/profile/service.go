package profile

import (
	"context"
	"errors"
	"io"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/profile"
	dto "math-ai.com/math-ai/internal/application/dto/profile"
	query "math-ai.com/math-ai/internal/application/query/profile"
	"math-ai.com/math-ai/internal/application/transaction"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	domain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	semesterDomain "math-ai.com/math-ai/internal/domain/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const (
	// avatarFolder is the S3 prefix new avatar uploads land under.
	avatarFolder = "profile-avatars"
	// avatarUrlTTL bounds how long a generated avatar preview URL is valid.
	// Short enough that a stale link in the wild expires quickly; long
	// enough that a single screen render doesn't have to refresh mid-view.
	avatarUrlTTL = 1 * time.Hour
	// refImageUrlTTL bounds presigned URLs we hand back for embedded
	// program/grade/semester image_keys.
	refImageUrlTTL = 1 * time.Hour
)

type Service struct {
	getProfileByIdQuery       *query.GetProfileByIdQueryHandler
	listProfilesByUserIdQuery *query.ListProfilesByUserIdQueryHandler
	createProfileCmd          *command.CreateProfileCommandHandler
	updateProfileCmd          *command.UpdateProfileCommandHandler
	softDeleteProfileCmd      *command.SoftDeleteProfileCommandHandler
	forceDeleteProfileCmd     *command.ForceDeleteProfileCommandHandler
	setAvatarKeyCmd           *command.SetAvatarKeyCommandHandler
	repo                      domain.IRepository
	programRepo               programDomain.IRepository
	gradeRepo                 gradeDomain.IRepository
	semesterRepo              semesterDomain.IRepository
	storageProvider           *storage.Adapter
}

// NewService wires the profile module. storageProvider may be nil — in a
// deploy where storage is disabled, avatar-related operations return a
// targeted error, and AvatarUrl on responses degrades to nil.
//
// The program/grade/semester repos are held directly so list/get responses
// can embed the full reference objects via batched IN-lookups (no N+1).
func NewService(
	repo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	semesterRepo semesterDomain.IRepository,
) *Service {
	return &Service{
		getProfileByIdQuery:       query.NewGetProfileByIdQueryHandler(repo),
		listProfilesByUserIdQuery: query.NewListProfilesByUserIdQueryHandler(repo),
		createProfileCmd:          command.NewCreateProfileCommandHandler(uow),
		updateProfileCmd:          command.NewUpdateProfileCommandHandler(uow),
		softDeleteProfileCmd:      command.NewSoftDeleteProfileCommandHandler(uow),
		forceDeleteProfileCmd:     command.NewForceDeleteProfileCommandHandler(uow),
		setAvatarKeyCmd:           command.NewSetAvatarKeyCommandHandler(uow),
		repo:                      repo,
		programRepo:               programRepo,
		gradeRepo:                 gradeRepo,
		semesterRepo:              semesterRepo,
		storageProvider:           storageProvider,
	}
}

func (s *Service) GetProfileById(ctx context.Context, req *dto.GetProfileByIdReq) (*dto.GetProfileByIdRes, error) {
	if err := ValidateGetProfile(ctx, req); err != nil {
		return nil, err
	}

	p, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: req.ProfileID})
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{p}, req.Language)
	if err != nil {
		return nil, err
	}
	return &dto.GetProfileByIdRes{Profile: responses[0]}, nil
}

func (s *Service) ListProfilesByUserId(ctx context.Context, req *dto.ListProfilesReq) (*dto.ListProfilesRes, error) {
	if err := ValidateListProfiles(ctx, req); err != nil {
		return nil, err
	}

	profiles, err := s.listProfilesByUserIdQuery.Handle(ctx, query.ListProfilesByUserIdQuery{UserId: req.UserID})
	if err != nil {
		return nil, err
	}

	responses, err := s.composeProfileResponses(ctx, profiles, "")
	if err != nil {
		return nil, err
	}

	// responses := dto.DomainListToResponse(profiles)

	for _, resp := range responses {
		s.populateAvatarUrl(ctx, resp)
	}

	return &dto.ListProfilesRes{Profiles: responses}, nil
}

func (s *Service) CreateProfile(ctx context.Context, req *dto.CreateProfileReq) (*dto.CreateProfileRes, error) {
	log := logger.From(ctx)

	if err := ValidateCreateProfile(ctx, req); err != nil {
		return nil, err
	}

	avatarKey, err := s.uploadAvatarIfPresent(ctx, req)
	if err != nil {
		return nil, err
	}

	var dob mtime.MathTime
	if req.Dob != nil && *req.Dob != "" {
		parsed, err := mtime.ParseDate(*req.Dob)
		if err != nil {
			return nil, err
		}
		dob = parsed
	}

	created, err := s.createProfileCmd.Handle(ctx, command.CreateProfileCommand{
		UserID:     req.UserID,
		Name:       req.Name,
		Role:       req.Role,
		Dob:        &dob,
		ProgramID:  req.ProgramID,
		GradeID:    req.GradeID,
		SemesterID: req.SemesterID,
		Note:       req.Note,
		AvatarKey:  avatarKey,
	})
	if err != nil {
		return nil, err
	}

	log.Info("profile.created",
		"profile_id", created.ProfileId(),
		"user_id", created.UserId(),
		"name_len", len(created.Name()),
	)

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{created}, "")
	if err != nil {
		return nil, err
	}
	return &dto.CreateProfileRes{Profile: responses[0]}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, req *dto.UpdateProfileReq) (*dto.UpdateProfileRes, error) {
	if err := ValidateUpdateProfile(ctx, req); err != nil {
		return nil, err
	}

	existProfile, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: req.ProfileID})
	if err != nil {
		return nil, err
	}
	if existProfile == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	avatarKey, err := s.updateAvatarIfPresent(ctx, req)
	if err != nil {
		return nil, err
	}

	var dob mtime.MathTime
	if req.Dob != nil && *req.Dob != "" {
		parsed, err := mtime.ParseDate(*req.Dob)
		if err != nil {
			return nil, err
		}
		dob = parsed
	}

	updated, err := s.updateProfileCmd.Handle(ctx, command.UpdateProfileCommand{
		ProfileID:  req.ProfileID,
		Name:       req.Name,
		Dob:        &dob,
		ProgramID:  req.ProgramID,
		GradeID:    req.GradeID,
		SemesterID: req.SemesterID,
		Note:       req.Note,
		AvatarKey:  avatarKey,
	})
	if err != nil {
		return nil, err
	}

	// delete oldAvatar
	if existProfile.AvatarKey() != nil && *existProfile.AvatarKey() != "" &&
		avatarKey != nil && *existProfile.AvatarKey() != *avatarKey {
		if err := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{
			Key: *existProfile.AvatarKey(),
		}); err != nil {
			return nil, err
		}
	}

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{updated}, "")
	if err != nil {
		return nil, err
	}
	return &dto.UpdateProfileRes{Profile: responses[0]}, nil
}

func (s *Service) SoftDeleteProfile(ctx context.Context, req *dto.DeleteProfileReq) (*dto.DeleteProfileRes, error) {
	if err := ValidateDeleteProfile(ctx, req); err != nil {
		return nil, err
	}

	p, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: req.ProfileID})
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	if err := s.softDeleteProfileCmd.Handle(ctx, command.SoftDeleteProfileCommand{
		ProfileID: req.ProfileID,
	}); err != nil {
		return nil, err
	}
	return &dto.DeleteProfileRes{}, nil
}

func (s *Service) ForceDeleteProfile(ctx context.Context, req *dto.DeleteProfileReq) (*dto.DeleteProfileRes, error) {
	if err := ValidateDeleteProfile(ctx, req); err != nil {
		return nil, err
	}

	existProfile, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: req.ProfileID})
	if err != nil {
		return nil, err
	}
	if existProfile == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	p, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: req.ProfileID})
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	if err := s.forceDeleteProfileCmd.Handle(ctx, command.ForceDeleteProfileCommand{
		ProfileID: req.ProfileID,
	}); err != nil {
		return nil, err
	}

	// delete oldAvatar
	if existProfile.AvatarKey() != nil && *existProfile.AvatarKey() != "" {
		if err := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{
			Key: *existProfile.AvatarKey(),
		}); err != nil {
			return nil, err
		}
	}

	return &dto.DeleteProfileRes{}, nil
}

// UploadAvatar streams the request body to S3 then persists the resulting
// key against the profile in a single transaction. Returns the key plus a
// short-lived presigned URL for immediate display.
func (s *Service) UploadAvatar(ctx context.Context, profileID string, filename, contentType string, file io.Reader) (*dto.UploadAvatarRes, error) {
	if profileID == "" {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile_id is required"))
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			errors.New("storage adapter is not configured"))
	}
	if file == nil || filename == "" {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil,
			errors.New("avatar file is required"))
	}

	// Verify the profile exists *before* uploading so we don't leave
	// orphan S3 objects when the caller passes a bogus profile_id.
	existing, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: profileID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	if err := s.storageProvider.ValidateFileType(ctx, &storage.ValidateFileTypeRequest{
		Filename:    filename,
		ContentType: contentType,
	}); err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storageProvider.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        file,
		Filename:    filename,
		ContentType: contentType,
		Folder:      avatarFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || uploaded.Key == "" {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_UPLOAD_FAILED, nil,
			errors.New("upload returned an empty key"))
	}

	if err := s.setAvatarKeyCmd.Handle(ctx, command.SetAvatarKeyCommand{
		ProfileID: profileID,
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
		Expiration: avatarUrlTTL,
	})
	if err != nil {
		// The key is persisted; failing here just means we can't return
		// a preview URL now. Log and return the key without it.
		logger.From(ctx).Warnf("profile.avatar presign failed profile_id=%s err=%v", profileID, err)
		signed = ""
	}

	logger.From(ctx).Info("profile.avatar_uploaded",
		"profile_id", profileID,
		"avatar_key", uploaded.Key,
	)

	return &dto.UploadAvatarRes{
		ProfileID: profileID,
		AvatarKey: uploaded.Key,
		AvatarUrl: signed,
	}, nil
}

// collectRefIds extracts the distinct program/grade/semester UUIDs across the
// given profiles. uuid.Nil is skipped — defensive, since the columns are
// NOT NULL in the schema but the domain entity doesn't enforce that.
func collectRefIds(profiles []*domain.Profile) (progIds, gradeIds, semIds []string) {
	progSeen := make(map[string]struct{}, len(profiles))
	gradeSeen := make(map[string]struct{}, len(profiles))
	semSeen := make(map[string]struct{}, len(profiles))
	for _, p := range profiles {
		if id := p.ProgramId(); id != nil {
			if _, ok := progSeen[*id]; !ok {
				progSeen[*id] = struct{}{}
				progIds = append(progIds, *id)
			}
		}
		if id := p.GradeId(); id != nil {
			if _, ok := gradeSeen[*id]; !ok {
				gradeSeen[*id] = struct{}{}
				gradeIds = append(gradeIds, *id)
			}
		}
		if id := p.SemesterId(); id != nil {
			if _, ok := semSeen[*id]; !ok {
				semSeen[*id] = struct{}{}
				semIds = append(semIds, *id)
			}
		}
	}
	return
}
