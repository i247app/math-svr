package profile

import (
	"context"
	"io"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/profile"
	dto "math-ai.com/math-ai/internal/application/dto/profile"
	query "math-ai.com/math-ai/internal/application/query/profile"
	"math-ai.com/math-ai/internal/application/transaction"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	domain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	schoolDomain "math-ai.com/math-ai/internal/domain/school"
	semesterDomain "math-ai.com/math-ai/internal/domain/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
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
	getProfileByIdQuery   *query.GetProfileByIdQueryHandler
	listProfilesQuery     *query.ListProfilesQueryHandler
	createProfileCmd      *command.CreateProfileCommandHandler
	updateProfileCmd      *command.UpdateProfileCommandHandler
	softDeleteProfileCmd  *command.SoftDeleteProfileCommandHandler
	forceDeleteProfileCmd *command.ForceDeleteProfileCommandHandler
	setAvatarKeyCmd       *command.SetAvatarKeyCommandHandler
	assignSchoolCmd       *command.AssignSchoolCommandHandler
	removeSchoolCmd       *command.RemoveSchoolCommandHandler
	repo                  domain.IRepository
	programRepo           programDomain.IRepository
	gradeRepo             gradeDomain.IRepository
	semesterRepo          semesterDomain.IRepository
	schoolRepo            schoolDomain.IRepository
	storageProvider       *storage.Adapter
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
	schoolRepo schoolDomain.IRepository,
) *Service {
	return &Service{
		getProfileByIdQuery:   query.NewGetProfileByIdQueryHandler(repo),
		listProfilesQuery:     query.NewListProfilesQueryHandler(repo),
		createProfileCmd:      command.NewCreateProfileCommandHandler(uow),
		updateProfileCmd:      command.NewUpdateProfileCommandHandler(uow),
		softDeleteProfileCmd:  command.NewSoftDeleteProfileCommandHandler(uow),
		forceDeleteProfileCmd: command.NewForceDeleteProfileCommandHandler(uow),
		setAvatarKeyCmd:       command.NewSetAvatarKeyCommandHandler(uow),
		assignSchoolCmd:       command.NewAssignSchoolCommandHandler(uow),
		removeSchoolCmd:       command.NewRemoveSchoolCommandHandler(uow),
		repo:                  repo,
		programRepo:           programRepo,
		gradeRepo:             gradeRepo,
		semesterRepo:          semesterRepo,
		schoolRepo:            schoolRepo,
		storageProvider:       storageProvider,
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
			ErrProfileNotFound)
	}

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{p})
	if err != nil {
		return nil, err
	}
	return &dto.GetProfileByIdRes{Profile: responses[0]}, nil
}

// ListProfiles paginates over active profiles using the optional filters
// in req. composeProfileResponses hydrates embedded program/grade/
// semester/school objects and signs the avatar URL in the same pass, so
// the handler doesn't need to know about reference batching.
func (s *Service) ListProfiles(ctx context.Context, req *dto.ListProfilesReq) (*dto.ListProfilesRes, error) {
	if err := ValidateListProfiles(ctx, req); err != nil {
		return nil, err
	}

	profiles, pg, err := s.listProfilesQuery.Handle(ctx, query.ListProfilesQuery{
		UserID:        req.UserID,
		Role:          req.Role,
		ProfileStatus: req.ProfileStatus,
		SchoolID:      req.SchoolID,
		ProgramID:     req.ProgramID,
		GradeID:       req.GradeID,
		SemesterID:    req.SemesterID,
		IsDefault:     req.IsDefault,
		Search:        req.Search,
		Page:          req.Page,
		Limit:         req.Size,
	})
	if err != nil {
		return nil, err
	}

	responses, err := s.composeProfileResponses(ctx, profiles)
	if err != nil {
		return nil, err
	}

	return &dto.ListProfilesRes{
		Profiles:   responses,
		Pagination: pg,
	}, nil
}

