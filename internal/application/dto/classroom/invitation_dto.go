package classroom

import (
	domain "math-ai.com/math-ai/internal/domain/classroom"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// InvitationResponse is the wire shape for ma_classroom_invitations
// rows. Optional times follow the same "empty string when zero" pattern
// as MemberResponse so the client gets a stable JSON shape regardless of
// the row's lifecycle state. Token is intentionally excluded from the
// default response — it's a secret meant for the share-by-link path
// (deferred) and should not leak through the manager/recipient roster
// endpoints.
type InvitationResponse struct {
	ID                    int64   `json:"id"`
	InvitationID          int64   `json:"invitation_id"`
	ClassroomID           int64   `json:"classroom_id"`
	InviterProfileID      int64   `json:"inviter_profile_id"`
	InvitedProfileID      *int64  `json:"invited_profile_id,omitempty"`
	InviteeIdentifier     *string `json:"invitee_identifier,omitempty"`
	InviteeIdentifierType *string `json:"invitee_identifier_type,omitempty"`
	ProposedRole          string  `json:"proposed_role"`
	Message               *string `json:"message,omitempty"`
	SentDt                string  `json:"sent_dt,omitempty"`
	ExpiresDt             string  `json:"expires_dt,omitempty"`
	RespondedDt           string  `json:"responded_dt,omitempty"`
	ResponseProfileID     *int64  `json:"response_profile_id,omitempty"`
	CancelledByProfileID  *int64  `json:"cancelled_by_profile_id,omitempty"`
	Note                  *string `json:"note,omitempty"`
	InvitationStatus      *string `json:"invitation_status,omitempty"`
	CreateDt              string  `json:"create_dt"`
	ModifyDt              string  `json:"modify_dt"`
}

// InvitationTarget describes one recipient in a bulk-create request.
// IdentifierType is one of EMAIL / PHONE / PROFILE_ID — the command
// resolves EMAIL/PHONE via ma_aliases at create time and falls through
// to silent-create-with-null-invited-profile when no profile matches,
// per the Phase 5 design.
//
// ProposedRole is optional; nil/empty defaults to STUDENT. Only OWNER
// callers may supply CO_TEACHER (enforced at the service layer).
type InvitationTarget struct {
	IdentifierType string  `json:"identifier_type"`
	Identifier     int64   `json:"identifier"`
	ProposedRole   *string `json:"proposed_role,omitempty"`
}

// CreateInvitationReq is a bulk-create payload. The handler walks each
// target inside one UoW and reports per-target outcomes through
// CreateInvitationRes.Skipped instead of aborting the batch on the first
// "already a member" or "already invited" — gives the UI partial
// success without forcing the caller to retry one-by-one.
type CreateInvitationReq struct {
	ProfileID   int64              `json:"profile_id"`
	ClassroomID int64              `json:"classroom_id"`
	Targets     []InvitationTarget `json:"targets"`
	Message     *string            `json:"message,omitempty"`
	// ExpiresDt is optional; when omitted the command defaults to
	// now + 7 days. Must be in the future when supplied.
	ExpiresDt mtime.MathTime `json:"expires_dt,omitempty"`
}

// SkippedInvitation reports why a single target did not produce an
// invitation row (already a member, already pending invitation, profile
// lookup failed, etc.). Reason is the MathError status code as an
// integer so the client can map it to a localized message without
// re-parsing free-form strings.
type SkippedInvitation struct {
	IdentifierType string `json:"identifier_type"`
	Identifier     string `json:"identifier"`
	Reason         int64  `json:"reason"`
	Message        string `json:"message,omitempty"`
}

type CreateInvitationRes struct {
	Invitations []*InvitationResponse `json:"invitations"`
	Skipped     []SkippedInvitation   `json:"skipped"`
}

// ListInvitationsByClassroomReq powers the manager's roster view.
// Caller must be OWNER / CO_TEACHER of the classroom (or the original
// inviter); enforcement lives at the service layer.
type ListInvitationsByClassroomReq struct {
	ProfileID   int64   `json:"profile_id"`
	ClassroomID int64   `json:"classroom_id"`
	Status      *string `json:"status,omitempty"`
	Page        int64   `json:"page"`
	Size        int64   `json:"size"`
}

type ListInvitationsByClassroomRes struct {
	Invitations []*InvitationResponse  `json:"invitations"`
	Pagination  *pagination.Pagination `json:"pagination"`
}

// ListMyPendingInvitationsReq returns every invitation targeted at the
// caller's profile. Status defaults to PENDING when omitted so the
// common "do I have anything to act on?" call stays a single field.
type ListMyPendingInvitationsReq struct {
	ProfileID int64   `json:"profile_id"`
	Status    *string `json:"status,omitempty"`
	Page      int64   `json:"page"`
	Size      int64   `json:"size"`
}

type ListMyPendingInvitationsRes struct {
	Invitations []*InvitationResponse  `json:"invitations"`
	Pagination  *pagination.Pagination `json:"pagination"`
}

// AcceptInvitationReq is the invitee's "yes" action. Inside the
// command's UoW: verify PENDING + not expired + classroom still
// joinable + caller is the intended invitee + max_members not reached,
// then flip ACCEPTED and mint/reactivate the matching member row
// atomically.
type AcceptInvitationReq struct {
	ProfileID    int64 `json:"profile_id"`
	InvitationID int64 `json:"invitation_id"`
}

type AcceptInvitationRes struct {
	Member *MemberResponse `json:"member"`
}

// RejectInvitationReq is the invitee's "no" action. No member row is
// created; the invitation flips REJECTED with responded_dt = now.
type RejectInvitationReq struct {
	ProfileID    int64 `json:"profile_id"`
	InvitationID int64 `json:"invitation_id"`
}

type RejectInvitationRes struct{}

// CancelInvitationReq lets a classroom manager (OWNER / CO_TEACHER) or
// the original inviter cancel a PENDING invitation before it's
// accepted. Status flips to CANCELLED and cancelled_by_profile_id
// records the caller.
type CancelInvitationReq struct {
	ProfileID    int64 `json:"profile_id"`
	InvitationID int64 `json:"invitation_id"`
}

type CancelInvitationRes struct{}

// ResendInvitationReq refreshes an existing invitation in place rather
// than minting a new row — keeps idx_invited_profile lean and avoids
// duplicate-pending logic. EXPIRED rows flip back to PENDING with new
// expires_dt; PENDING rows just get a new expires_dt. CANCELLED /
// REJECTED / ACCEPTED rows are terminal and rejected.
type ResendInvitationReq struct {
	ProfileID    int64          `json:"profile_id"`
	InvitationID int64          `json:"invitation_id"`
	ExpiresDt    mtime.MathTime `json:"expires_dt,omitempty"`
}

type ResendInvitationRes struct {
	Invitation *InvitationResponse `json:"invitation"`
}

// InvitationDomainToResponse maps a domain Invitation onto the wire
// shape. Optional time fields render to "" when invalid so the JSON
// envelope stays consistent regardless of which lifecycle state the
// row is in.
func InvitationDomainToResponse(inv *domain.Invitation) *InvitationResponse {
	if inv == nil {
		return nil
	}
	resp := &InvitationResponse{
		ID:                    inv.Id(),
		InvitationID:          inv.InvitationId(),
		ClassroomID:           inv.ClassroomId(),
		InviterProfileID:      inv.InviterProfileId(),
		InvitedProfileID:      inv.InvitedProfileId(),
		InviteeIdentifier:     inv.InviteeIdentifier(),
		InviteeIdentifierType: inv.InviteeIdentifierType(),
		ProposedRole:          inv.ProposedRole(),
		Message:               inv.Message(),
		ResponseProfileID:     inv.ResponseProfileId(),
		CancelledByProfileID:  inv.CancelledByProfileId(),
		Note:                  inv.Note(),
		InvitationStatus:      inv.InvitationStatus(),
		CreateDt:              inv.CreateDt().String(),
		ModifyDt:              inv.ModifyDt().String(),
	}
	if inv.SentDt().IsValid() {
		resp.SentDt = inv.SentDt().String()
	}
	if inv.ExpiresDt().IsValid() {
		resp.ExpiresDt = inv.ExpiresDt().String()
	}
	if inv.RespondedDt().IsValid() {
		resp.RespondedDt = inv.RespondedDt().String()
	}
	return resp
}

func InvitationDomainListToResponse(invs []*domain.Invitation) []*InvitationResponse {
	result := make([]*InvitationResponse, len(invs))
	for i, inv := range invs {
		result[i] = InvitationDomainToResponse(inv)
	}
	return result
}
