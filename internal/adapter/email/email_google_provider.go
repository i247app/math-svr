package email

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"

	lib_email "math-ai.com/math-ai/internal/libs/email"
)

// EmailGoogleProvider adapts the Gmail-API-backed libs/email client to the
// EmailProvider contract.
type EmailGoogleProvider struct {
	client *lib_email.GoogleEmailClient
}

func NewEmailGoogleProvider(client *lib_email.GoogleEmailClient) *EmailGoogleProvider {
	return &EmailGoogleProvider{client: client}
}

func (e *EmailGoogleProvider) Name() EmailProviderName {
	return ProviderGoogle
}

func (e *EmailGoogleProvider) Send(ctx context.Context, msg Message) error {
	if err := msg.Validate(); err != nil {
		return err
	}

	raw, err := e.buildMIME(msg)
	if err != nil {
		return err
	}
	return e.client.SendRaw(ctx, raw)
}

// buildMIME renders the Message as a RFC 2822 payload. If attachments are
// present, the body is wrapped in multipart/mixed; otherwise a single
// text/plain, text/html, or multipart/alternative body is emitted.
func (e *EmailGoogleProvider) buildMIME(msg Message) ([]byte, error) {
	from := msg.From
	if from == "" {
		from = e.client.Sender()
	}
	if from == "" {
		return nil, fmt.Errorf("google provider: From is required")
	}

	var buf bytes.Buffer
	headers := textproto.MIMEHeader{}
	headers.Set("MIME-Version", "1.0")
	headers.Set("From", from)
	headers.Set("To", strings.Join(msg.To, ", "))
	if len(msg.Cc) > 0 {
		headers.Set("Cc", strings.Join(msg.Cc, ", "))
	}
	if len(msg.Bcc) > 0 {
		headers.Set("Bcc", strings.Join(msg.Bcc, ", "))
	}
	headers.Set("Subject", mime.QEncoding.Encode("utf-8", msg.Subject))

	if len(msg.Attachments) == 0 {
		return writeSimpleBody(&buf, headers, msg)
	}
	return writeMixedBody(&buf, headers, msg)
}

func writeSimpleBody(buf *bytes.Buffer, headers textproto.MIMEHeader, msg Message) ([]byte, error) {
	switch {
	case msg.BodyHTML != "" && msg.BodyText != "":
		return writeAlternativeBody(buf, headers, msg.BodyText, msg.BodyHTML)
	case msg.BodyHTML != "":
		headers.Set("Content-Type", `text/html; charset="utf-8"`)
		writeHeaders(buf, headers)
		buf.WriteString(msg.BodyHTML)
	default:
		headers.Set("Content-Type", `text/plain; charset="utf-8"`)
		writeHeaders(buf, headers)
		buf.WriteString(msg.BodyText)
	}
	return buf.Bytes(), nil
}

func writeAlternativeBody(buf *bytes.Buffer, headers textproto.MIMEHeader, text, html string) ([]byte, error) {
	mw := multipart.NewWriter(buf)
	headers.Set("Content-Type", `multipart/alternative; boundary="`+mw.Boundary()+`"`)
	writeHeaders(buf, headers)

	if err := writePart(mw, `text/plain; charset="utf-8"`, text); err != nil {
		return nil, err
	}
	if err := writePart(mw, `text/html; charset="utf-8"`, html); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeMixedBody(buf *bytes.Buffer, headers textproto.MIMEHeader, msg Message) ([]byte, error) {
	mw := multipart.NewWriter(buf)
	headers.Set("Content-Type", `multipart/mixed; boundary="`+mw.Boundary()+`"`)
	writeHeaders(buf, headers)

	// Body section as a nested part.
	var body bytes.Buffer
	bodyHeaders := textproto.MIMEHeader{}
	if _, err := writeSimpleBody(&body, bodyHeaders, Message{BodyText: msg.BodyText, BodyHTML: msg.BodyHTML}); err != nil {
		return nil, err
	}
	partHeader := textproto.MIMEHeader{}
	for k, v := range bodyHeaders {
		partHeader[k] = v
	}
	part, err := mw.CreatePart(partHeader)
	if err != nil {
		return nil, err
	}
	// writeSimpleBody wrote its own headers + body into `body`; strip the
	// duplicate headers by copying only past the blank line.
	if _, err := part.Write(stripInnerHeaders(body.Bytes())); err != nil {
		return nil, err
	}

	for _, att := range msg.Attachments {
		ct := att.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		attHeader := textproto.MIMEHeader{}
		attHeader.Set("Content-Type", ct)
		attHeader.Set("Content-Transfer-Encoding", "base64")
		attHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
		w, err := mw.CreatePart(attHeader)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(base64Chunked(att.Data)); err != nil {
			return nil, err
		}
	}

	if err := mw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeHeaders(buf *bytes.Buffer, h textproto.MIMEHeader) {
	for k, vs := range h {
		for _, v := range vs {
			buf.WriteString(k)
			buf.WriteString(": ")
			buf.WriteString(v)
			buf.WriteString("\r\n")
		}
	}
	buf.WriteString("\r\n")
}

func writePart(mw *multipart.Writer, contentType, body string) error {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", contentType)
	w, err := mw.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(body))
	return err
}

// stripInnerHeaders drops the headers + blank line from a rendered part so
// callers can re-emit it inside a different MIME container.
func stripInnerHeaders(data []byte) []byte {
	idx := bytes.Index(data, []byte("\r\n\r\n"))
	if idx < 0 {
		return data
	}
	return data[idx+4:]
}
