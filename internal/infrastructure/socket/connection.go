package socket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// ConnConfig tunes a single connection's pumps. Zero values fall back to the
// package defaults below.
type ConnConfig struct {
	BufferSize   int           // outbound queue depth before backpressure kicks in
	PingInterval time.Duration // server->client ping cadence (liveness)
	WriteTimeout time.Duration // per-frame write / ping deadline
	ReadLimit    int64         // max inbound frame size
}

const (
	defaultBufferSize   = 64
	defaultPingInterval = 30 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultReadLimit    = 32 * 1024 // 32 KiB
)

func (c ConnConfig) withDefaults() ConnConfig {
	if c.BufferSize <= 0 {
		c.BufferSize = defaultBufferSize
	}
	if c.PingInterval <= 0 {
		c.PingInterval = defaultPingInterval
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = defaultWriteTimeout
	}
	if c.ReadLimit <= 0 {
		c.ReadLimit = defaultReadLimit
	}
	return c
}

// Conn is a Hub-managed WebSocket connection. Exactly one goroutine reads
// (readPump) and one writes (writePump); all outbound frames flow through the
// buffered send channel so the Hub never blocks on a single slow client.
type Conn struct {
	id     string
	userID int64
	ws     *websocket.Conn
	hub    *Hub
	cfg    ConnConfig

	// onMsg handles decoded control frames other than ping (subscribe /
	// unsubscribe). Set by the module layer; nil in transport-only tests.
	onMsg func(c *Conn, in Inbound)

	send chan []byte

	closeOnce   sync.Once
	closeWSOnce sync.Once
	closed      chan struct{}
	closeCode   websocket.StatusCode
	closeReason string
}

// ID returns the process-unique connection id.
func (c *Conn) ID() string { return c.id }

// UserID returns the authenticated user this connection belongs to.
func (c *Conn) UserID() int64 { return c.userID }

// SetOnMessage installs the control-frame handler. Call before Serve.
func (c *Conn) SetOnMessage(fn func(c *Conn, in Inbound)) { c.onMsg = fn }

// enqueue attempts a non-blocking send of a pre-marshaled frame. It returns
// false when the connection is closing or its buffer is full (backpressure);
// the Hub treats false as a slow-consumer signal and closes the connection.
func (c *Conn) enqueue(frame []byte) bool {
	select {
	case <-c.closed:
		return false
	default:
	}
	select {
	case c.send <- frame:
		return true
	default:
		return false
	}
}

// Send marshals and enqueues a single message. Used for connection-scoped
// frames (ack / pong / error); topic fan-out marshals once in the Hub. It is
// non-blocking and returns false under backpressure or after close.
func (c *Conn) Send(o Outbound) bool {
	frame, err := json.Marshal(o)
	if err != nil {
		return false
	}
	return c.enqueue(frame)
}

// triggerClose records the close code/reason and signals both pumps exactly
// once. It is NON-BLOCKING: it never touches the socket, so it is safe to call
// from the Hub during shutdown or from Publish on the backpressure path without
// stalling the caller. The actual graceful ws.Close is performed by the
// writePump (the single write owner) — see closeWS.
func (c *Conn) triggerClose(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		c.closeCode = code
		c.closeReason = reason
		close(c.closed)
	})
}

// closeWS performs the graceful close handshake exactly once, sending the close
// frame with the recorded code/reason. This also unblocks the readPump's
// pending Read. It MUST be called only from the writePump goroutine so it never
// races another writer, while coder/websocket permits it to run concurrently
// with the single reader.
func (c *Conn) closeWS() {
	c.closeWSOnce.Do(func() {
		if c.ws != nil {
			_ = c.ws.Close(c.closeCode, c.closeReason)
		}
	})
}

// Serve runs the connection until the peer closes it, a pump errors, or the
// Hub closes it. It blocks (call it from the HTTP handler goroutine) and always
// unregisters from the Hub before returning.
func (c *Conn) Serve(ctx context.Context) {
	if c.ws != nil {
		c.ws.SetReadLimit(c.cfg.ReadLimit)
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		c.writePump(ctx)
	})

	c.readPump(ctx)
	// readPump returned (peer close, read limit, or an already-triggered close).
	// Signal the writePump, which performs the graceful ws.Close on its way out.
	c.triggerClose(websocket.StatusNormalClosure, "")
	wg.Wait()

	c.hub.unregister(c)
}

// readPump is the single reader. It decodes control frames and answers pings
// inline; everything else is delegated to onMsg.
func (c *Conn) readPump(ctx context.Context) {
	log := logger.From(ctx)
	for {
		log.Infof("Reading from connection")
		typ, data, err := c.ws.Read(ctx)
		if err != nil {
			// Peer close, read limit, or context cancel. Signal the writePump so
			// it runs the graceful close on its way out.
			c.triggerClose(websocket.StatusNormalClosure, "")
			return
		}
		log.Infof("websocket.MessageText: %v", websocket.MessageText)
		if typ != websocket.MessageText {
			continue
		}

		log.Infof("Received message: %v", string(data))

		var in Inbound
		if err := json.Unmarshal(data, &in); err != nil {
			// Malformed frame is a transport concern; the module attaches
			// domain SOCKET_* codes to semantic errors (unknown topic, etc.).
			c.Send(Outbound{Type: TypeError, Event: "invalid_frame"})
			continue
		}

		if in.Type == TypePing {
			c.Send(Outbound{Type: TypePong})
			continue
		}
		if c.onMsg != nil {
			c.onMsg(c, in)
		}
	}
}

// writePump is the single writer: it drains the send queue and emits periodic
// pings. It owns the graceful ws.Close (via the deferred closeWS), which also
// unblocks the readPump. Any write or ping failure, a Hub-signalled close, or a
// cancelled context ends the loop.
func (c *Conn) writePump(ctx context.Context) {
	log := logger.From(ctx)

	ticker := time.NewTicker(c.cfg.PingInterval)
	defer ticker.Stop()
	defer c.closeWS()

	for {
		select {
		case <-ctx.Done():
			c.triggerClose(websocket.StatusGoingAway, "context done")
			return
		case <-c.closed:
			return
		case frame := <-c.send:
			log.Infof("Writing frame: %s", string(frame))

			wctx, cancel := context.WithTimeout(ctx, c.cfg.WriteTimeout)
			err := c.ws.Write(wctx, websocket.MessageText, frame)
			cancel()
			if err != nil {
				c.triggerClose(websocket.StatusInternalError, "write error")
				return
			}
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, c.cfg.WriteTimeout)
			err := c.ws.Ping(pctx)
			cancel()
			if err != nil {
				c.triggerClose(websocket.StatusPolicyViolation, "ping timeout")
				return
			}
		}
	}
}
