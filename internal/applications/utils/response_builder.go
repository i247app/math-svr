package utils

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	domain_contact "math-ai.com/math-ai/internal/core/domain/contact"
	domain_grade "math-ai.com/math-ai/internal/core/domain/grade"
	domain_profile "math-ai.com/math-ai/internal/core/domain/profile"
	domain_semester "math-ai.com/math-ai/internal/core/domain/semester"
	domain_user "math-ai.com/math-ai/internal/core/domain/user"
)

const (
	AvatarPresignedURLExpiration   = 24 * time.Hour
	GradePresignedURLExpiration    = 24 * time.Hour
	SemesterPresignedURLExpiration = 24 * time.Hour
)

// ResponseBuilder provides common response building utilities
type ResponseBuilder struct {
	storageService di.IStorageService
}

// NewResponseBuilder creates a new ResponseBuilder instance
func NewResponseBuilder(storageService di.IStorageService) *ResponseBuilder {
	return &ResponseBuilder{
		storageService: storageService,
	}
}

// generatePresignedURL creates a presigned URL for a file key
// Returns empty string if key is nil or empty, or if generation fails
func (r *ResponseBuilder) generatePresignedURL(ctx context.Context, key *string, expiration time.Duration) string {
	if key == nil || *key == "" {
		return ""
	}

	_, presignedURL, err := r.storageService.CreatePresignedUrl(ctx, &dto.CreatePresignedUrlRequest{
		Key:        *key,
		Expiration: expiration,
	})
	if err != nil {
		// Don't fail the request if presigned URL generation fails
		// File data is still valid, just without the presigned URL
		return ""
	}

	return presignedURL
}

// BuildUserResponse creates a UserResponse with presigned avatar URL
func (r *ResponseBuilder) BuildUserResponse(ctx context.Context, user *domain_user.User) *dto.UserResponse {
	res := dto.UserResponseFromDomain(user)

	// Generate presigned URL for avatar if exists
	if user.AvatarKey() != nil && *user.AvatarKey() != "" {
		presignedURL := r.generatePresignedURL(ctx, user.AvatarKey(), AvatarPresignedURLExpiration)
		if presignedURL != "" {
			res.AvatarURL = &presignedURL
		}
	}

	return &res
}

// BuildUserResponses creates UserResponses with presigned avatar URLs
func (r *ResponseBuilder) BuildUserResponses(ctx context.Context, users []*domain_user.User) []*dto.UserResponse {
	responses := make([]*dto.UserResponse, len(users))

	for i, user := range users {
		responses[i] = r.BuildUserResponse(ctx, user)
	}

	return responses
}

// BuildGradeResponse creates a GradeResponse with presigned icon URL
func (r *ResponseBuilder) BuildGradeResponse(ctx context.Context, grade *domain_grade.Grade) *dto.GradeResponse {
	res := dto.GradeResponseFromDomain(grade)

	// Generate presigned URL for icon if exists
	if grade.ImageKey() != nil && *grade.ImageKey() != "" {
		presignedURL := r.generatePresignedURL(ctx, grade.ImageKey(), GradePresignedURLExpiration)
		if presignedURL != "" {
			res.ImageUrl = &presignedURL
		}
	}

	return &res
}

// BuildGradeResponses creates GradeResponses with presigned icon URLs
func (r *ResponseBuilder) BuildGradeResponses(ctx context.Context, grades []*domain_grade.Grade) []*dto.GradeResponse {
	responses := make([]*dto.GradeResponse, len(grades))

	for i, grade := range grades {
		responses[i] = r.BuildGradeResponse(ctx, grade)
	}

	return responses
}

// BuildProfileResponse creates a ProfileResponse with presigned avatar URL
func (r *ResponseBuilder) BuildProfileResponse(ctx context.Context, profile *domain_profile.Profile) *dto.ProfileResponse {
	res := dto.ProfileResponseFromDomain(profile)

	// Generate presigned URL for avatar if exists
	if profile.AvatarKey() != nil && *profile.AvatarKey() != "" {
		presignedURL := r.generatePresignedURL(ctx, profile.AvatarKey(), AvatarPresignedURLExpiration)
		if presignedURL != "" {
			res.AvatarPreviewURL = &presignedURL
		}
	}

	return &res
}

// BuildSemesterResponse creates a SemesterResponse with presigned icon URL
func (r *ResponseBuilder) BuildSemesterResponse(ctx context.Context, semester *domain_semester.Semester) *dto.SemesterResponse {
	res := dto.SemesterResponseFromDomain(semester)

	// Generate presigned URL for icon if exists
	if semester.ImageKey() != nil && *semester.ImageKey() != "" {
		presignedURL := r.generatePresignedURL(ctx, semester.ImageKey(), SemesterPresignedURLExpiration)
		if presignedURL != "" {
			res.ImageUrl = &presignedURL
		}
	}

	return &res
}

// BuildSemesterResponses creates SemesterResponses with presigned icon URLs
func (r *ResponseBuilder) BuildSemesterResponses(ctx context.Context, semesters []*domain_semester.Semester) []*dto.SemesterResponse {
	responses := make([]*dto.SemesterResponse, len(semesters))

	for i, semester := range semesters {
		responses[i] = r.BuildSemesterResponse(ctx, semester)
	}

	return responses
}

func (r *ResponseBuilder) BuildContactUsResponses(ctx context.Context, contacts []*domain_contact.Contact) []*dto.ContactResponse {
	responses := make([]*dto.ContactResponse, len(contacts))

	for i, contact := range contacts {
		responses[i] = r.BuildContactUsResponse(ctx, contact)
	}

	return responses
}

func (r *ResponseBuilder) BuildContactUsResponse(ctx context.Context, contact *domain_contact.Contact) *dto.ContactResponse {
	res := dto.ContactUsResponseFromDomain(contact)

	return &res
}
