package session

// DeleteSessionReq names one session by its token — the string the client
// sends as metadata.authorization ("Bearer " optional). The body key is
// `token` so LogRequestMiddleware redacts it: it is a live credential.
type DeleteSessionReq struct {
	Token string `json:"token"`
}

// DeleteSessionRes says whose session was deleted. UID is null for a
// session that never carried one (signed out, or never signed in).
type DeleteSessionRes struct {
	UID      *int64 `json:"uid"`
	IsSecure bool   `json:"is_secure"`
}

// MarkSessionSecureReq sets is_secure on the session stored under Token
// (same token rules as DeleteSessionReq). IsSecure is required: a missing
// value must not silently read as false and sign the user out.
type MarkSessionSecureReq struct {
	Token    string `json:"token"`
	IsSecure *bool  `json:"is_secure"`
}

// MarkSessionSecureRes is the session after the change. UID is null for a
// session no user is signed into (only is_secure: false is accepted there).
type MarkSessionSecureRes struct {
	UID      *int64 `json:"uid"`
	IsSecure bool   `json:"is_secure"`
}
