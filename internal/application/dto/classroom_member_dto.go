package dto

import (
	domain "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/mtime"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// MemberProfileSummary is the slim profile shape carried on a member
// response. Mirrors ClassroomOwnerSummary so list endpoints can render a
// member row without pulling the full ProfileResponse. AvatarURL is a
// short-lived presigned URL when storage is wired and the profile has
// an avatar_key; nil otherwise.
type MemberProfileSummary struct {
	ProfileID int64   `json:"profile_id"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarKey *string `json:"avatar_key,omitempty"`
	AvatarURL *string `json:"avatar_url"`
}

// MemberResponse is the wire shape for ma_classroom_members rows.
// joined / left / removed / last_seen timestamps are *mtime.MathTime
// (Format20FSP, omitted when unset) like every other time field. MemberProfile is hydrated by the
// service from ma_profiles using a batched IN-lookup so a page costs
// one extra round trip regardless of size; nil when the profile lookup
// is skipped or the profile row no longer exists.
type MemberResponse struct {
	MemberID           int64                 `json:"member_id"`
	ClassroomID        int64                 `json:"classroom_id"`
	ProfileID          int64                 `json:"profile_id"`
	MemberProfile      *MemberProfileSummary `json:"member_profile,omitempty"`
	MemberRole         string                `json:"member_role"`
	InvitationID       *int64                `json:"invitation_id,omitempty"`
	JoinedDt           *mtime.MathTime       `json:"joined_dt,omitempty"`
	LeftDt             *mtime.MathTime       `json:"left_dt,omitempty"`
	RemovedByProfileID *int64                `json:"removed_by_profile_id,omitempty"`
	RemovedDt          *mtime.MathTime       `json:"removed_dt,omitempty"`
	LastSeenDt         *mtime.MathTime       `json:"last_seen_dt,omitempty"`
	Note               *string               `json:"note,omitempty"`
	MemberStatus       *string               `json:"member_status,omitempty"`
	CreateDt           mtime.MathTime        `json:"create_dt"`
	ModifyDt           mtime.MathTime        `json:"modify_dt"`
}

// JoinByCodeReq lets any profile submit a join request to a
// classroom by presenting a non-expired classroom_code. ProfileID is
// the caller's acting profile. NOTE: this no longer joins the
// classroom directly — it creates a PENDING_REQUEST on
// ma_classroom_members that the classroom owner must approve.
type JoinByCodeReq struct {
	ProfileID     int64  `json:"profile_id"`
	ClassroomCode string `json:"classroom_code"`
}

type JoinByCodeRes struct {
	Request *JoinRequestResponse `json:"request"`
}

// LeaveClassroomReq lets the caller leave a classroom they're a member
// of. OWNER is rejected — they must transfer ownership first.
type LeaveClassroomReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
}

type LeaveClassroomRes struct{}

// RemoveMemberReq removes a target profile from a classroom. Caller
// must be OWNER (any target) or CO_TEACHER (STUDENT targets only).
type RemoveMemberReq struct {
	ProfileID       int64 `json:"profile_id"`
	ClassroomID     int64 `json:"classroom_id"`
	TargetProfileID int64 `json:"target_profile_id"`
}

type RemoveMemberRes struct{}

// UpdateMemberRoleReq promotes a STUDENT to CO_TEACHER or demotes a
// CO_TEACHER back to STUDENT. OWNER cannot be set or unset this way —
// use TransferOwnership instead.
type UpdateMemberRoleReq struct {
	ProfileID       int64  `json:"profile_id"`
	ClassroomID     int64  `json:"classroom_id"`
	TargetProfileID int64  `json:"target_profile_id"`
	NewRole         string `json:"new_role"`
}

type UpdateMemberRoleRes struct {
	Member *MemberResponse `json:"member"`
}

// TransferOwnershipReq atomically swaps OWNER between two ACTIVE
// members in the same classroom. The new owner must be a TEACHER role
// profile and an active member. The outgoing owner becomes CO_TEACHER
// so they keep manager rights but no longer block leave.
type TransferOwnershipReq struct {
	ProfileID         int64 `json:"profile_id"`
	ClassroomID       int64 `json:"classroom_id"`
	NewOwnerProfileID int64 `json:"new_owner_profile_id"`
}

type TransferOwnershipRes struct{}

// ListMembersReq enumerates members of a classroom the caller belongs
// to. Optional Role / Status filters narrow the result.
type ListMembersReq struct {
	ProfileID   int64   `json:"profile_id"`
	ClassroomID int64   `json:"classroom_id"`
	Role        *string `json:"role,omitempty"`
	Status      *string `json:"status,omitempty"`
	Page        int64   `json:"page"`
	Size        int64   `json:"size"`
}

type ListMembersRes struct {
	Members    []*MemberResponse      `json:"members"`
	Pagination *pagination.Pagination `json:"pagination"`
}

func MemberDomainToResponse(m *domain.Member) *MemberResponse {
	if m == nil {
		return nil
	}
	resp := &MemberResponse{
		MemberID:           m.MemberId(),
		ClassroomID:        m.ClassroomId(),
		ProfileID:          m.ProfileId(),
		MemberRole:         m.MemberRole(),
		InvitationID:       m.InvitationId(),
		RemovedByProfileID: m.RemovedByProfileId(),
		Note:               m.Note(),
		MemberStatus:       m.MemberStatus(),
		CreateDt:           mtime.NewTime(m.CreateDt()),
		ModifyDt:           mtime.NewTime(m.ModifyDt()),
	}
	if !m.JoinedDt().IsZero() {
		resp.JoinedDt = mtime.NewMathTimePtr(m.JoinedDt())
	}
	if !m.LeftDt().IsZero() {
		resp.LeftDt = mtime.NewMathTimePtr(m.LeftDt())
	}
	if !m.RemovedDt().IsZero() {
		resp.RemovedDt = mtime.NewMathTimePtr(m.RemovedDt())
	}
	if !m.LastSeenDt().IsZero() {
		resp.LastSeenDt = mtime.NewMathTimePtr(m.LastSeenDt())
	}
	return resp
}

func MemberDomainListToResponse(members []*domain.Member) []*MemberResponse {
	result := make([]*MemberResponse, len(members))
	for i, m := range members {
		result[i] = MemberDomainToResponse(m)
	}
	return result
}
