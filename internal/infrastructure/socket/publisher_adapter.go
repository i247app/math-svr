package socket

import (
	"context"

	appsocket "math-ai.com/math-ai/internal/application/socket"
)

// HubPublisher adapts the in-memory Hub to the application/socket.Publisher
// port. It drops the ctx (the in-memory fan-out is synchronous and never
// blocks) and always returns nil — a networked implementation would honour both.
type HubPublisher struct {
	hub *Hub
}

// NewPublisher wraps a Hub as a Publisher for wiring into the app Resource.
func NewPublisher(hub *Hub) *HubPublisher { return &HubPublisher{hub: hub} }

func (p *HubPublisher) Publish(_ context.Context, topic, event string, data any) error {
	p.hub.Publish(topic, event, data)
	return nil
}

func (p *HubPublisher) BroadcastUser(_ context.Context, userID int64, event string, data any) error {
	p.hub.BroadcastUser(userID, event, data)
	return nil
}

// Compile-time guarantee that HubPublisher satisfies the port.
var _ appsocket.Publisher = (*HubPublisher)(nil)
