package twilio

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// MessagesService is the Messages resource of the Twilio REST API. It
// is constructed and owned by *Client; callers reach it via
// client.Messages.
//
// Pointer-style methods so the embedded *Client field can be unexported
// without forcing copies.
type MessagesService struct {
	c *Client
}

// CreateMessageParams is the input to MessagesService.Create.
//
// To and Body are required. Exactly one of From or MessagingServiceSID
// must resolve at send time, either supplied here or via the parent
// Client's defaults (Config.From / Config.MessagingServiceSID) — Twilio
// rejects requests with neither (error 21603).
//
// StatusCallback is a public webhook URL Twilio POSTs delivery
// transitions to. Optional; not used until the webhook feature lands.
type CreateMessageParams struct {
	To                  string
	Body                string
	From                string
	MessagingServiceSID string
	StatusCallback      string
}

// MessageResource mirrors the JSON shape Twilio returns on a successful
// Messages.Create. Fields the codebase does not consume today are still
// captured so a future caller can read them without changing the wire
// surface.
type MessageResource struct {
	SID          string  `json:"sid"`
	AccountSID   string  `json:"account_sid"`
	To           string  `json:"to"`
	From         string  `json:"from"`
	Body         string  `json:"body"`
	Status       string  `json:"status"`
	ErrorCode    *int    `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	DateCreated  string  `json:"date_created"`
	DateUpdated  string  `json:"date_updated"`
	URI          string  `json:"uri"`
}

func (s *MessagesService) SendByMessagingServiceSID(ctx context.Context, p CreateMessageParams) (*MessageResource, error) {
	if p.To == "" {
		return nil, errors.New("twilio: To is required")
	}
	if p.Body == "" {
		return nil, errors.New("twilio: Body is required")
	}
	if p.MessagingServiceSID == "" {
		return nil, errors.New("twilio: MessagingServiceSID is required")
	}

	form := url.Values{}
	form.Set("To", p.To)
	form.Set("Body", p.Body)
	form.Set("MessagingServiceSid", p.MessagingServiceSID)

	path := fmt.Sprintf("/2010-04-01/Accounts/%s/Messages.json", url.PathEscape(s.c.accountSID))
	resp, err := s.c.http.Post(ctx, path, form)
	if err != nil {
		return nil, fmt.Errorf("twilio: messages.create transport: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, parseAPIError(resp.StatusCode, resp.Bytes())
	}

	var out MessageResource
	if err := resp.JSON(&out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	return &out, nil
}

func (s *MessagesService) SendByFrom(ctx context.Context, p CreateMessageParams) (*MessageResource, error) {
	if p.To == "" {
		return nil, errors.New("twilio: To is required")
	}
	if p.Body == "" {
		return nil, errors.New("twilio: Body is required")
	}
	if p.From == "" {
		return nil, errors.New("twilio: From is required")
	}

	form := url.Values{}
	form.Set("To", p.To)
	form.Set("Body", p.Body)
	form.Set("From", p.From)

	path := fmt.Sprintf("/2010-04-01/Accounts/%s/Messages.json", url.PathEscape(s.c.accountSID))
	resp, err := s.c.http.Post(ctx, path, form)
	if err != nil {
		return nil, fmt.Errorf("twilio: messages.create transport: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, parseAPIError(resp.StatusCode, resp.Bytes())
	}

	var out MessageResource
	if err := resp.JSON(&out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	return &out, nil
}
