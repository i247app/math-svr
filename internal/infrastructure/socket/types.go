// Package socket is the in-process WebSocket runtime: a Hub that owns the
// connection registry and topic fan-out, and a Conn that pumps frames over a
// single coder/websocket connection. It is transport-only — authorization and
// business meaning of topics live in the module layer (module/socket).
//
// The Hub is held on the app Resource (like JobRuntime) and consumed by
// producers through the application/socket.Publisher port, so a Redis-backed
// fan-out can replace the in-memory one later without touching callers.
package socket

// Envelope message types (the "type" field of every frame).
const (
	// Client -> server control frames.
	TypeSubscribe   = "subscribe"
	TypeUnsubscribe = "unsubscribe"
	TypePing        = "ping"

	// Server -> client frames.
	TypeEvent = "event"
	TypeAck   = "ack"
	TypeError = "error"
	TypePong  = "pong"
)

// Inbound is a control message decoded from a client text frame.
type Inbound struct {
	Type  string `json:"type"`
	Topic string `json:"topic,omitempty"`
}

// Outbound is a message serialized to a client text frame. MStatus mirrors the
// REST envelope so the mobile client can reuse its status handling for errors.
type Outbound struct {
	Type    string `json:"type"`
	Topic   string `json:"topic,omitempty"`
	Event   string `json:"event,omitempty"`
	Data    any    `json:"data,omitempty"`
	MStatus int    `json:"mstatus,omitempty"`
}
