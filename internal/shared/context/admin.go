package context

import "context"

// AdminVia says how a request passed the admin gate: an ADMIN account's
// session, or the X-Api-Admin-Key header. Handlers read it to attribute
// the action — a key carries no account, so there is no uid to log.
type AdminVia string

const (
	AdminViaSession AdminVia = "session"
	AdminViaAPIKey  AdminVia = "api_key"
)

const ctxKeyAdminVia contextKey = "math-ai.admin_via"

func WithAdminVia(ctx context.Context, via AdminVia) context.Context {
	return context.WithValue(ctx, ctxKeyAdminVia, via)
}

// GetAdminVia returns "" when the request did not pass an admin gate.
func GetAdminVia(ctx context.Context) AdminVia {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(ctxKeyAdminVia).(AdminVia)
	return v
}
