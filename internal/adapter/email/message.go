package email

import "errors"

// Attachment is a file to ship with the email. ContentType may be empty, in
// which case providers infer from Filename.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Message is the provider-agnostic email payload.
//
// From is optional; providers fall back to their configured default sender
// (SMTP From, Gmail SA impersonation subject) when it's empty. At least one
// of BodyText/BodyHTML must be set.
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	BodyText    string
	BodyHTML    string
	Attachments []Attachment
}

func (m Message) Validate() error {
	if len(m.To) == 0 {
		return errors.New("email: To is required")
	}
	if m.Subject == "" {
		return errors.New("email: Subject is required")
	}
	if m.BodyText == "" && m.BodyHTML == "" {
		return errors.New("email: BodyText or BodyHTML is required")
	}
	return nil
}
