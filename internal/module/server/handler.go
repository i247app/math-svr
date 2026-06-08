package server

import (
	"net/http"
	"os"
	"syscall"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
)

// shutdownDelay leaves enough time for net/http to flush the response
// body and for the client TCP socket to drain before the SIGTERM-driven
// shutdown path closes the listener.
const shutdownDelay = 250 * time.Millisecond

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// HandleShutdown acknowledges the request, then schedules SIGTERM against
// our own PID so the existing gex signal handler (SIGINT/SIGTERM) runs
// every OnShutdown hook — JobRuntime drain + session serialization — and
// then gracefully stops the HTTP server. We do not call os.Exit so the
// deferred cleanup in cmd/mathsvr (app.Close → DB + log file) still runs.
func (h *Handler) HandleShutdown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	uid := int64(-1)
	if sess := session.GetRequestSession(r); sess != nil {
		if id, ok := sess.UID(); ok {
			uid = id
		}
	}
	log.Warnf("server.shutdown.requested uid=%d remote=%s", uid, r.RemoteAddr)

	response.WriteJsonNoContent(w)

	time.AfterFunc(shutdownDelay, func() {
		if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
			logger.From(ctx).Errorf("server.shutdown.signal_failed err=%v", err)
		}
	})
}
