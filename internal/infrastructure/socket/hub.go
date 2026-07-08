package socket

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// ErrTooManyConnections is returned by NewConn when the user is already at the
// per-user connection cap. The caller must reject the handshake.
var ErrTooManyConnections = errors.New("socket: too many connections for user")

// Config tunes the Hub and the default per-connection settings.
type Config struct {
	MaxConnsPerUser int
	BufferSize      int
	PingInterval    time.Duration
	WriteTimeout    time.Duration
	ReadLimit       int64
}

// Hub owns the connection registry and topic fan-out. All map access is guarded
// by mu; frame delivery happens outside the lock via the connections' non-
// blocking send queues, so one slow client can never stall the Hub.
type Hub struct {
	mu sync.RWMutex

	conns  map[string]*Conn            // id -> conn
	byUser map[int64]map[string]*Conn  // uid -> id -> conn
	topics map[string]map[string]*Conn // topic -> id -> conn
	subs   map[string]map[string]bool  // id -> set of subscribed topics

	maxPerUser int
	connCfg    ConnConfig
	nextID     atomic.Uint64
}

// NewHub builds an empty Hub. Zero-valued config fields use package defaults.
func NewHub(cfg Config) *Hub {
	connCfg := ConnConfig{
		BufferSize:   cfg.BufferSize,
		PingInterval: cfg.PingInterval,
		WriteTimeout: cfg.WriteTimeout,
		ReadLimit:    cfg.ReadLimit,
	}.withDefaults()

	maxPerUser := cfg.MaxConnsPerUser
	if maxPerUser <= 0 {
		maxPerUser = defaultMaxConnsPerUser
	}

	return &Hub{
		conns:      make(map[string]*Conn),
		byUser:     make(map[int64]map[string]*Conn),
		topics:     make(map[string]map[string]*Conn),
		subs:       make(map[string]map[string]bool),
		maxPerUser: maxPerUser,
		connCfg:    connCfg,
	}
}

const defaultMaxConnsPerUser = 5

// NewConn registers a fresh connection for userID and returns it ready to
// Serve. It enforces the per-user cap. ws may be nil in transport-only tests.
func (h *Hub) NewConn(ws *websocket.Conn, userID int64) (*Conn, error) {
	id := strconv.FormatUint(h.nextID.Add(1), 10)
	c := &Conn{
		id:     id,
		userID: userID,
		ws:     ws,
		hub:    h,
		cfg:    h.connCfg,
		send:   make(chan []byte, h.connCfg.BufferSize),
		closed: make(chan struct{}),
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if existing := h.byUser[userID]; len(existing) >= h.maxPerUser {
		return nil, ErrTooManyConnections
	}

	h.conns[id] = c
	if h.byUser[userID] == nil {
		h.byUser[userID] = make(map[string]*Conn)
	}
	h.byUser[userID][id] = c
	h.subs[id] = make(map[string]bool)
	return c, nil
}

// unregister removes a connection and all of its subscriptions. Idempotent.
func (h *Hub) unregister(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.conns[c.id]; !ok {
		return
	}
	delete(h.conns, c.id)

	if userConns := h.byUser[c.userID]; userConns != nil {
		delete(userConns, c.id)
		if len(userConns) == 0 {
			delete(h.byUser, c.userID)
		}
	}

	for topic := range h.subs[c.id] {
		if members := h.topics[topic]; members != nil {
			delete(members, c.id)
			if len(members) == 0 {
				delete(h.topics, topic)
			}
		}
	}
	delete(h.subs, c.id)
}

// Subscribe adds c to a topic's fan-out set. No-op if already subscribed or the
// connection has been unregistered.
func (h *Hub) Subscribe(c *Conn, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.conns[c.id]; !ok {
		return
	}
	if h.topics[topic] == nil {
		h.topics[topic] = make(map[string]*Conn)
	}
	h.topics[topic][c.id] = c
	h.subs[c.id][topic] = true
}

// Unsubscribe removes c from a topic's fan-out set.
func (h *Hub) Unsubscribe(c *Conn, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if members := h.topics[topic]; members != nil {
		delete(members, c.id)
		if len(members) == 0 {
			delete(h.topics, topic)
		}
	}
	if set := h.subs[c.id]; set != nil {
		delete(set, topic)
	}
}

// Publish fans an event out to every connection subscribed to topic. The frame
// is marshaled once. A connection under backpressure is closed rather than
// blocking the fan-out. Returns the number of connections the frame reached.
func (h *Hub) Publish(topic, event string, data any) int {
	frame, err := json.Marshal(Outbound{
		Type:  TypeEvent,
		Topic: topic,
		Event: event,
		Data:  data,
	})
	if err != nil {
		return 0
	}
	return h.deliver(h.targets(h.topics[topic]), frame)
}

// BroadcastUser fans an event out to every connection owned by userID
// regardless of topic subscription (used for direct, user-addressed pushes).
func (h *Hub) BroadcastUser(userID int64, event string, data any) int {
	frame, err := json.Marshal(Outbound{
		Type:  TypeEvent,
		Event: event,
		Data:  data,
	})
	if err != nil {
		return 0
	}
	return h.deliver(h.targets(h.byUser[userID]), frame)
}

// targets snapshots a connection set under the read lock so delivery happens
// without holding mu.
func (h *Hub) targets(set map[string]*Conn) []*Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(set) == 0 {
		return nil
	}
	out := make([]*Conn, 0, len(set))
	for _, c := range set {
		out = append(out, c)
	}
	return out
}

// deliver enqueues frame to each target, closing any that are backed up.
func (h *Hub) deliver(targets []*Conn, frame []byte) int {
	delivered := 0
	for _, c := range targets {
		if c.enqueue(frame) {
			delivered++
		} else {
			c.triggerClose(websocket.StatusPolicyViolation, "slow consumer")
		}
	}
	return delivered
}

// CloseAll signals every connection to close (used on graceful shutdown). The
// connections' own Serve loops perform the unregister and socket close.
func (h *Hub) CloseAll(code websocket.StatusCode, reason string) {
	for _, c := range h.targets(h.conns) {
		c.triggerClose(code, reason)
	}
}

// Shutdown closes every connection with StatusGoingAway. Call from the graceful
// shutdown hook; each connection's Serve loop performs its own cleanup.
func (h *Hub) Shutdown(reason string) {
	h.CloseAll(websocket.StatusGoingAway, reason)
}

// Stats is a point-in-time snapshot for metrics/observability.
type Stats struct {
	Connections int
	Users       int
	Topics      int
}

// Stats returns current registry sizes.
func (h *Hub) Stats() Stats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return Stats{
		Connections: len(h.conns),
		Users:       len(h.byUser),
		Topics:      len(h.topics),
	}
}
