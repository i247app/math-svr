package enum

// IdentityCodeType is who a row is to the product, carried by
// ma_users.identity_code and ma_profiles.identity_code.
//
// It is deliberately orthogonal to the two status columns beside it:
// user_status says whether the row is still in force, profile_status
// whether the profile carries enough information. Identity says whether
// there is a registered person behind it at all.
//
// VERIFIED is locked to ProfileStatusTypeOfficial — both come out of one
// function so the two columns can never drift. Nothing else writes it.
type IdentityCodeType string

const (
	// IdentityCodeGuest is a visitor who has never registered. They hold a
	// real user + profile row (so exams, journeys and analytics need no
	// special case) with role and phone still NULL.
	IdentityCodeGuest IdentityCodeType = "GUEST"
	// IdentityCodeUser has registered: a phone, a role, an alias to log in
	// with.
	IdentityCodeUser IdentityCodeType = "USER"
	// IdentityCodeVerified has registered AND proven the role — a teacher
	// with an issuing body and a teacher id, a student with a student id.
	IdentityCodeVerified IdentityCodeType = "VERIFIED"
)

func (c IdentityCodeType) String() string {
	return string(c)
}

func (c IdentityCodeType) IsValid() bool {
	switch c {
	case IdentityCodeGuest, IdentityCodeUser, IdentityCodeVerified:
		return true
	default:
		return false
	}
}

// IsGuest is the check every guest-facing rule reads, so no caller has to
// compare against the literal.
func (c IdentityCodeType) IsGuest() bool {
	return c == IdentityCodeGuest
}

func ListIdentityCodes() []string {
	return []string{
		IdentityCodeGuest.String(),
		IdentityCodeUser.String(),
		IdentityCodeVerified.String(),
	}
}
