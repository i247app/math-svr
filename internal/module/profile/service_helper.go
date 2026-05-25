package profile

import (
	"context"

	"math-ai.com/math-ai/internal/adapter/storage"
	gradeDto "math-ai.com/math-ai/internal/application/dto/grade"
	dto "math-ai.com/math-ai/internal/application/dto/profile"
	programDto "math-ai.com/math-ai/internal/application/dto/program"
	semesterDto "math-ai.com/math-ai/internal/application/dto/semester"
	domain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// composeProfileResponses resolves embedded program/grade/semester objects
// for the given profiles using exactly 3 batched IN-queries — regardless of
// how many profiles are passed in. The shape is:
//
//	1 profile fetch (caller's responsibility, already done) +
//	1 programs-by-ids + 1 grades-by-ids + 1 semesters-by-ids
//	= 4 round-trips total, never O(N).
//
// AvatarUrl and each ref's ImageUrl are signed in the same pass.
func (s *Service) composeProfileResponses(ctx context.Context, profiles []*domain.Profile, lang enum.LanguageType) ([]*dto.ProfileResponse, error) {
	if len(profiles) == 0 {
		return []*dto.ProfileResponse{}, nil
	}
	if lang == "" {
		lang = enum.LanguageTypeEnglish
	}

	progIds, gradeIds, semIds := collectRefIds(profiles)

	programs, err := s.programRepo.ListProgramsByIds(ctx, progIds, lang)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	grades, err := s.gradeRepo.ListGradesByIds(ctx, gradeIds, lang)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	semesters, err := s.semesterRepo.ListSemestersByIds(ctx, semIds, lang)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	progMap := make(map[string]*programDto.ProgramResponse, len(programs))
	for _, p := range programs {
		r := programDto.DomainToResponse(p)
		s.signProgramImageUrl(ctx, r)
		progMap[p.ProgramId().String()] = r
	}
	gradeMap := make(map[string]*gradeDto.GradeResponse, len(grades))
	for _, g := range grades {
		r := gradeDto.DomainToResponse(g)
		s.signGradeImageUrl(ctx, r)
		gradeMap[g.GradeId().String()] = r
	}
	semMap := make(map[string]*semesterDto.SemesterResponse, len(semesters))
	for _, sem := range semesters {
		r := semesterDto.DomainToResponse(sem)
		s.signSemesterImageUrl(ctx, r)
		semMap[sem.SemesterId().String()] = r
	}

	out := make([]*dto.ProfileResponse, len(profiles))
	for i, p := range profiles {
		resp := dto.DomainToResponse(p)
		if p.ProgramId() != nil {
			resp.Program = progMap[p.ProgramId().String()]
		}
		if p.GradeId() != nil {
			resp.Grade = gradeMap[p.GradeId().String()]
		}
		if p.SemesterId() != nil {
			resp.Semester = semMap[p.SemesterId().String()]
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
