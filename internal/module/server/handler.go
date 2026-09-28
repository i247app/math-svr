package server

import (
	"errors"
	"net/http"
	"time"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
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

// HandleShutdown acknowledges the request, then triggers a self-shutdown so
// the existing gex signal handler (SIGINT/SIGTERM) runs every OnShutdown hook
// — JobRuntime drain + session serialization — and then gracefully stops the
// HTTP server. We do not call os.Exit so the deferred cleanup in cmd/mathsvr
// (app.Close → DB + log file) still runs.
//
// The actual self-shutdown mechanism is OS-specific: on Unix it raises
// SIGTERM against our own PID (selfShutdown in shutdown_unix.go), while on
// Windows — which has no POSIX signals — it falls back to a plain exit
// (shutdown_windows.go). Windows is build/dev-only here.
func (h *Handler) HandleShutdown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.From(ctx)

	if !stopping.CompareAndSwap(false, true) {
		response.WriteJson(w, nil, errs.NewError(ctx, status.SERVER_SHUTTING_DOWN, nil,
			errors.New("a shutdown or reload is already in progress")))
		return
	}

	log.Warnf("server.shutdown.requested uid=%d remote=%s", requesterUID(r), r.RemoteAddr)

	response.WriteJsonNoContent(w, nil)

	time.AfterFunc(shutdownDelay, func() {
		if err := selfShutdown(); err != nil {
			logger.From(ctx).Errorf("server.shutdown.signal_failed err=%v", err)
		}
	})
}

// requesterUID is the caller's uid for the audit log line, -1 when unknown.
func requesterUID(r *http.Request) int64 {
	if sess := session.GetRequestSession(r); sess != nil {
		if id, ok := sess.UID(); ok {
			return id
		}
	}
	return -1
}
