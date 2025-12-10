package resolvers

import (
	"github.com/graphql-go/graphql"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type ProfileResolver struct {
	profileService di.IProfileService
}

func NewProfileResolver(profileService di.IProfileService) *ProfileResolver {
	return &ProfileResolver{
		profileService: profileService,
	}
}

// FetchProfile resolves fetching a user's profile
func (r *ProfileResolver) FetchProfile(params graphql.ResolveParams) (interface{}, error) {
	uid, ok := params.Args["uid"].(string)
	if !ok {
		return nil, nil
	}

	req := &dto.FetchProfileRequest{
		UID: uid,
	}

	statusCode, profile, err := r.profileService.FetchProfile(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return profile, nil
}

// CreateProfile resolves profile creation
func (r *ProfileResolver) CreateProfile(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	req := &dto.CreateProfileRequest{
		UID:        input["uid"].(string),
		GradeID:    input["grade_id"].(string),
		SemesterID: input["semester_id"].(string),
	}

	statusCode, profile, err := r.profileService.CreateProfile(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return profile, nil
}

// UpdateProfile resolves profile updates
func (r *ProfileResolver) UpdateProfile(params graphql.ResolveParams) (interface{}, error) {
	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	uid, ok := input["uid"].(string)
	if !ok {
		return nil, nil
	}

	req := &dto.UpdateProfileRequest{
		UID: uid,
	}

	if gradeID, ok := input["grade_id"].(string); ok {
		req.GradeID = &gradeID
	}
	if semesterID, ok := input["semester_id"].(string); ok {
		req.SemesterID = &semesterID
	}

	statusCode, profile, err := r.profileService.UpdateProfile(params.Context, req)
	if err != nil || statusCode != status.SUCCESS {
		return nil, err
	}

	return profile, nil
}
