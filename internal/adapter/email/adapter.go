package email

import (
	"context"
	"errors"
	"fmt"
)

// Adapter is the entry point callers use to dispatch email. It holds one or
// more registered providers and a designated default used by Send.
type Adapter struct {
	providers   map[EmailProviderName]EmailProvider
	defaultName EmailProviderName
}

func NewAdapter() *Adapter {
	return &Adapter{providers: make(map[EmailProviderName]EmailProvider)}
}

// Register adds a provider to the adapter. The first provider registered
// becomes the default; callers may override that with SetDefault.
func (a *Adapter) Register(provider EmailProvider) {
	if a.providers == nil {
		a.providers = make(map[EmailProviderName]EmailProvider)
	}
	a.providers[provider.Name()] = provider
	if a.defaultName == "" {
		a.defaultName = provider.Name()
	}
}

// SetDefault selects which registered provider Send dispatches to. It returns
// an error if the provider name has not been registered.
func (a *Adapter) SetDefault(name EmailProviderName) error {
	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("email: provider %q is not registered", name)
	}
	a.defaultName = name
	return nil
}

// Send dispatches msg through the adapter's default provider.
func (a *Adapter) Send(ctx context.Context, msg Message) error {
	if a.defaultName == "" {
		return errors.New("email: no default provider configured")
	}
	return a.SendVia(ctx, a.defaultName, msg)
}

// SendVia dispatches msg through the provider registered under name.
func (a *Adapter) SendVia(ctx context.Context, name EmailProviderName, msg Message) error {
	provider, ok := a.providers[name]
	if !ok {
		return fmt.Errorf("email: provider %q is not registered", name)
	}
	return provider.Send(ctx, msg)
}
