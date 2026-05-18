package otp_delivery

// ChannelName identifies a registered OTP delivery channel. New channels
// (voice, push, in-app banner, WhatsApp, etc.) are added by appending a
// constant here, a *_deliverer.go file, and a factory case.
type ChannelName string

const (
	ChannelSMS   ChannelName = "sms"
	ChannelEmail ChannelName = "email"
)
