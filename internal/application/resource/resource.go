package resource

import (
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/adapter/bot"
	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/adapter/notification"
	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	"math-ai.com/math-ai/internal/adapter/sms"
	"math-ai.com/math-ai/internal/adapter/storage"
	appsocket "math-ai.com/math-ai/internal/application/socket"
	"math-ai.com/math-ai/internal/infrastructure/config"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/httproute"
	jobruntime "math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/metrics"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/infrastructure/socket"

	"github.com/i247app/gex"
	"github.com/i247app/gex/sessionprovider"
)

// Resource carries the long-lived I/O dependencies wired at boot.
// Adapter pointers may be nil when the corresponding adapter is
// disabled in this deploy (e.g. SMS in local dev without Twilio
// credentials); callers must nil-guard.
type Resource struct {
	Env                  *config.Env
	HostConfig           gex.HostConfig
	DB                   *database.DatabaseWithLogs
	SessionManager       *session.SessionManager
	SessionProvider      sessionprovider.SessionProvider
	EmailProvider        *email.Adapter
	SMSProvider          *sms.Adapter
	StorageProvider      *storage.Adapter
	BotProvider          *bot.Adapter
	NotificationProvider *notification.Adapter
	OtpDelivery          *otp_delivery.Adapter

	// JobRegistry holds the static name → handler map for both
	// CronJobs and TaskHandlers. Populated in bootstrap before the
	// Runtime starts; read-mostly after that.
	JobRegistry *jobruntime.Registry

	// JobRuntime owns the scheduler goroutines + task worker pool.
	// Nil only during bootstrap; once Start has returned, all access
	// is safe. Stop is called from the gex OnShutdown hook.
	JobRuntime *jobruntime.Runtime

	// Metrics holds the Prometheus registry + HTTP collectors. Nil when
	// OBS_METRICS_ENABLED=false — every Metrics method is a no-op on nil,
	// and the /metrics route is not registered.
	Metrics *metrics.Metrics

	// RouteClassifier maps a request to its registered route template so the
	// metrics `route` label and span names stay low-cardinality. Created at
	// boot; populated by routes.SetupHttpRoutes as each route is registered.
	RouteClassifier *httproute.Classifier

	// SocketHub owns realtime WebSocket connections + topic fan-out. Nil when
	// SOCKET_ENABLED=false — the /ws/connect route is then not registered.
	SocketHub *socket.Hub

	// SocketPublisher is the producer-facing port over SocketHub. Nil when the
	// socket runtime is disabled — producers must nil-guard.
	SocketPublisher appsocket.Publisher
}

func (a *Resource) GetRequestSession(r *http.Request) (*session.AppSession, error) {
	sess := session.GetRequestSession(r)
	if sess == nil {
		return nil, fmt.Errorf("session not found")
	}

	return sess, nil
}

func (a *Resource) GetRequestUID(r *http.Request) (int64, error) {
	sess, err := a.GetRequestSession(r)
	if err != nil {
		return 0, err
	}

	// ## We can't do this because sometimes you want the uid even if not authenticated
	// if isSecure, ok := sess.Get("is_secure"); !ok || !isSecure.(bool) {
	// 	return 0, fmt.Errorf("session is not secure, auth required")
	// }

	id, ok := sess.UID()
	if !ok {
		return 0, fmt.Errorf("uid missing from session (did you forget to send the Authorization header?)")
	}

	return id, nil
}
