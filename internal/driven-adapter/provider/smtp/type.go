package smtp

// SMTPEmailRequest represents an SMTP email request payload
type SMTPEmailRequest struct {
	To          []string          `json:"to"`
	Cc          []string          `json:"cc,omitempty"`
	Bcc         []string          `json:"bcc,omitempty"`
	From        string            `json:"from"`
	Subject     string            `json:"subject"`
	TextBody    string            `json:"text_body,omitempty"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Attachments []SMTPAttachment  `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// SMTPAttachment represents an email attachment
type SMTPAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"` // Base64 encoded
}

// SMTPEmailResponse represents an SMTP email response
type SMTPEmailResponse struct {
	MessageID string   `json:"message_id"`
	Status    string   `json:"status"`
	Accepted  []string `json:"accepted,omitempty"`
	Rejected  []string `json:"rejected,omitempty"`
}
