package eino

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// connTrace captures connection-level facts about a single outbound LLM
// request so we can tell, per call, whether the connection to the vendor was
// brand-new (TLS handshake) or reused from the keep-alive pool.
//
// Caveat: httptrace only fires if the vendor SDK threads the request context
// through net/http. If a call shows reused=false AND tls_handshake_ms=0, the
// trace did not capture (e.g. the SDK ignored the ctx or used gRPC) — fall
// back to ss/tcpdump at the OS level.
type connTrace struct {
	gotConn   bool
	reused    bool
	wasIdle   bool
	idleMs    int64
	tlsStart  time.Time
	tlsDoneMs int64
	remote    string
}

// withConnTrace attaches an httptrace.ClientTrace to ctx and returns the
// trace sink to read after the request completes.
func withConnTrace(ctx context.Context) (context.Context, *connTrace) {
	ct := &connTrace{}
	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			ct.gotConn = true
			ct.reused = info.Reused
			ct.wasIdle = info.WasIdle
			ct.idleMs = info.IdleTime.Milliseconds()
			if info.Conn != nil {
				ct.remote = info.Conn.RemoteAddr().String()
			}
		},
		TLSHandshakeStart: func() { ct.tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !ct.tlsStart.IsZero() {
				ct.tlsDoneMs = time.Since(ct.tlsStart).Milliseconds()
			}
		},
	}
	return httptrace.WithClientTrace(ctx, trace), ct
}
