package socket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"math-ai.com/math-ai/internal/domain/shared/status"
	socketrt "math-ai.com/math-ai/internal/infrastructure/socket"
)

// newTestServer stands up an httptest server that accepts a WebSocket and
// serves it through a real socket.Service for the given uid. Auth/middleware
// are intentionally out of scope here — this exercises the transport + Hub +
// dispatch end to end.
func newTestServer(t *testing.T, hub *socketrt.Hub, uid int64) *httptest.Server {
	t.Helper()
	svc := NewService(hub, nil, nil) // DefaultAuthorizer, same-origin only
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		svc.Connect(r.Context(), c, uid)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func wsURL(httpURL string) string { return "ws" + strings.TrimPrefix(httpURL, "http") }

// dial opens a client connection with a bounded context.
func dial(t *testing.T, ctx context.Context, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.Dial(ctx, wsURL(srv.URL), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

// pingBarrier sends a ping and waits for the pong. Receiving the pong proves
// the server's Serve loop is running — which means the auto-subscriptions
// (performed before Serve) are already in place, making later publishes
// deterministic.
func pingBarrier(t *testing.T, ctx context.Context, c *websocket.Conn) {
	t.Helper()
	if err := wsjson.Write(ctx, c, socketrt.Inbound{Type: socketrt.TypePing}); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	var out socketrt.Outbound
	if err := wsjson.Read(ctx, c, &out); err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if out.Type != socketrt.TypePong {
		t.Fatalf("barrier: got %q, want pong", out.Type)
	}
}

func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestSocket_PingPong(t *testing.T) {
	hub := socketrt.NewHub(socketrt.Config{})
	srv := newTestServer(t, hub, 1)
	ctx := testContext(t)

	c := dial(t, ctx, srv)
	pingBarrier(t, ctx, c)
}

func TestSocket_SubscribeAllowedThenReceiveEvent(t *testing.T) {
	const uid = int64(7)
	hub := socketrt.NewHub(socketrt.Config{})
	srv := newTestServer(t, hub, uid)
	ctx := testContext(t)

	c := dial(t, ctx, srv)
	pingBarrier(t, ctx, c) // barrier: auto-subscriptions are now in place

	// Publish to the auto-subscribed notifications topic; the client must
	// receive the event.
	hub.Publish("notifications:7", "notification.created", map[string]any{"id": 99})

	var evt socketrt.Outbound
	if err := wsjson.Read(ctx, c, &evt); err != nil {
		t.Fatalf("read event: %v", err)
	}
	if evt.Type != socketrt.TypeEvent || evt.Topic != "notifications:7" || evt.Event != "notification.created" {
		t.Fatalf("unexpected event frame: %+v", evt)
	}
}

func TestSocket_SubscribeForbiddenTopic(t *testing.T) {
	hub := socketrt.NewHub(socketrt.Config{})
	srv := newTestServer(t, hub, 7)
	ctx := testContext(t)

	c := dial(t, ctx, srv)
	// Resource topics are denied by DefaultAuthorizer.
	if err := wsjson.Write(ctx, c, socketrt.Inbound{Type: socketrt.TypeSubscribe, Topic: "classroom:1"}); err != nil {
		t.Fatalf("write subscribe: %v", err)
	}
	var out socketrt.Outbound
	if err := wsjson.Read(ctx, c, &out); err != nil {
		t.Fatalf("read error frame: %v", err)
	}
	if out.Type != socketrt.TypeError || out.MStatus != int(status.SOCKET_TOPIC_FORBIDDEN) {
		t.Fatalf("got %+v, want error frame with SOCKET_TOPIC_FORBIDDEN", out)
	}
}

func TestSocket_SubscribeOwnTopicAcked(t *testing.T) {
	hub := socketrt.NewHub(socketrt.Config{})
	srv := newTestServer(t, hub, 7)
	ctx := testContext(t)

	c := dial(t, ctx, srv)
	if err := wsjson.Write(ctx, c, socketrt.Inbound{Type: socketrt.TypeSubscribe, Topic: "user:7"}); err != nil {
		t.Fatalf("write subscribe: %v", err)
	}
	var out socketrt.Outbound
	if err := wsjson.Read(ctx, c, &out); err != nil {
		t.Fatalf("read ack: %v", err)
	}
	if out.Type != socketrt.TypeAck || out.Topic != "user:7" {
		t.Fatalf("got %+v, want ack for user:7", out)
	}
}

func TestSocket_ShutdownClosesClientGoingAway(t *testing.T) {
	hub := socketrt.NewHub(socketrt.Config{})
	srv := newTestServer(t, hub, 1)
	ctx := testContext(t)

	c := dial(t, ctx, srv)
	pingBarrier(t, ctx, c) // ensure the connection is registered before shutdown

	// A real client is always reading; the close handshake needs the peer to be
	// reading to reply. Read concurrently, then shut the Hub down.
	readErr := make(chan error, 1)
	go func() {
		_, _, err := c.Read(ctx)
		readErr <- err
	}()
	time.Sleep(50 * time.Millisecond) // let the reader block on Read

	hub.Shutdown("server shutting down")

	select {
	case err := <-readErr:
		if websocket.CloseStatus(err) != websocket.StatusGoingAway {
			t.Fatalf("close status = %v, want StatusGoingAway (err=%v)", websocket.CloseStatus(err), err)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for the client to observe the close")
	}
}
