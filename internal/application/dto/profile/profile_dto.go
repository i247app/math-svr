package profile

import (
	"io"

	gradeDto "math-ai.com/math-ai/internal/application/dto/grade"
	programDto "math-ai.com/math-ai/internal/application/dto/program"
	schoolDto "math-ai.com/math-ai/internal/application/dto/school"
	semesterDto "math-ai.com/math-ai/internal/application/dto/semester"
	domain "math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ProfileResponse embeds the full Program / Grade / Semester objects so the
// client renders them without a follow-up call. The raw *_id fields live on
// each embedded object — keeping them on the parent too would be redundant.
type ProfileResponse struct {
	ID            int64                         `json:"id"`
	ProfileID     int64                         `json:"profile_id"`
	UserID        int64                         `json:"user_id"`
	Name          string                        `json:"name"`
	Role          string                        `json:"role"`
	AvatarKey     *string                       `json:"avatar_key,omitempty"`
	AvatarUrl     *string                       `json:"avatar_url"` // pre-signed url from avatar_key
	Dob           string                        `json:"dob,omitempty"`
	SchoolID      *int64                        `json:"school_id,omitempty"`
	School        *schoolDto.SchoolResponse     `json:"school,omitempty"`
	ProgramID     *int64                        `json:"program_id,omitempty"`
	Program       *programDto.ProgramResponse   `json:"program,omitempty"`
	GradeID       *int64                        `json:"grade_id,omitempty"`
	Grade         *gradeDto.GradeResponse       `json:"grade,omitempty"`
	SemesterID    *int64                        `json:"semester_id,omitempty"`
	Semester      *semesterDto.SemesterResponse `json:"semester,omitempty"`
	IsDefault     bool                          `json:"is_default"`
	IDType        *string                       `json:"id_type,omitempty"`
	TeacherID     *string                       `json:"teacher_id,omitempty"`
	StudentID     *string                       `json:"student_id,omitempty"`
	ProfileStatus *string                       `json:"profile_status,omitempty"`
	CreateDt      string                        `json:"create_dt"`
	ModifyDt      string                        `json:"modify_dt"`
}

type CreateProfileReq struct {
	UserID     int64   `json:"user_id"`
	Name       string  `json:"name"`
	Role       string  `json:"role"`
	IsDefault  bool    `json:"is_default"`
	Dob        *string `json:"dob,omitempty"`
	SchoolID   *int64  `json:"school_id,omitempty"`
	ProgramID  *int64  `json:"program_id"`
	GradeID    *int64  `json:"grade_id"`
	SemesterID *int64  `json:"semester_id"`
	IDType     *string `json:"id_type,omitempty"`    // TEACHER only — e.g. MOET, PUBLIC_ID
	TeacherID  *string `json:"teacher_id,omitempty"` // TEACHER only
	StudentID  *string `json:"student_id,omitempty"` // STUDENT only
	Note       *string `json:"note,omitempty"`

	// Avatar is a client-supplied reference to an object already in our
	// storage (bare S3 key or full URL pointing at the bucket). Normalized
	// to a canonical key server-side. Mutually exclusive with AvatarFile.
	Avatar string `json:"avatar,omitempty"`

	AvatarFile        io.Reader `json:"-"` // multipart file reader
	AvatarFilename    string    `json:"-"` // original filename
	AvatarContentType string    `json:"-"` // MIME type
}

type CreateProfileRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type UpdateProfileReq struct {
	ProfileID  int64   `json:"profile_id"`
	Name       *string `json:"name,omitempty"`
	Role       *string `json:"role,omitempty"`
	IsDefault  *bool   `json:"is_default,omitempty"`
	Dob        *string `json:"dob,omitempty"`
	SchoolID   *int64  `json:"school_id,omitempty"`
	ProgramID  *int64  `json:"program_id,omitempty"`
	GradeID    *int64  `json:"grade_id,omitempty"`
	SemesterID *int64  `json:"semester_id,omitempty"`
	IDType     *string `json:"id_type,omitempty"`    // TEACHER only — e.g. MOET, PUBLIC_ID
	TeacherID  *string `json:"teacher_id,omitempty"` // TEACHER only
	StudentID  *string `json:"student_id,omitempty"` // STUDENT only
	Note       *string `json:"note,omitempty"`

	// Avatar is a client-supplied reference. Pointer semantics: nil =
	// leave avatar_key untouched, non-nil = replace. Mutually exclusive
	// with AvatarFile.
	Avatar *string `json:"avatar,omitempty"`

	AvatarFile        io.Reader `json:"-"` // multipart file reader
	AvatarFilename    string    `json:"-"` // original filename
	AvatarContentType string    `json:"-"` // MIME type
}

type UpdateProfileRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type GetProfileByIdReq struct {
	ProfileID int64             `json:"profile_id"`
	Language  enum.LanguageType `json:"language,omitempty"`
}

type GetProfileByIdRes struct {
	Profile *ProfileResponse `json:"profile"`
}

// ListProfilesReq powers /profiles/list. All filter fields are optional —
// callers can scope by user_id (the original behaviour) or by any
// combination of role, profile_status, school/program/grade/semester ids,
// is_default flag, and a free-text search against name. Pagination is
// standard (page / size, defaults applied by pagination.NewPagination).
type ListProfilesReq struct {
	UserID        *int64            `json:"user_id,omitempty"`
	Role          *string           `json:"role,omitempty"`
	ProfileStatus *string           `json:"profile_status,omitempty"`
	SchoolID      *int64            `json:"school_id,omitempty"`
	ProgramID     *int64            `json:"program_id,omitempty"`
	GradeID       *int64            `json:"grade_id,omitempty"`
	SemesterID    *int64            `json:"semester_id,omitempty"`
	IsDefault     *bool             `json:"is_default,omitempty"`
	Search        *string           `json:"search,omitempty"`
	Page          int64             `json:"page"`
	Size          int64             `json:"size"`
	Language      enum.LanguageType `json:"language,omitempty"`
}

type ListProfilesRes struct {
	Profiles   []*ProfileResponse     `json:"profiles"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type DeleteProfileReq struct {
	ProfileID int64 `json:"profile_id"`
}

type DeleteProfileRes struct{}

// AssignSchoolReq and RemoveSchoolReq are dedicated to the school-link
// flow. The standard /profiles/update endpoint can also set school_id
// (via COALESCE on a non-nil pointer), but cannot clear it — a JSON
// nil there means "leave unchanged". These endpoints carry a single
// well-typed intent: assign overwrites, remove clears.
type AssignSchoolReq struct {
	ProfileID int64 `json:"profile_id"`
	SchoolID  int64 `json:"school_id"`
}

type AssignSchoolRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type RemoveSchoolReq struct {
	ProfileID int64 `json:"profile_id"`
}

type RemoveSchoolRes struct {
	Profile *ProfileResponse `json:"profile"`
}

type UploadAvatarRes struct {
	ProfileID int64  `json:"profile_id"`
	AvatarKey string `json:"avatar_key"`
	AvatarUrl string `json:"avatar_url"`
}

// UploadFileRequest represents a file upload request
type UploadFileRequest struct {
	File        io.Reader
	Filename    string
	ContentType string
	Folder      string
}

type UploadFileResponse struct {
	URL        string `json:"url"`         // Original S3 URL
	PreviewURL string `json:"preview_url"` // Preview URL for UI display
	Key        string `json:"key"`         // S3 object key
	Filename   string `json:"filename"`    // Original filename
	Size       int64  `json:"size"`        // File size in bytes
}

type DeleteFileRequest struct {
	Key string `json:"key"` // S3 object key or full URL
}

type DeleteFileResponse struct {
	Message string `json:"message"`
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
		Role:          p.Role(),
		AvatarKey:     p.AvatarKey(),
		Dob:           p.Dob().String(),
		SchoolID:      p.SchoolId(),
		ProgramID:     p.ProgramId(),
		GradeID:       p.GradeId(),
		SemesterID:    p.SemesterId(),
		IsDefault:     p.IsDefault(),
		IDType:        p.IdType(),
		TeacherID:     p.TeacherId(),
		StudentID:     p.StudentId(),
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
