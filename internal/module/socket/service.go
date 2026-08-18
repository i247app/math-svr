// Package socket is the presentation layer for the realtime WebSocket channel:
// it accepts upgrades, authorizes topic subscriptions, and drives each
// connection's lifecycle over the infrastructure/socket Hub.
package socket

import (
	"context"

	"github.com/coder/websocket"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/socket"
)

// Service orchestrates connection lifecycle and topic authorization. It holds
// the concrete Hub (transport) and the topic Authorizer (policy).
type Service struct {
	hub            *socket.Hub
	authz          Authorizer
	originPatterns []string
}

// NewService builds the service. A nil authorizer falls back to
// DefaultAuthorizer (own-topics-only). originPatterns is passed to
// websocket.Accept as the CSRF origin allowlist.
func NewService(hub *socket.Hub, authz Authorizer, originPatterns []string) *Service {
	if authz == nil {
		authz = DefaultAuthorizer{}
	}
	return &Service{hub: hub, authz: authz, originPatterns: originPatterns}
}

// OriginPatterns exposes the accept-time origin allowlist to the handler.
func (s *Service) OriginPatterns() []string { return s.originPatterns }

// Connect takes an already-accepted socket, registers it for userID, and blocks
// serving it until close. It enforces the per-user connection cap.
func (s *Service) Connect(ctx context.Context, ws *websocket.Conn, userID int64) {
	log := logger.From(ctx)

	conn, err := s.hub.NewConn(ws, userID)
	if err != nil {
		_ = ws.Close(websocket.StatusTryAgainLater, "too many connections")
		log.Warnf("socket.connect_rejected uid=%d err=%v", userID, err)
		return
	}

	// Auto-subscribe the user's personal topics, then install the control-frame
	// dispatcher before serving so no early client frame is missed.
	s.hub.Subscribe(conn, personalTopic(userID))
	s.hub.Subscribe(conn, notificationsTopic(userID))
	conn.SetOnMessage(s.onMessage)

	log.Infof("socket.connected uid=%d conn=%s", userID, conn.ID())
	conn.Serve(ctx) // blocks until the connection closes
	log.Infof("socket.disconnected uid=%d conn=%s", userID, conn.ID())
}

// onMessage handles subscribe / unsubscribe control frames. Ping/pong and
// malformed frames are handled in the transport layer.
func (s *Service) onMessage(c *socket.Conn, in socket.Inbound) {
	switch in.Type {
	case socket.TypeSubscribe:
		if in.Topic == "" {
			s.sendErr(c, in.Topic, status.SOCKET_MISSING_TOPIC)
			return
		}
		if !s.authz.CanSubscribe(c.UserID(), in.Topic) {
			s.sendErr(c, in.Topic, status.SOCKET_TOPIC_FORBIDDEN)
			return
		}
		s.hub.Subscribe(c, in.Topic)
		s.sendAck(c, socket.TypeSubscribe, in.Topic)

	case socket.TypeUnsubscribe:
		if in.Topic == "" {
			s.sendErr(c, in.Topic, status.SOCKET_MISSING_TOPIC)
			return
		}
		s.hub.Unsubscribe(c, in.Topic)
		s.sendAck(c, socket.TypeUnsubscribe, in.Topic)

	default:
		s.sendErr(c, in.Topic, status.SOCKET_INVALID_MESSAGE)
	}
}

func (s *Service) sendAck(c *socket.Conn, event, topic string) {
	c.Send(socket.Outbound{
		Type:    socket.TypeAck,
		Event:   event,
		Topic:   topic,
		MStatus: int(status.SUCCESS),
	})
}

func (s *Service) sendErr(c *socket.Conn, topic string, code status.StatusCode) {
	c.Send(socket.Outbound{
		Type:    socket.TypeError,
		Topic:   topic,
		MStatus: int(code),
	})
}
