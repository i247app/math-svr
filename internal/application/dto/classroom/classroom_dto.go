package classroom

import (
	"io"

	domain "math-ai.com/math-ai/internal/domain/classroom"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ClassroomResponse is the surface returned to API consumers. Time fields
// render via mtime.MathTime.String() so the wire format stays consistent
// with the rest of the project.
type ClassroomResponse struct {
	ID                  int64   `json:"id"`
	ClassroomID         int64   `json:"classroom_id"`
	OwnerProfileID      int64   `json:"owner_profile_id"`
	Name                string  `json:"name"`
	Description         *string `json:"description,omitempty"`
	SchoolID            *int64  `json:"school_id,omitempty"`
	ProgramIDs          []int64 `json:"program_ids"`
	GradeID             *int64  `json:"grade_id,omitempty"`
	InviteCode          *string `json:"invite_code,omitempty"`
	InviteCodeExpiresDt string  `json:"invite_code_expires_dt,omitempty"`
	MaxMembers          *int64  `json:"max_members,omitempty"`
	MemberCount         int64   `json:"member_count"`
	StudentCount        int64   `json:"student_count"`
	TeacherCount        int64   `json:"teacher_count"`
	CoverKey            *string `json:"cover_key,omitempty"`
	CoverURL            *string `json:"cover_url,omitempty"`
	Note                *string `json:"note,omitempty"`
	ClassroomStatus     *string `json:"classroom_status,omitempty"`
	CreateDt            string  `json:"create_dt"`
	ModifyDt            string  `json:"modify_dt"`
}

// CreateClassroomReq carries the caller's acting profile_id — see
// CLAUDE.md §0 Q1 resolution: every classroom mutation takes profile_id
// in the body, validated against the session's user_id at the service
// edge. ProfileID here is the *owner-to-be*.
type CreateClassroomReq struct {
	ProfileID   int64   `json:"profile_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	SchoolID    *int64  `json:"school_id,omitempty"`
	// ProgramIDs is the set of curriculum programs (books) the new
	// classroom carries. May be empty — a classroom with zero programs
	// is allowed. The validator dedupes and the command writes one
	// ma_classroom_programs row per id inside the create UoW.
	ProgramIDs []int64 `json:"program_ids,omitempty"`
	GradeID    *int64  `json:"grade_id,omitempty"`
	MaxMembers *int64  `json:"max_members,omitempty"`
	CoverKey   *string `json:"cover_key,omitempty"`
	Note       *string `json:"note,omitempty"`
	// InviteCode is an optional client-supplied join code (e.g. a
	// human-friendly token like "MATH4B"). Bounded by VARCHAR(16) and
	// must be unique across all classrooms; the command rejects a
	// duplicate with CLASSROOM_INVITE_CODE_TAKEN. Leave empty to ship a
	// classroom without a code (targeted invitations only).
	InviteCode *string `json:"invite_code,omitempty"`
	// InviteCodeExpiresDt is the optional expiry that pairs with
	// InviteCode. A zero value (omitted in JSON) means the code never
	// expires. Required to be in the future when supplied.
	InviteCodeExpiresDt mtime.MathTime `json:"invite_code_expires_dt,omitempty"`

	// File upload fields for handling cover image
	AvatarFile        io.Reader `json:"-"`
	AvatarFilename    string    `json:"-"`
	AvatarContentType string    `json:"-"`
}

type CreateClassroomRes struct {
	Classroom *ClassroomResponse `json:"classroom"`
}

// UpdateClassroomReq is a PATCH-style payload. Nil pointer = leave alone;
// non-nil value overwrites. ProfileID is the caller; ClassroomID picks
// the target.
type UpdateClassroomReq struct {
	ProfileID   int64   `json:"profile_id"`
	ClassroomID int64   `json:"classroom_id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	SchoolID    *int64  `json:"school_id,omitempty"`
	// ProgramIDs uses replace-set semantics:
	//   nil        — leave the existing program links untouched
	//   non-nil [] — remove every program link
	//   non-nil X  — make the active link set exactly X (insert new,
	//                delete removed)
	// The handler distinguishes nil from [] by whether the JSON key was
	// present, so multipart callers should omit the field entirely to
	// signal "don't touch".
	ProgramIDs *[]int64 `json:"program_ids,omitempty"`
	GradeID    *int64   `json:"grade_id,omitempty"`
	MaxMembers *int64   `json:"max_members,omitempty"`
	Note       *string  `json:"note,omitempty"`
	AvatarKey  *string  `json:"avatar_key,omitempty"`

	// File upload fields for handling cover image
	AvatarFile        io.Reader `json:"-"`
	AvatarFilename    string    `json:"-"`
	AvatarContentType string    `json:"-"`
}

type UpdateClassroomRes struct {
	Classroom *ClassroomResponse `json:"classroom"`
}

type GetClassroomReq struct {
	// ProfileID is the caller — required to membership-gate the read.
	// Populated by the handler from the request body (POST) or query
	// string (GET); the validator treats both paths the same.
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
}

type GetClassroomRes struct {
	Classroom *ClassroomResponse `json:"classroom"`
}

// ListClassroomsReq powers "my classrooms" (ProfileID set) and owner
// filters. When neither ProfileID nor OwnerProfileID is set the handler
// rejects the request — the unauthenticated "all classrooms" path is
// intentionally not offered.
type ListClassroomsReq struct {
	ProfileID      int64  `json:"profile_id"`
	OwnerProfileID *int64 `json:"owner_profile_id,omitempty"`
	SchoolID       *int64 `json:"school_id,omitempty"`
	// ProgramID and ProgramIDs are unioned (OR semantics): a classroom
	// matches when it carries any of the named programs. Either field
	// or both may be sent; ProgramID is kept for legacy single-filter
	// callers, ProgramIDs is the new multi-filter shape.
	ProgramID       *int64  `json:"program_id,omitempty"`
	ProgramIDs      []int64 `json:"program_ids,omitempty"`
	GradeID         *int64  `json:"grade_id,omitempty"`
	Search          *string `json:"search,omitempty"`
	IncludeArchived bool    `json:"include_archived"`
	Page            int64   `json:"page"`
	Size            int64   `json:"size"`
}

type ListClassroomsRes struct {
	Classrooms []*ClassroomResponse   `json:"classrooms"`
	Pagination *pagination.Pagination `json:"pagination"`
}

// ArchiveClassroomReq + RestoreClassroomReq + DeleteClassroomReq all
// share the same shape; keeping them as separate types lets us evolve
// each independently (e.g. force-delete may grow an admin-only reason
// field later) without breaking JSON contracts.
type ArchiveClassroomReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
}

