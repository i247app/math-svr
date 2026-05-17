package profile

import (
	"github.com/google/uuid"
	gradeDto "math-ai.com/math-ai/internal/application/dto/grade"
	programDto "math-ai.com/math-ai/internal/application/dto/program"
	semesterDto "math-ai.com/math-ai/internal/application/dto/semester"
	domain "math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ProfileResponse embeds the full Program / Grade / Semester objects so the
// client renders them without a follow-up call. The raw *_id fields live on
// each embedded object — keeping them on the parent too would be redundant.
type ProfileResponse struct {
	ID            int64                         `json:"id"`
	ProfileID     uuid.UUID                     `json:"profile_id"`
	UserID        uuid.UUID                     `json:"user_id"`
	Name          string                        `json:"name"`
	AvatarKey     *string                       `json:"avatar_key,omitempty"`
	AvatarUrl     *string                       `json:"avatar_url"` // pre-signed url from avatar_key
	Dob           string                        `json:"dob,omitempty"`
	ProgramID     uuid.UUID                     `json:"program_id,omitempty"`
	Program       *programDto.ProgramResponse   `json:"program,omitempty"`
	GradeID       uuid.UUID                     `json:"grade_id,omitempty"`
	Grade         *gradeDto.GradeResponse       `json:"grade,omitempty"`
	SemesterID    uuid.UUID                     `json:"semester_id,omitempty"`
	Semester      *semesterDto.SemesterResponse `json:"semester,omitempty"`
	ProfileStatus *string                       `json:"profile_status,omitempty"`
	CreateDt      string                        `json:"create_dt"`
	ModifyDt      string                        `json:"modify_dt"`
}

type CreateProfileReq struct {
	UserID     uuid.UUID `json:"user_id"`
	Name       string    `json:"name"`
	Dob        *string   `json:"dob,omitempty"`
	ProgramID  uuid.UUID `json:"program_id"`
	GradeID    uuid.UUID `json:"grade_id"`
	SemesterID uuid.UUID `json:"semester_id"`
	Note       *string   `json:"note,omitempty"`
}

type CreateProfileRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type UpdateProfileReq struct {
	ProfileID  uuid.UUID  `json:"profile_id"`
	Name       *string    `json:"name,omitempty"`
	Dob        *string    `json:"dob,omitempty"`
	ProgramID  *uuid.UUID `json:"program_id,omitempty"`
	GradeID    *uuid.UUID `json:"grade_id,omitempty"`
	SemesterID *uuid.UUID `json:"semester_id,omitempty"`
	Note       *string    `json:"note,omitempty"`
}

type UpdateProfileRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type GetProfileByIdReq struct {
	ProfileID uuid.UUID         `json:"profile_id"`
	Language  enum.LanguageType `json:"language,omitempty"`
}

type GetProfileByIdRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type ListProfilesReq struct {
	UserID   uuid.UUID         `json:"user_id"`
	Language enum.LanguageType `json:"language,omitempty"`
}

type ListProfilesRes struct {
	Profiles []*ProfileResponse `json:"profiles"`
}

type DeleteProfileReq struct {
	ProfileID uuid.UUID `json:"profile_id"`
}

type DeleteProfileRes struct{}

type UploadAvatarRes struct {
	ProfileID uuid.UUID `json:"profile_id"`
	AvatarKey string    `json:"avatar_key"`
	AvatarUrl string    `json:"avatar_url"`
}

func DomainToResponse(p *domain.Profile) *ProfileResponse {
	if p == nil {
		return nil
	}

	return &ProfileResponse{
		ID:            p.Id(),
		ProfileID:     p.ProfileId(),
		UserID:        p.UserId(),
		Name:          p.Name(),
		AvatarKey:     p.AvatarKey(),
		Dob:           p.Dob().String(),
		ProgramID:     p.ProgramId(),
		GradeID:       p.GradeId(),
		SemesterID:    p.SemesterId(),
		ProfileStatus: p.ProfileStatus(),
		CreateDt:      p.CreateDt().String(),
		ModifyDt:      p.ModifyDt().String(),
	}
}

func DomainListToResponse(profiles []*domain.Profile) []*ProfileResponse {
	result := make([]*ProfileResponse, len(profiles))
	for i, p := range profiles {
		result[i] = DomainToResponse(p)
	}
	return result
}
