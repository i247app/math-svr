package dto

import (
	domain "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/mtime"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// JoinRequestResponse is the wire shape for an ma_classroom_members
// row viewed through the join-request lens (member_status =
// PENDING_REQUEST). Structurally similar to InvitationResponse but
// surfaces the requester profile rather than an inviter — there is
// no inviter on a self-initiated request.
type JoinRequestResponse struct {
	RequestID    int64            `json:"request_id"`
	ClassroomID  int64            `json:"classroom_id"`
	ProfileID    int64            `json:"profile_id"`
	Requester    *ProfileResponse `json:"requester,omitempty"`
	MemberRole   string           `json:"member_role"`
	MemberStatus *string          `json:"member_status,omitempty"`
	RequestedDt  *mtime.MathTime  `json:"requested_dt,omitempty"`
	JoinedDt     *mtime.MathTime  `json:"joined_dt,omitempty"`
	LeftDt       *mtime.MathTime  `json:"left_dt,omitempty"`
	RemovedDt    *mtime.MathTime  `json:"removed_dt,omitempty"`
	Note         *string          `json:"note,omitempty"`
	CreateDt     mtime.MathTime   `json:"create_dt"`
	ModifyDt     mtime.MathTime   `json:"modify_dt"`
}

// ApproveJoinRequestReq lets the classroom owner approve a pending
// request from a target profile. Target identified by
// (classroom_id, target_profile_id).
type ApproveJoinRequestReq struct {
	ProfileID       int64 `json:"profile_id"`
	ClassroomID     int64 `json:"classroom_id"`
	TargetProfileID int64 `json:"target_profile_id"`
}

type ApproveJoinRequestRes struct {
	Request *JoinRequestResponse `json:"request"`
}

// RejectJoinRequestReq is the owner-side rejection counterpart.
type RejectJoinRequestReq struct {
	ProfileID       int64 `json:"profile_id"`
	ClassroomID     int64 `json:"classroom_id"`
	TargetProfileID int64 `json:"target_profile_id"`
}

type RejectJoinRequestRes struct{}

// CancelJoinRequestReq lets the requester withdraw their own pending
// request. ClassroomID identifies which request — caller's profile is
// the implicit (classroom, profile) lookup key.
type CancelJoinRequestReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
}

type CancelJoinRequestRes struct{}

// ListJoinRequestsByClassroomReq enumerates pending requests for a
// classroom — owner-only.
type ListJoinRequestsByClassroomReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
	Page        int64 `json:"page"`
	Size        int64 `json:"size"`
}

type ListJoinRequestsByClassroomRes struct {
	Requests   []*JoinRequestResponse `json:"requests"`
	Pagination *pagination.Pagination `json:"pagination"`
}

// ListMyJoinRequestsReq enumerates the caller's outstanding join
// requests across every classroom.
type ListMyJoinRequestsReq struct {
	ProfileID int64 `json:"profile_id"`
	Page      int64 `json:"page"`
	Size      int64 `json:"size"`
}

type ListMyJoinRequestsRes struct {
	Requests   []*JoinRequestResponse `json:"requests"`
	Pagination *pagination.Pagination `json:"pagination"`
}

// JoinRequestDomainToResponse renders a ma_classroom_members row as
// the join-request wire shape. The optional requester is hydrated by
// the service in batch (see ListJoinRequestsByClassroom).
func JoinRequestDomainToResponse(m *domain.Member, requester *ProfileResponse) *JoinRequestResponse {
	if m == nil {
		return nil
	}
	resp := &JoinRequestResponse{
		RequestID:    m.MemberId(),
		ClassroomID:  m.ClassroomId(),
		ProfileID:    m.ProfileId(),
		MemberRole:   m.MemberRole(),
		MemberStatus: m.MemberStatus(),
		Note:         m.Note(),
		CreateDt:     mtime.NewTime(m.CreateDt()),
		ModifyDt:     mtime.NewTime(m.ModifyDt()),
	}
	if !m.InviteDt().IsZero() {
		resp.RequestedDt = mtime.NewMathTimePtr(m.InviteDt())
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
	if requester != nil {
		resp.Requester = requester
	}
	return resp
}

// JoinRequestDomainListToResponse mirrors InvitationDomainListToResponse:
// hydrate the requester via a precomputed map[profileId]*ProfileResponse
// so the service can batch the profile lookup in one query.
func JoinRequestDomainListToResponse(rows []*domain.Member, hashRequesters map[int64]*ProfileResponse) []*JoinRequestResponse {
	out := make([]*JoinRequestResponse, len(rows))
	for i, m := range rows {
		var requester *ProfileResponse
		if hashRequesters != nil {
			requester = hashRequesters[m.ProfileId()]
		}
		out[i] = JoinRequestDomainToResponse(m, requester)
	}
	return out
}
