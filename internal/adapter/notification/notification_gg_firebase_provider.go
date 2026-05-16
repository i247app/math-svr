package notification

type FirebaseProvider struct {
}

func (f *FirebaseProvider) Name() NotificationProviderName {
	return ProviderFirebase
}
