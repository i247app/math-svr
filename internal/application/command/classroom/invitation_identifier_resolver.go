package command

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// resolvedInvitationTarget is the per-target outcome of the resolver:
// on success invitedProfileID is non-nil; otherwise skipReason carries
// the per-target reason and the caller appends it to the bulk skip-list
// without aborting the batch.
type resolvedInvitationTarget struct {
	invitedProfileID *int64
	skipReason       status.StatusCode
}

// resolveInvitationTarget maps an (identifierType, identifier) pair to
// a single profile_id. Unlike the legacy ma_classroom_invitations flow,
// ma_classroom_members requires a non-null profile_id, so EMAIL/PHONE
// targets that don't resolve to exactly one profile are skipped with
// PROFILE_NOT_FOUND rather than silently inserted. The caller is
// expected to run this inside a UoW so the alias / profile lookups
// share the surrounding transaction.
func resolveInvitationTarget(ctx context.Context, repos transaction.Repositories, identifierType, identifier string) (resolvedInvitationTarget, error) {
	out := resolvedInvitationTarget{}
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return out, errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_IDENTIFIER, nil,
			errors.New("identifier is required"))
	}

	switch identifierType {
	case string(enum.ClassroomInviteeIdentifierTypeProfileId):
		// PROFILE_ID arrives as a string for transport but is always an
		// int64 in storage. Reject malformed input so the caller gets a
		// precise INVALID_IDENTIFIER instead of a silent miss.
		profileID := utils.StringToInt64(trimmed, 0)
		if profileID == 0 {
			return out, errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_IDENTIFIER, nil,
				errors.New("profile_id must be a positive integer"))
		}
		p, err := repos.Profile.FindByProfileId(ctx, profileID)
		if err != nil {
			return out, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if p == nil {
			out.skipReason = status.PROFILE_NOT_FOUND
			return out, nil
		}
		id := p.ProfileId()
		out.invitedProfileID = &id
		return out, nil

	case string(enum.ClassroomInviteeIdentifierTypeEmail),
		string(enum.ClassroomInviteeIdentifierTypePhone):
		// Look up the alias row; missing alias means the identifier
		// belongs to nobody on the platform yet → per-target skip. A
		// matching alias points at a user; the user may carry several
		// profiles, in which case we can't decide which to invite and
		// also skip. Only the exactly-one-profile case resolves.
		alias, err := repos.Alias.FindByAka(ctx, trimmed)
		if err != nil {
			return out, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if alias == nil {
			out.skipReason = status.PROFILE_NOT_FOUND
			return out, nil
		}
		profiles, err := repos.Profile.ListByUserId(ctx, alias.UserId())
		if err != nil {
			return out, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(profiles) != 1 {
			out.skipReason = status.PROFILE_NOT_FOUND
			return out, nil
		}
		id := profiles[0].ProfileId()
		out.invitedProfileID = &id
		return out, nil

	default:
		return out, errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_IDENTIFIER_TYPE, nil,
			errors.New("identifier_type must be EMAIL, PHONE, or PROFILE_ID"))
	}
}
