package profile

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/adapter/storage"
	gradeDto "math-ai.com/math-ai/internal/application/dto/grade"
	dto "math-ai.com/math-ai/internal/application/dto/profile"
	programDto "math-ai.com/math-ai/internal/application/dto/program"
	schoolDto "math-ai.com/math-ai/internal/application/dto/school"
	semesterDto "math-ai.com/math-ai/internal/application/dto/semester"
	domain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
)

// uploadAvatarIfPresent ships the multipart avatar (if any) to S3 and returns
// the resulting key. Returns (nil, nil) when no avatar was submitted. The
// key lands on ma_users.avatar_key via BuildUser inside the create
// transaction.
func (s *Service) uploadAvatarIfPresent(ctx context.Context, req *dto.CreateProfileReq) (*string, error) {
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

func (s *Service) updateAvatarIfPresent(ctx context.Context, req *dto.UpdateProfileReq) (*string, error) {
	log := logger.From(ctx)
	log.Info("Starting uploadAvatarIfPresent")

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

	log.Info("avatar.uploaded",
		"profile_id", req.ProfileID,
		"avatar_key", uploaded.Key,
	)
	return &uploaded.Key, nil
}

// normalizeAvatarKey resolves a client-supplied avatar reference into a
// canonical S3 key under profile-avatars/. Host allowlist + prefix
// enforcement live in the storage provider; the service layer just
// translates a plain error into a typed MathError.
func (s *Service) normalizeAvatarKey(ctx context.Context, raw string, invalidStatus status.StatusCode) (string, error) {
	if s.storageProvider == nil {
		return "", errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil,
			errors.New("storage adapter is not configured"))
	}
	key, err := s.storageProvider.NormalizeKey(ctx, &storage.NormalizeKeyRequest{
		Raw: raw,
	})
	if err != nil {
		return "", errs.NewError(ctx, invalidStatus, nil, err)
	}
	if key == "" {
		return "", errs.NewError(ctx, invalidStatus, nil,
			errors.New("avatar reference resolved to empty key"))
	}
	return key, nil
}

// composeProfileResponses resolves embedded program/grade/semester/school
// objects for the given profiles using batched IN-queries — regardless of
// how many profiles are passed in. The shape is:
//
//	1 profile fetch (caller's responsibility, already done) +
//	1 programs-by-ids + 1 grades-by-ids + 1 semesters-by-ids +
//	1 schools-by-ids (skipped when no profile references a school, and
//	when the school repo isn't wired)
//	= up to 5 round-trips total, never O(N).
//
// AvatarUrl and each ref's ImageUrl are signed in the same pass.
// School lookups tolerate a missing/soft-deleted school silently —
// resp.School stays nil, resp.SchoolID is preserved so the client can
// still tell what the profile refers to.
func (s *Service) composeProfileResponses(ctx context.Context, profiles []*domain.Profile, lang enum.LanguageType) ([]*dto.ProfileResponse, error) {
	if len(profiles) == 0 {
		return []*dto.ProfileResponse{}, nil
	}

	language := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	if lang != "" {
		language = lang
	}

	progIds, gradeIds, semIds, schoolIds := collectRefIds(profiles)

	programs, err := s.programRepo.ListProgramsByIds(ctx, progIds, language)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	grades, err := s.gradeRepo.ListGradesByIds(ctx, gradeIds, language)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	semesters, err := s.semesterRepo.ListSemestersByIds(ctx, semIds, language)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	progMap := make(map[string]*programDto.ProgramResponse, len(programs))
	for _, p := range programs {
		r := programDto.DomainToResponse(p)
		s.signProgramImageUrl(ctx, r)
		progMap[p.ProgramId()] = r
	}
	gradeMap := make(map[string]*gradeDto.GradeResponse, len(grades))
	for _, g := range grades {
		r := gradeDto.DomainToResponse(g)
		s.signGradeImageUrl(ctx, r)
		gradeMap[g.GradeId()] = r
	}
	semMap := make(map[string]*semesterDto.SemesterResponse, len(semesters))
	for _, sem := range semesters {
		r := semesterDto.DomainToResponse(sem)
		s.signSemesterImageUrl(ctx, r)
		semMap[sem.SemesterId()] = r
	}

	schoolMap := make(map[string]*schoolDto.SchoolResponse)
	if len(schoolIds) > 0 && s.schoolRepo != nil {
		schools, err := s.schoolRepo.ListSchoolsByIds(ctx, schoolIds)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		for _, sc := range schools {
			r := schoolDto.DomainToResponse(sc)
			s.signSchoolImageUrl(ctx, r)
			schoolMap[sc.SchoolId()] = r
		}
	}

	out := make([]*dto.ProfileResponse, len(profiles))
	for i, p := range profiles {
		resp := dto.DomainToResponse(p)
		if p.ProgramId() != nil {
			resp.Program = progMap[*p.ProgramId()]
		}
		if p.GradeId() != nil {
			resp.Grade = gradeMap[*p.GradeId()]
		}
		if p.SemesterId() != nil {
			resp.Semester = semMap[*p.SemesterId()]
		}
		if p.SchoolId() != nil {
			resp.School = schoolMap[*p.SchoolId()]
		}
		s.populateAvatarUrl(ctx, resp)
		out[i] = resp
	}
	return out, nil
}

// populateAvatarUrl mutates resp in place to add a short-lived presigned
// URL when the profile carries an avatar_key. No-op if storage is disabled
// or the profile has no avatar.
func (s *Service) populateAvatarUrl(ctx context.Context, resp *dto.ProfileResponse) {
	if resp == nil || s.storageProvider == nil || resp.AvatarKey == nil || *resp.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.AvatarKey,
		Expiration: avatarUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("profile.avatar presign failed profile_id=%s err=%v", resp.ProfileID, err)
		return
	}
	resp.AvatarUrl = &url
}

func (s *Service) signProgramImageUrl(ctx context.Context, r *programDto.ProgramResponse) {
	if r == nil || s.storageProvider == nil || r.ImageKey == nil || *r.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *r.ImageKey,
		Expiration: refImageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("profile.program image presign failed program_id=%s err=%v", r.ProgramID, err)
		return
	}
	r.ImageUrl = &url
}

func (s *Service) signGradeImageUrl(ctx context.Context, r *gradeDto.GradeResponse) {
	if r == nil || s.storageProvider == nil || r.ImageKey == nil || *r.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *r.ImageKey,
		Expiration: refImageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("profile.grade image presign failed grade_id=%s err=%v", r.GradeID, err)
		return
	}
	r.ImageUrl = &url
}

func (s *Service) signSchoolImageUrl(ctx context.Context, r *schoolDto.SchoolResponse) {
	if r == nil || s.storageProvider == nil || r.ImageKey == nil || *r.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *r.ImageKey,
		Expiration: refImageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("profile.school image presign failed school_id=%s err=%v", r.SchoolID, err)
		return
	}
	r.ImageUrl = &url
}

func (s *Service) signSemesterImageUrl(ctx context.Context, r *semesterDto.SemesterResponse) {
	if r == nil || s.storageProvider == nil || r.ImageKey == nil || *r.ImageKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *r.ImageKey,
		Expiration: refImageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("profile.semester image presign failed semester_id=%s err=%v", r.SemesterID, err)
		return
	}
	r.ImageUrl = &url
}
