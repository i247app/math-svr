package email

type EmailProviderName string

const (
	ProviderGomail EmailProviderName = "gomail"
	ProviderGoogle EmailProviderName = "google"
)
