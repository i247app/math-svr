package notification

import (
	"context"
	"strconv"

	notifAdapter "math-ai.com/math-ai/internal/adapter/notification"
	domain "math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// pushService is the typed seam over the notification adapter — it keeps the
// FCM wiring out of the orchestration service (mirrors quiz/bot_service.go).
// The adapter pointer is nil when NOTIFICATION_PROVIDER is ""/"disabled";
// enabled() lets callers skip delivery cleanly.
type pushService struct {
	adapter *notifAdapter.Adapter
}

func newPushService(adapter *notifAdapter.Adapter) *pushService {
	return &pushService{adapter: adapter}
}

func (p *pushService) enabled() bool { return p.adapter != nil }

// send fans n out to tokens through the adapter, mapping the persisted
// notification onto the push payload. action_type / action_data ride along in
// the data map so the mobile client can deep-link.
func (p *pushService) send(ctx context.Context, tokens []string, n *domain.Notification) (*notifAdapter.SendResult, error) {
	log := logger.From(ctx)
	data := map[string]string{
		"notification_id": strconv.FormatInt(n.NotificationId(), 10),
	}
	if n.ActionType() != nil {
		data["action_type"] = *n.ActionType()
	}
	if n.ActionData() != nil {
		data["action_data"] = *n.ActionData()
	}
	if n.Category() != nil {
		data["category"] = *n.Category()
	}

	for _, token := range tokens {
		log.Infof("FCM to token: %s", token)
	}

	return p.adapter.Send(ctx, notifAdapter.PushMessage{
		Tokens: tokens,
		Title:  n.Title(),
		Body:   n.ShortText(),
		Data:   data,
	})
}