type ArchiveClassroomRes struct{}

type RestoreClassroomReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
}

type RestoreClassroomRes struct{}

type DeleteClassroomReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
}

type DeleteClassroomRes struct{}

func DomainToResponse(c *domain.Classroom) *ClassroomResponse {
	if c == nil {
		return nil
	}
	programIDs := c.ProgramIds()
	if programIDs == nil {
		programIDs = []int64{}
	}
	resp := &ClassroomResponse{
		ID:              c.Id(),
		ClassroomID:     c.ClassroomId(),
		OwnerProfileID:  c.OwnerProfileId(),
		Name:            c.Name(),
		Description:     c.Description(),
		SchoolID:        c.SchoolId(),
		ProgramIDs:      programIDs,
		GradeID:         c.GradeId(),
		InviteCode:      c.InviteCode(),
		MaxMembers:      c.MaxMembers(),
		MemberCount:     c.MemberCount(),
		StudentCount:    c.StudentCount(),
		TeacherCount:    c.TeacherCount(),
		CoverKey:        c.CoverKey(),
		Note:            c.Note(),
		ClassroomStatus: c.ClassroomStatus(),
		CreateDt:        c.CreateDt().String(),
		ModifyDt:        c.ModifyDt().String(),
	}
	if c.InviteCodeExpiresDt().IsValid() {
		resp.InviteCodeExpiresDt = c.InviteCodeExpiresDt().String()
	}
	return resp
}

func DomainListToResponse(classrooms []*domain.Classroom) []*ClassroomResponse {
	result := make([]*ClassroomResponse, len(classrooms))
	for i, c := range classrooms {
		result[i] = DomainToResponse(c)
	}
	return result
}
