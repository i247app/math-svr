package resource

import (
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
