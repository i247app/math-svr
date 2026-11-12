package notification

import "errors"

// PushMessage is the provider-agnostic push payload. Tokens is the set of
// device registration tokens to fan out to; Data is the optional custom
// key/value map the mobile client uses for deep-linking (action_type,
// action_data, notification_id, ...).
type PushMessage struct {
	Tokens []string
	Title  string
	Body   string
	Data   map[string]string
}

func (m PushMessage) Validate() error {
	if len(m.Tokens) == 0 {
		return errors.New("notification: at least one token is required")
	}
	if m.Title == "" && m.Body == "" {
		return errors.New("notification: Title or Body is required")
	}
	return nil
}

// SendResult summarises a fan-out send. InvalidTokens lists registration
// tokens the provider rejected as dead (unregistered / malformed) so the
// caller can prune them from ma_devices.
type SendResult struct {
	SuccessCount  int
	FailureCount  int
	InvalidTokens []string
}
