package socket

import (
	"encoding/json"
	"testing"

	"github.com/coder/websocket"
)

// newTestConn registers a connection with a nil websocket. Enqueue/close logic
// is transport-independent, so the Hub's routing can be exercised without a
// real socket by draining the send channel directly.
func newTestConn(t *testing.T, h *Hub, userID int64) *Conn {
	t.Helper()
	c, err := h.NewConn(nil, userID)
	if err != nil {
		t.Fatalf("NewConn(uid=%d): %v", userID, err)
	}
	return c
}

func drain(t *testing.T, c *Conn) (Outbound, bool) {
	t.Helper()
	select {
	case frame := <-c.send:
		var o Outbound
		if err := json.Unmarshal(frame, &o); err != nil {
			t.Fatalf("unmarshal frame: %v", err)
		}
		return o, true
	default:
		return Outbound{}, false
	}
}

func isClosed(c *Conn) bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func TestHubPublishFanOut(t *testing.T) {
	h := NewHub(Config{})
	a := newTestConn(t, h, 1)
	b := newTestConn(t, h, 2)
	c := newTestConn(t, h, 3) // not subscribed

	h.Subscribe(a, "room:1")
	h.Subscribe(b, "room:1")

	if n := h.Publish("room:1", "msg", map[string]int{"x": 1}); n != 2 {
		t.Fatalf("delivered = %d, want 2", n)
	}

	for _, sub := range []*Conn{a, b} {
		got, ok := drain(t, sub)
		if !ok {
			t.Fatalf("conn %s got no frame", sub.id)
		}
		if got.Type != TypeEvent || got.Topic != "room:1" || got.Event != "msg" {
			t.Fatalf("unexpected frame: %+v", got)
		}
	}
	if _, ok := drain(t, c); ok {
		t.Fatal("non-subscriber received a frame")
	}
}

func TestHubUnsubscribe(t *testing.T) {
	h := NewHub(Config{})
	a := newTestConn(t, h, 1)
	h.Subscribe(a, "room:1")
	h.Unsubscribe(a, "room:1")

	if n := h.Publish("room:1", "msg", nil); n != 0 {
		t.Fatalf("delivered = %d, want 0 after unsubscribe", n)
	}
	if st := h.Stats(); st.Topics != 0 {
		t.Fatalf("topics = %d, want 0 (empty topic pruned)", st.Topics)
	}
}

func TestHubBroadcastUser(t *testing.T) {
	h := NewHub(Config{})
	a1 := newTestConn(t, h, 7)
	a2 := newTestConn(t, h, 7)
	other := newTestConn(t, h, 8)

	if n := h.BroadcastUser(7, "ping", nil); n != 2 {
		t.Fatalf("delivered = %d, want 2", n)
	}
	if _, ok := drain(t, a1); !ok {
		t.Fatal("a1 got no frame")
	}
	if _, ok := drain(t, a2); !ok {
		t.Fatal("a2 got no frame")
	}
	if _, ok := drain(t, other); ok {
		t.Fatal("other user received a frame")
	}
}

func TestHubBackpressureClosesSlowConsumer(t *testing.T) {
	h := NewHub(Config{BufferSize: 1})
	a := newTestConn(t, h, 1)
	h.Subscribe(a, "t")

	// First publish fills the buffer (depth 1); second overflows.
	h.Publish("t", "e1", nil)
	if n := h.Publish("t", "e2", nil); n != 0 {
		t.Fatalf("delivered = %d, want 0 (buffer full)", n)
	}
	if !isClosed(a) {
		t.Fatal("slow consumer was not closed on backpressure")
	}
}

func TestHubMaxConnsPerUser(t *testing.T) {
	h := NewHub(Config{MaxConnsPerUser: 2})
	_ = newTestConn(t, h, 1)
	_ = newTestConn(t, h, 1)
	if _, err := h.NewConn(nil, 1); err != ErrTooManyConnections {
		t.Fatalf("err = %v, want ErrTooManyConnections", err)
	}
	// A different user is unaffected.
	if _, err := h.NewConn(nil, 2); err != nil {
		t.Fatalf("NewConn(uid=2): %v", err)
	}
}

func TestHubUnregisterCleansTopics(t *testing.T) {
	h := NewHub(Config{})
	a := newTestConn(t, h, 1)
	h.Subscribe(a, "room:1")
	h.unregister(a)

	st := h.Stats()
	if st.Connections != 0 || st.Users != 0 || st.Topics != 0 {
		t.Fatalf("stats after unregister = %+v, want all zero", st)
	}
	// Republishing must not panic and delivers to nobody.
	if n := h.Publish("room:1", "e", nil); n != 0 {
		t.Fatalf("delivered = %d, want 0", n)
	}
}

func TestHubCloseAll(t *testing.T) {
	h := NewHub(Config{})
	a := newTestConn(t, h, 1)
	b := newTestConn(t, h, 2)

	h.CloseAll(websocket.StatusGoingAway, "shutdown")

	if !isClosed(a) || !isClosed(b) {
		t.Fatal("CloseAll did not signal every connection")
	}
}
