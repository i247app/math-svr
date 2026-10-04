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
