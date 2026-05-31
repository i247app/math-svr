package classroom

import (
	domain "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// MemberResponse is the wire shape for ma_classroom_members rows.
// joined / left / removed / last_seen timestamps are rendered as
// strings via mtime.MathTime.String() so the client gets a consistent
// time format across the project.
type MemberResponse struct {
	ID                 int64   `json:"id"`
	MemberID           int64   `json:"member_id"`
	ClassroomID        int64   `json:"classroom_id"`
	ProfileID          int64   `json:"profile_id"`
	MemberRole         string  `json:"member_role"`
	InvitationID       *int64  `json:"invitation_id,omitempty"`
	JoinedDt           string  `json:"joined_dt,omitempty"`
	LeftDt             string  `json:"left_dt,omitempty"`
	RemovedByProfileID *int64  `json:"removed_by_profile_id,omitempty"`
	RemovedDt          string  `json:"removed_dt,omitempty"`
	LastSeenDt         string  `json:"last_seen_dt,omitempty"`
	Note               *string `json:"note,omitempty"`
	MemberStatus       *string `json:"member_status,omitempty"`
	CreateDt           string  `json:"create_dt"`
	ModifyDt           string  `json:"modify_dt"`
}

// JoinByCodeReq lets any profile join a classroom by presenting a
// non-expired invite_code. ProfileID is the caller's acting profile.
type JoinByCodeReq struct {
	ProfileID  int64  `json:"profile_id"`
	InviteCode string `json:"invite_code"`
}

type JoinByCodeRes struct {
	Member *MemberResponse `json:"member"`
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
		ID:                 m.Id(),
		MemberID:           m.MemberId(),
		ClassroomID:        m.ClassroomId(),
		ProfileID:          m.ProfileId(),
		MemberRole:         m.MemberRole(),
		InvitationID:       m.InvitationId(),
		RemovedByProfileID: m.RemovedByProfileId(),
		Note:               m.Note(),
		MemberStatus:       m.MemberStatus(),
		CreateDt:           m.CreateDt().String(),
		ModifyDt:           m.ModifyDt().String(),
	}
	if m.JoinedDt().IsValid() {
		resp.JoinedDt = m.JoinedDt().String()
	}
	if m.LeftDt().IsValid() {
		resp.LeftDt = m.LeftDt().String()
	}
	if m.RemovedDt().IsValid() {
		resp.RemovedDt = m.RemovedDt().String()
	}
	if m.LastSeenDt().IsValid() {
		resp.LastSeenDt = m.LastSeenDt().String()
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
