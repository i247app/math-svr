package email

import "encoding/base64"

// base64Chunked standard-base64-encodes data with CRLF line breaks every
// 76 characters per RFC 2045. Used for attachment bodies in MIME multipart
// messages built for the Gmail API.
func base64Chunked(data []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString(data)
	const lineLen = 76
	if len(encoded) <= lineLen {
		return []byte(encoded)
	}
	var out []byte
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		out = append(out, encoded[i:end]...)
		out = append(out, '\r', '\n')
	}
	return out
}
