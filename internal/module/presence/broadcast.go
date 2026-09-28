package presence

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// PresenceChangedEvent is the wire name for an online/offline transition.
// Part of the client contract — renaming it stops deployed apps updating the
// dot. Lives next to its producer, like chat.MessageCreatedEvent.
const PresenceChangedEvent = "presence.changed"

// Publisher is the slice of the realtime channel this module needs.
type Publisher interface {
	BroadcastUser(ctx context.Context, uid int64, event string, data any) error
}

// PresenceChangedPayload is what classmates receive. It carries the user id
// rather than a profile id because presence is an account-level fact; the
// client already knows which profiles map to which user from the member list.
type PresenceChangedPayload struct {
	UID        int64  `json:"uid"`
	IsOnline   bool   `json:"is_online"`
	State      string `json:"state"`
	LastSeenDt string `json:"last_seen_dt,omitempty"`
}

// announceOnline tells classmates the user came online, unless a pending
// offline broadcast was cancelled — in that case nobody was ever told they
// left, so there is nothing to announce.
func (s *Service) announceOnline(ctx context.Context, uid int64) {
	if s.debounce.cancel(uid) {
		return
	}
	s.broadcast(ctx, uid, true, "")
}

// announceOffline waits out the debounce window, then re-reads the truth
// before telling anyone.
//
// The re-read is the point: the user may have reconnected on another device
// during the window, in which case the count is back above zero and no offline
// event should go out at all.
func (s *Service) announceOffline(ctx context.Context, uid int64) {
	// The request context dies with the connection that triggered this, so the
	// timer gets a detached one — otherwise the callback would fire into an
	// already-cancelled context and every read inside it would fail.
	detached := context.WithoutCancel(ctx)

	s.debounce.schedule(uid, func() {
		p, err := s.repo.FindByUid(detached, uid)
		if err != nil {
			logger.From(detached).Warnf("presence.broadcast_read_failed uid=%d err=%v", uid, err)
			return
		}
		if p == nil || p.IsOnline() {
			return // came back during the window
		}
		lastSeen := ""
		if !p.LastSeenDt().Time.IsZero() {
			lastSeen = p.LastSeenDt().Time.UTC().Format(time.RFC3339)
		}
		s.broadcast(detached, uid, false, lastSeen)
	})
}

// broadcast fans the transition out to everyone sharing a classroom with the
// user.
//
// Addressed per user id rather than through a classroom topic: that needs no
// subscription and no membership-aware Authorizer, which stays deny-by-default.
// If a classroom ever grows past a few hundred members, switch to a
// classroom:{id} topic — the Hub already supports it.
func (s *Service) broadcast(ctx context.Context, uid int64, online bool, lastSeen string) {
	if s.publisher == nil || s.classroomMemberRepo == nil {
		return
	}
	log := logger.From(ctx)

	peers, err := s.classroomMemberRepo.ListPeerUidsByUid(ctx, uid)
	if err != nil {
		log.Warnf("presence.peers_failed uid=%d err=%v", uid, err)
		return
	}
	if len(peers) == 0 {
		return
	}

	state := "OFFLINE"
	if online {
		state = "ONLINE"
	}
	payload := &PresenceChangedPayload{
		UID:        uid,
		IsOnline:   online,
		State:      state,
		LastSeenDt: lastSeen,
	}

	for _, peerId := range peers {
		// Best-effort, like every realtime publish: the database is
		// authoritative and the member list re-reads it on next open.
		if err := s.publisher.BroadcastUser(ctx, peerId, PresenceChangedEvent, payload); err != nil {
			log.Warnf("presence.broadcast_failed uid=%d peer=%d err=%v", uid, peerId, err)
		}
	}
}

// Shutdown cancels pending offline broadcasts so no timer fires into a
// half-torn-down process.
func (s *Service) Shutdown() {
	if s.debounce != nil {
		s.debounce.stopAll()
	}
}