func (s *Service) CreateProfile(ctx context.Context, req *dto.CreateProfileReq) (*dto.CreateProfileRes, error) {
	log := logger.From(ctx)

	if err := ValidateCreateProfile(ctx, req); err != nil {
		return nil, err
	}

	// Two mutually-exclusive avatar sources (validator enforced "at
	// most one"): a client-supplied reference (URL or S3 key) or a
	// multipart file upload. The string path is the new flow; it
	// avoids the S3 round-trip entirely.
	var (
		avatarKey      *string
		uploadedOnThis bool
	)
	switch {
	case strings.TrimSpace(req.Avatar) != "":
		key, err := s.normalizeAvatarKey(ctx, req.Avatar, status.PROFILE_AVATAR_INVALID_REFERENCE)
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
		Phone:      req.Phone,
		Email:      req.Email,
		Role:       req.Role,
		IsDefault:  req.IsDefault,
		Dob:        &dob,
		SchoolID:   req.SchoolID,
		ProgramID:  req.ProgramID,
		GradeID:    req.GradeID,
		SemesterID: req.SemesterID,
		IDType:     req.IDType,
		TeacherID:  req.TeacherID,
		StudentID:  req.StudentID,
		Note:       req.Note,
		AvatarKey:  avatarKey,
	})
	if err != nil {
		// Only delete S3 objects we just uploaded — a referenced key
		// belongs to the client's prior upload and we must not GC it.
		if uploadedOnThis && avatarKey != nil && s.storageProvider != nil {
			if delErr := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: *avatarKey}); delErr != nil {
				log.Warnf("profile.create avatar orphan cleanup failed key=%s err=%v", *avatarKey, delErr)
			}
		}
		return nil, err
	}

	log.Info("profile.created",
		"profile_id", created.ProfileId(),
		"user_id", created.UserId(),
		"name_len", len(created.Name()),
	)

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{created})
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
			ErrProfileNotFound)
	}

	// Same mutex as create — string ref OR file upload, never both.
	// Nil avatarKey means "leave avatar_key unchanged".
	var avatarKey *string
	switch {
	case req.Avatar != nil:
		key, err := s.normalizeAvatarKey(ctx, *req.Avatar, status.PROFILE_AVATAR_INVALID_REFERENCE)
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
		Phone:      req.Phone,
		Email:      req.Email,
		Role:       req.Role,
		IsDefault:  req.IsDefault,
		Dob:        &dob,
		SchoolID:   req.SchoolID,
		ProgramID:  req.ProgramID,
		GradeID:    req.GradeID,
		SemesterID: req.SemesterID,
		IDType:     req.IDType,
		TeacherID:  req.TeacherID,
		StudentID:  req.StudentID,
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

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{updated})
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
			ErrProfileNotFound)
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
			ErrProfileNotFound)
	}

	p, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: req.ProfileID})
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileNotFound)
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
func (s *Service) UploadAvatar(ctx context.Context, profileID int64, filename, contentType string, file io.Reader) (*dto.UploadAvatarRes, error) {
	if profileID == 0 {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileIDRequired)
	}
	if s.storageProvider == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			ErrStorageAdapterNotConfigured)
	}
	if file == nil || filename == "" {
		return nil, errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil,
			ErrAvatarFileRequired)
	}

	// Verify the profile exists *before* uploading so we don't leave
	// orphan S3 objects when the caller passes a bogus profile_id.
	existing, err := s.getProfileByIdQuery.Handle(ctx, query.GetProfileByIdQuery{ProfileId: profileID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileNotFound)
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
			ErrUploadReturnedEmptyKey)
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
		logger.From(ctx).Warnf("profile.avatar presign failed profile_id=%d err=%v", profileID, err)
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

// AssignSchool links a school to the profile. Both rows are looked up
// inside the same transaction so concurrent soft-deletes of either side
// fail the assignment cleanly.
func (s *Service) AssignSchool(ctx context.Context, req *dto.AssignSchoolReq) (*dto.AssignSchoolRes, error) {
	if err := ValidateAssignSchool(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.assignSchoolCmd.Handle(ctx, command.AssignSchoolCommand{
		ProfileID: req.ProfileID,
		SchoolID:  req.SchoolID,
	})
	if err != nil {
		return nil, err
	}

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{updated})
	if err != nil {
		return nil, err
	}
	return &dto.AssignSchoolRes{Profile: responses[0]}, nil
}

// RemoveSchool clears profile.school_id. Idempotent — calling it on a
// profile with no school link is a no-op write and a successful return.
func (s *Service) RemoveSchool(ctx context.Context, req *dto.RemoveSchoolReq) (*dto.RemoveSchoolRes, error) {
	if err := ValidateRemoveSchool(ctx, req); err != nil {
		return nil, err
	}

	updated, err := s.removeSchoolCmd.Handle(ctx, command.RemoveSchoolCommand{
		ProfileID: req.ProfileID,
	})
	if err != nil {
		return nil, err
	}

	responses, err := s.composeProfileResponses(ctx, []*domain.Profile{updated})
	if err != nil {
		return nil, err
	}
	return &dto.RemoveSchoolRes{Profile: responses[0]}, nil
}

// collectRefIds extracts the distinct program/grade/semester/school UUIDs
// across the given profiles. nil ids are skipped — the columns are
// nullable in schema and the domain entity carries through that nil.
func collectRefIds(profiles []*domain.Profile) (progIds, gradeIds, semIds, schoolIds []int64) {
	progSeen := make(map[int64]struct{}, len(profiles))
	gradeSeen := make(map[int64]struct{}, len(profiles))
	semSeen := make(map[int64]struct{}, len(profiles))
	schoolSeen := make(map[int64]struct{}, len(profiles))
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
		if id := p.SchoolId(); id != nil {
			if _, ok := schoolSeen[*id]; !ok {
				schoolSeen[*id] = struct{}{}
				schoolIds = append(schoolIds, *id)
			}
		}
	}
	return
}
