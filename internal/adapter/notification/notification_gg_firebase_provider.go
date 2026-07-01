package notification

import (
	"context"

	"math-ai.com/math-ai/internal/libs/firebase"
)

// FirebaseProvider sends push notifications through Firebase Cloud Messaging.
// It is a thin translator: PushMessage in, libs/firebase call, SendResult out.
type FirebaseProvider struct {
	client *firebase.Client
}

func NewFirebaseProvider(client *firebase.Client) *FirebaseProvider {
	return &FirebaseProvider{client: client}
}

func (f *FirebaseProvider) Name() NotificationProviderName {
	return ProviderFirebase
}

func (f *FirebaseProvider) Send(ctx context.Context, msg PushMessage) (*SendResult, error) {
	res, err := f.client.SendMulticast(ctx, msg.Tokens, firebase.PushPayload{
		Title: msg.Title,
		Body:  msg.Body,
		Data:  msg.Data,
	})
	if err != nil {
		return nil, err
	}
	return &SendResult{
		SuccessCount:  res.SuccessCount,
		FailureCount:  res.FailureCount,
		InvalidTokens: res.InvalidTokens,
	}, nil
}
