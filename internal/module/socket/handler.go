package socket

import (
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

// Handler exposes the WebSocket upgrade endpoint.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// HandleConnect upgrades an authenticated request to a WebSocket connection.
//
// Auth is enforced HERE rather than via AuthRequiredMiddleware: a WebSocket
// client keys off the handshake's HTTP status, so an auth failure must return a
// real 401 BEFORE the upgrade — not the app-wide "HTTP 200 + mstatus in body"
// envelope, which a WS client reports only as an opaque "unexpected response".
// The session itself is still resolved upstream by the WS-aware
// GexSessionMiddleware; we just require it to be secure and carry a uid.
//
// Once websocket.Accept succeeds the connection owns the response, so any
// post-accept problem closes the socket with a status code instead.
func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess := session.GetRequestSession(r)
	if sess == nil {
		writeHandshakeUnauthorized(w, "session not found")
		return
	}
	if !sess.IsSecure() {
		writeHandshakeUnauthorized(w, "session is not secure, login required")
		return
	}
	uid, ok := sess.UID()
	if !ok {
		writeHandshakeUnauthorized(w, "uid missing from session")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.svc.OriginPatterns(),
	})
	if err != nil {
		// Accept has already written the HTTP error response (e.g. 403 on a
		// disallowed Origin). Nothing more to do.
		return
	}

	// Blocks until the peer, a pump error, or Hub shutdown closes the socket.
	h.svc.Connect(ctx, conn, uid)
}

// writeHandshakeUnauthorized rejects a WebSocket handshake with a real HTTP 401
// so the client can distinguish an auth failure from a transport error. The
// JSON body mirrors the app's error envelope for any HTTP-level inspector.
func writeHandshakeUnauthorized(w http.ResponseWriter, debug string) {
	body, _ := json.Marshal(struct {
		MStatus  int    `json:"mstatus"`
		MMessage string `json:"mmessage"`
		Debug    string `json:"debug,omitempty"`
	}{
		MStatus:  int(status.SOCKET_UNAUTHORIZED),
		MMessage: string(status.GetVNMessage(status.SOCKET_UNAUTHORIZED)),
		Debug:    debug,
	})
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(body)
}
