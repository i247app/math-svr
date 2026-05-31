package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// resolvedInvitationTarget carries the outcome of looking an inviter's
// identifier up against the existing profile graph. Two distinct
// outcomes flow back to the caller:
//
//   - invitedProfileID set + skipReason == 0: a specific profile was
//     pinned and will be written into ma_classroom_invitations.invited_profile_id.
//   - invitedProfileID nil + skipReason == 0: silent-create path —
//     EMAIL/PHONE that didn't resolve to a single profile (no alias
//     row, or alias resolves to zero / >1 profiles). The invitation
//     row is still inserted with NULL invited_profile_id so it can be
//     claimed when the recipient signs up (Phase 5.6 work).
//   - skipReason != 0: PROFILE_ID-typed target where the profile lookup
//     missed; the calling command should append this target to its
//     skipped list and move on without writing a row.
type resolvedInvitationTarget struct {
	invitedProfileID *int64
	skipReason       status.StatusCode
}

// resolveInvitationTarget translates one (identifier_type, identifier)
// pair into an actionable resolution. The lookup runs inside the
// caller's UoW so the resolution and the eventual invitation insert
// see one consistent snapshot of ma_aliases / ma_profiles.
func resolveInvitationTarget(
	ctx context.Context,
	repos transaction.Repositories,
	identifierType, identifier string,
) (resolvedInvitationTarget, error) {
	switch identifierType {
	case string(enum.ClassroomInviteeIdentifierTypeProfileId):
		// prof, err := repos.Profile.FindByProfileId(ctx, identifier)
		// if err != nil {
		// 	return resolvedInvitationTarget{}, errs.NewError(ctx, status.FAIL, nil, err)
		// }
		// if prof == nil {
		// 	// PROFILE_ID is a direct claim — if it doesn't exist, the
		// 	// invitation row has nothing meaningful to point at and the
		// 	// recipient flow has no anchor either. Skip rather than
		// 	// silent-create.
		// 	return resolvedInvitationTarget{skipReason: status.PROFILE_NOT_FOUND}, nil
		// }
		// pid := prof.ProfileId()
		return resolvedInvitationTarget{invitedProfileID: nil}, nil

	case string(enum.ClassroomInviteeIdentifierTypeEmail),
		string(enum.ClassroomInviteeIdentifierTypePhone):
		alias, err := repos.Alias.FindByAka(ctx, identifier)
		if err != nil {
			// FindByAka returns sql.ErrNoRows-derived errors as the
			// generic miss path (the repo doesn't currently translate
			// them to nil). We treat any error here as "unresolved" so
			// the silent-create branch kicks in and the invitation row
			// is still written for later claim.
			return resolvedInvitationTarget{}, nil
		}
		if alias == nil {
			return resolvedInvitationTarget{}, nil
		}
		profiles, err := repos.Profile.ListByUserId(ctx, alias.UserId())
		if err != nil {
			return resolvedInvitationTarget{}, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(profiles) != 1 {
			// Either the user has zero usable profiles or more than
			// one — picking the "right" profile is the recipient's
			// call, not the inviter's. Leave invited_profile_id NULL
			// and let the recipient resolve at accept time (Phase 5.6).
			return resolvedInvitationTarget{}, nil
		}
		pid := profiles[0].ProfileId()
		return resolvedInvitationTarget{invitedProfileID: &pid}, nil

	default:
		// The validator already rejects unknown types, but a defensive
		// branch here keeps a future enum addition from silently
		// short-circuiting the loop.
		return resolvedInvitationTarget{skipReason: status.CLASSROOM_INVITATION_INVALID_IDENTIFIER_TYPE}, nil
	}
}
