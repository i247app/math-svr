// Package socket is the presentation layer for the realtime WebSocket channel:
// it accepts upgrades, authorizes topic subscriptions, and drives each
// connection's lifecycle over the infrastructure/socket Hub.
package socket

import (
	"context"

	"github.com/coder/websocket"

	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/socket"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Service orchestrates connection lifecycle and topic authorization. It holds
// the concrete Hub (transport) and the topic Authorizer (policy).
type Service struct {
	hub            *socket.Hub
	authz          Authorizer
	originPatterns []string
	presence       PresenceTracker
}

// NewService builds the service. A nil authorizer falls back to
// DefaultAuthorizer (own-topics-only). originPatterns is passed to
// websocket.Accept as the CSRF origin allowlist. presence may be nil, which
// disables online/offline tracking without affecting the channel.
func NewService(hub *socket.Hub, authz Authorizer, originPatterns []string, presence PresenceTracker) *Service {
	if authz == nil {
		authz = DefaultAuthorizer{}
	}
	return &Service{hub: hub, authz: authz, originPatterns: originPatterns, presence: presence}
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

	s.markOnline(ctx, userID)

	log.Infof("socket.connected uid=%d conn=%s", userID, conn.ID())
	conn.Serve(ctx) // blocks until the connection closes
	log.Infof("socket.disconnected uid=%d conn=%s", userID, conn.ID())

	// Deliberately uses context.WithoutCancel: by the time Serve returns, ctx
	// is usually already cancelled (that is what ended the connection), and a
	// cancelled context would abort the very write that marks the user
	// offline — leaving them permanently green.
	s.markOffline(context.WithoutCancel(ctx), userID)
}

// markOnline / markOffline never propagate their error. A presence write is
// bookkeeping for a status dot; failing it is not a reason to refuse or tear
// down a working realtime connection, so the error is logged and swallowed.
func (s *Service) markOnline(ctx context.Context, userID int64) {
	if s.presence == nil {
		return
	}
	deviceUUID := utils.OptionalString(metadata.GetDeviceID(ctx))
	platform := utils.OptionalString(metadata.GetPlatform(ctx))

	if _, err := s.presence.MarkOnline(ctx, userID, deviceUUID, platform); err != nil {
		logger.From(ctx).Warnf("socket.presence_online_failed uid=%d err=%v", userID, err)
	}
}

func (s *Service) markOffline(ctx context.Context, userID int64) {
	if s.presence == nil {
		return
	}
	if _, err := s.presence.MarkOffline(ctx, userID); err != nil {
		logger.From(ctx).Warnf("socket.presence_offline_failed uid=%d err=%v", userID, err)
	}
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
