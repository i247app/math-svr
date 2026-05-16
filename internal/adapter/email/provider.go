package email

import "context"

type EmailProvider interface {
	Name() EmailProviderName
	Send(ctx context.Context, msg Message) error
}
