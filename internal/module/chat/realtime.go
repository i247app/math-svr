package chat

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/chat"
	chatDomain "math-ai.com/math-ai/internal/domain/chat"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// MessageCreatedEvent is the wire name clients key off. Treat it as part of
// the public contract: renaming it silently stops every deployed app from
// rendering incoming messages.
const MessageCreatedEvent = "chat.message.created"

// MessagePublisher is the narrow slice of the realtime channel this module
// needs. It is an interface rather than the appsocket.Publisher type so the
// module can be constructed and tested with realtime switched off.
type MessagePublisher interface {
	BroadcastUser(ctx context.Context, userID int64, event string, data any) error
}

// publishMessage pushes a new message to every participant except the sender.
//
// It addresses recipients by user id rather than through a topic. A direct
// thread has exactly two participants and both are known here, so a
// subscription would add a round trip and an authorization decision without
// changing what gets delivered. When group chat arrives, this is the place to
// switch to a conversation:{id} topic — the Hub and the Publisher port already
// support it.
//
// Delivery is best-effort by design: the row is committed, so a failure here
// costs a client one refresh, not a message.
func (s *Service) publishMessage(ctx context.Context, m *chatDomain.Message, payload *dto.MessageResponse) {
	if s.realtime == nil || m == nil {
		return
	}
	log := logger.From(ctx)

	participants, err := s.participantRepo.ListByConversationId(ctx, &chatDomain.ListParticipantsParams{
		ConversationId: m.ConversationId(),
	})
	if err != nil {
		log.Warnf("chat.publish_participants_failed conversation_id=%d err=%v", m.ConversationId(), err)
		return
	}

	senderProfileID := int64(0)
	if m.SenderProfileId() != nil {
		senderProfileID = *m.SenderProfileId()
	}

	for _, p := range participants {
		if p.ProfileId() == senderProfileID {
			continue
		}
		if err := s.realtime.BroadcastUser(ctx, p.UserId(), MessageCreatedEvent, payload); err != nil {
			// Never log the message body — a chat payload carries a child's
			// name and whatever they wrote.
			log.Warnf("chat.publish_failed conversation_id=%d message_id=%d uid=%d err=%v",
				m.ConversationId(), m.MessageId(), p.UserId(), err)
		}
	}
}
