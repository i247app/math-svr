package resource

import (
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/adapter/bot"
	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	"math-ai.com/math-ai/internal/adapter/sms"
	"math-ai.com/math-ai/internal/adapter/storage"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/session"

	"github.com/i247app/gex"
	"github.com/i247app/gex/sessionprovider"
)

// Resource carries the long-lived I/O dependencies wired at boot.
// Adapter pointers may be nil when the corresponding adapter is
// disabled in this deploy (e.g. SMS in local dev without Twilio
// credentials); callers must nil-guard.
type Resource struct {
	Env             *config.Env
	HostConfig      gex.HostConfig
	DB              *database.DatabaseWithLogs
	SessionManager  *session.SessionManager
	SessionProvider sessionprovider.SessionProvider
	EmailProvider   *email.Adapter
	SMSProvider     *sms.Adapter
	StorageProvider *storage.Adapter
	BotProvider     *bot.Adapter
	OtpDelivery     *otp_delivery.Adapter
}

func (a *Resource) GetRequestSession(r *http.Request) (*session.AppSession, error) {
	sess := session.GetRequestSession(r)
	if sess == nil {
		return nil, fmt.Errorf("session not found")
	}

	return sess, nil
}

func (a *Resource) GetRequestUID(r *http.Request) (string, error) {
	sess, err := a.GetRequestSession(r)
	if err != nil {
		return "", err
	}

	// ## We can't do this because sometimes you want the uid even if not authenticated
	// if isSecure, ok := sess.Get("is_secure"); !ok || !isSecure.(bool) {
	// 	return 0, fmt.Errorf("session is not secure, auth required")
	// }

	id, ok := sess.UID()
	if !ok {
		return "", fmt.Errorf("uid missing from session (did you forget to send the Authorization header?)")
	}

	return id, nil
}
