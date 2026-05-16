package notification

import "fmt"

type Adapter struct {
	providers   map[NotificationProviderName]NotificationProvider
	defaultName NotificationProviderName
}

// Register adds a provider to the adapter. The first provider registered
// becomes the default; callers may override that with SetDefault.
func (a *Adapter) Register(provider NotificationProvider) {
	if a.providers == nil {
		a.providers = make(map[NotificationProviderName]NotificationProvider)
	}
	a.providers[provider.Name()] = provider
	if a.defaultName == "" {
		a.defaultName = provider.Name()
	}
}

// SetDefault selects which registered provider Send dispatches to. It returns
// an error if the provider name has not been registered.
func (a *Adapter) SetDefault(name NotificationProviderName) error {
	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("notification: provider %q is not registered", name)
	}
	a.defaultName = name
	return nil
}
