package notification

import (
	"context"

	domain "math-ai.com/math-ai/internal/domain/notification"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	notiProvider "math-ai.com/math-ai/internal/provider/notification"
	"math-ai.com/math-ai/internal/shared/utils"
)

// pushService is the typed seam over the notification provider — it keeps the
// FCM wiring out of the orchestration service (mirrors exercise/llm_service.go).
// The provider pointer is nil when NOTIFICATION_PROVIDER is ""/"disabled";
// enabled() lets callers skip delivery cleanly.
type pushService struct {
	provider *notiProvider.Registry
}

func newPushService(provider *notiProvider.Registry) *pushService {
	return &pushService{provider: provider}
}

func (p *pushService) enabled() bool { return p.provider != nil }

// send fans n out to tokens through the provider, mapping the persisted
// notification onto the push payload. action_type / action_data ride along in
// the data map so the mobile client can deep-link.
func (p *pushService) send(ctx context.Context, tokens []string, n *domain.Notification) (*notiProvider.SendResult, error) {
	log := logger.From(ctx)
	data := map[string]string{
		"notification_id": utils.Int64ToString(n.NotificationId()),
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

	return p.provider.Send(ctx, notiProvider.PushMessage{
		Tokens: tokens,
		Title:  n.Title(),
		Body:   n.ShortText(),
		Data:   data,
	})
}
