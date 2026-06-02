package command

import "errors"

// Module-scoped sentinel errors for the classroom command package.
//
// These strings travel inside errs.NewError(...) as the base error and
// are surfaced to logs / the response envelope's `debug` field only —
// the user-facing message comes from the status code via
// MathError.GetStatusMessage(). Keeping them as package-level vars
// makes the per-command call sites read at a glance and avoids the
// drift that hand-typed strings invite.
var (
	// Classroom lifecycle / lookup.
	ErrClassroomNotFound            = errors.New("classroom not found")
	ErrClassroomNotFoundAfterInsert = errors.New("classroom not found after insert")
	ErrClassroomNotFoundAfterUpdate = errors.New("classroom not found after update")
	ErrClassroomArchived            = errors.New("classroom is archived")
	ErrClassroomAlreadyArchived     = errors.New("classroom is already archived")
	ErrClassroomNotArchived         = errors.New("classroom is not archived")
	ErrClassroomFull                = errors.New("classroom is full")

	// Classroom code (a.k.a. legacy "invite code") flows.
	ErrClassroomCodeTaken         = errors.New("invite code already taken")
	ErrClassroomCodeNotFound      = errors.New("invite code not found")
	ErrClassroomCodeExpired       = errors.New("invite code expired")
	ErrClassroomCodeMintExhausted = errors.New("could not mint a unique classroom code")

	// Membership / role.
	ErrNotClassroomMember           = errors.New("not a member of this classroom")
	ErrMembershipNotActive          = errors.New("membership is not active")
	ErrOwnerMustTransferBeforeLeave = errors.New("owner must transfer ownership before leaving")
	ErrOwnerCannotBeRemoved         = errors.New("owner cannot be removed")
	ErrTargetMemberNotFound         = errors.New("target member not found")
	ErrTargetMemberNotActive        = errors.New("target member is not active")
	ErrOwnerRoleUpdateForbidden     = errors.New("owner role cannot be changed via update; use transfer")
	ErrAlreadyActiveMember          = errors.New("already an active member")
	ErrPendingInvitationExists      = errors.New("a pending invitation already exists; accept it instead")
	ErrPendingJoinRequestExists     = errors.New("a pending join request already exists")

	// Invitations.
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationNotPending = errors.New("invitation is not pending")

	// Join requests.
	ErrJoinRequestNotFound   = errors.New("join request not found")
	ErrJoinRequestNotPending = errors.New("join request is not pending")

	// Transfer ownership.
	ErrTransferSameOwner       = errors.New("new owner must differ from current owner")
	ErrCallerNotCurrentOwner   = errors.New("caller is not the current owner")
	ErrNewOwnerNotMember       = errors.New("new owner must be an existing member")
	ErrNewOwnerNotActiveMember = errors.New("new owner must be an active member")

	// Invitation identifier resolver.
	ErrInvitationIdentifierRequired = errors.New("identifier is required")
	ErrProfileIDMustBePositive      = errors.New("profile_id must be a positive integer")
	ErrInvalidIdentifierType        = errors.New("identifier_type must be EMAIL, PHONE, or PROFILE_ID")
)
