package classroom

import (
	profileDTO "math-ai.com/math-ai/internal/application/dto/profile"
	domain "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// InvitationClassroomSummary is the slim classroom shape carried on an
// invitation response so the mobile client can render an invitation
// card (classroom name + cover) without an extra GetClassroom round
// trip. CoverURL is a short-lived presigned URL when storage is wired
// and the classroom has a cover_key; nil otherwise.
type InvitationClassroomSummary struct {
	ClassroomID     int64   `json:"classroom_id"`
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	ClassroomCode   *string `json:"classroom_code,omitempty"`
	SchoolID        *int64  `json:"school_id,omitempty"`
	GradeID         *int64  `json:"grade_id,omitempty"`
	CoverKey        *string `json:"cover_key,omitempty"`
	CoverURL        *string `json:"cover_url,omitempty"`
	ClassroomStatus *string `json:"classroom_status,omitempty"`
}

// InvitationResponse is the wire shape for an ma_classroom_members row
// viewed through the invitation lens. It re-uses the underlying member
// row but renames a few fields so a mobile client treating invitations
// as a distinct concept doesn't have to know the unified storage model.
// member_status carries PENDING / ACTIVE / REJECTED / etc — the same
// values as MemberResponse.MemberStatus. Classroom is hydrated by the
// service from ma_classrooms using a batched IN-lookup so a page costs
// one extra round trip regardless of size.
type InvitationResponse struct {
	InvitationID int64                       `json:"invitation_id"`
	ClassroomID  int64                       `json:"classroom_id"`
	Classroom    *InvitationClassroomSummary `json:"classroom,omitempty"`
	ProfileID    int64                       `json:"profile_id"`
	MemberRole   string                      `json:"member_role"`
	MemberStatus *string                     `json:"member_status,omitempty"`
	InviteBy     *int64                      `json:"invite_by,omitempty"`
	Inviter      *profileDTO.ProfileResponse `json:"inviter,omitempty"`
	InviteDt     string                      `json:"invite_dt,omitempty"`
	JoinedDt     string                      `json:"joined_dt,omitempty"`
	LeftDt       string                      `json:"left_dt,omitempty"`
	RemovedDt    string                      `json:"removed_dt,omitempty"`
	Note         *string                     `json:"note,omitempty"`
	CreateDt     string                      `json:"create_dt"`
	ModifyDt     string                      `json:"modify_dt"`
}

// InvitationTarget is the per-recipient payload accepted by Send.
// IdentifierType is one of EMAIL / PHONE / PROFILE_ID; Identifier is
// the corresponding string. ProposedRole defaults to STUDENT in the
// validator when blank.
type InvitationTarget struct {
	IdentifierType string `json:"identifier_type"`
	Identifier     string `json:"identifier"`
	ProposedRole   string `json:"proposed_role,omitempty"`
}

// SkippedInvitation mirrors a per-target outcome from Send that did
// not land an invitation row but did not abort the batch.
type SkippedInvitation struct {
	IdentifierType string `json:"identifier_type"`
	Identifier     string `json:"identifier"`
	Reason         int64  `json:"reason"`
	Message        string `json:"message"`
}

// SendInvitationReq invites one or more targets to a classroom. The
// caller (resolved from session) must be a manager (OWNER or
// CO_TEACHER) of the classroom — enforced at the service layer. The
// optional Note is copied into ma_classroom_members.note for each
// inserted/reactivated row.
type SendInvitationReq struct {
	ProfileID   int64 `json:"inviter_profile_id"`
	ClassroomID int64 `json:"classroom_id"`
	// Targets     []InvitationTarget `json:"targets"`
	Targets []int64 `json:"targets"`
	Note    *string `json:"note,omitempty"`
}

type SendInvitationRes struct {
	Invitations []*InvitationResponse `json:"invitations"`
	Skipped     []SkippedInvitation   `json:"skipped"`
}

type ListMyPendingInvitationsReq struct {
	ProfileID int64 `json:"profile_id"`
	Page      int64 `json:"page"`
	Size      int64 `json:"size"`
}

type ListMyPendingInvitationsRes struct {
	Invitations []*InvitationResponse  `json:"invitations"`
	Pagination  *pagination.Pagination `json:"pagination"`
}

type ListClassroomInvitationsReq struct {
	ProfileID   int64 `json:"profile_id"`
	ClassroomID int64 `json:"classroom_id"`
	Page        int64 `json:"page"`
	Size        int64 `json:"size"`
}

type ListClassroomInvitationsRes struct {
	Invitations []*InvitationResponse  `json:"invitations"`
	Pagination  *pagination.Pagination `json:"pagination"`
}

type AcceptInvitationReq struct {
	InviteeProfileID int64 `json:"invitee_profile_id"`
	InviterProfileID int64 `json:"inviter_profile_id"`
	ClassroomID      int64 `json:"classroom_id"`
}

type AcceptInvitationRes struct {
	Invitation *InvitationResponse `json:"invitation"`
}

type RejectInvitationReq struct {
	InviteeProfileID int64 `json:"invitee_profile_id"`
	InviterProfileID int64 `json:"inviter_profile_id"`
	ClassroomID      int64 `json:"classroom_id"`
}

type RejectInvitationRes struct{}

// CancelInvitationReq lets a classroom manager revoke a PENDING
// invitation by (classroom_id, target_profile_id).
type CancelInvitationReq struct {
	ProfileID       int64 `json:"profile_id"`
	ClassroomID     int64 `json:"classroom_id"`
	TargetProfileID int64 `json:"target_profile_id"`
}

type CancelInvitationRes struct{}

func InvitationDomainToResponse(m *domain.Member, inviter *profileDTO.ProfileResponse) *InvitationResponse {
	if m == nil {
		return nil
	}
	resp := &InvitationResponse{
		InvitationID: m.MemberId(),
		ClassroomID:  m.ClassroomId(),
		ProfileID:    m.ProfileId(),
		MemberRole:   m.MemberRole(),
		MemberStatus: m.MemberStatus(),
		InviteBy:     m.InviteBy(),
		Note:         m.Note(),
		CreateDt:     m.CreateDt().String(),
		ModifyDt:     m.ModifyDt().String(),
	}
	if m.InviteDt().IsValid() {
		resp.InviteDt = m.InviteDt().String()
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

	if inviter != nil {
		resp.Inviter = inviter
	}
	return resp
}

func InvitationDomainListToResponse(rows []*domain.Member, hashInviters map[int64]*profileDTO.ProfileResponse) []*InvitationResponse {
	out := make([]*InvitationResponse, len(rows))
	for i, m := range rows {
		if m.InviteBy() != nil {
			inviter, ok := hashInviters[*m.InviteBy()]
			if ok {
				out[i] = InvitationDomainToResponse(m, inviter)
			} else {
				out[i] = InvitationDomainToResponse(m, nil)
			}
		} else {
			out[i] = InvitationDomainToResponse(m, nil)
		}
	}
	return out
}
