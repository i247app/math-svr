package email

import "encoding/base64"

// encodeRawMessage base64url-encodes a RFC 2822 payload as required by the
// Gmail API's messages.send endpoint.
func encodeRawMessage(raw []byte) string {
	return base64.URLEncoding.EncodeToString(raw)
}
