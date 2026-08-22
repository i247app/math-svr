package chat

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// The preview is written into a VARCHAR(255) column and rendered in the inbox.
// Truncating on a byte boundary would split a Vietnamese character in half and
// store invalid UTF-8, which surfaces as a replacement glyph in the app.
func TestBuildPreview(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantEq  bool // preview must equal the input unchanged
	}{
		{name: "short ascii passes through", content: "hello", wantEq: true},
		{name: "short vietnamese passes through", content: "Chào em, hôm nay học tốt không?", wantEq: true},
		{name: "exactly at the limit passes through", content: strings.Repeat("a", previewRuneLimit), wantEq: true},
		{name: "long ascii is truncated", content: strings.Repeat("a", previewRuneLimit+50)},
		{name: "long vietnamese is truncated", content: strings.Repeat("ữ", previewRuneLimit+50)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildPreview(tc.content)

			if !utf8.ValidString(got) {
				t.Fatalf("preview is not valid UTF-8: %q", got)
			}
			if tc.wantEq {
				if got != tc.content {
					t.Fatalf("preview = %q, want it unchanged", got)
				}
				return
			}
			if got == tc.content {
				t.Fatalf("expected truncation, got the input back")
			}
			// limit runes plus the ellipsis.
			if n := utf8.RuneCountInString(got); n != previewRuneLimit+1 {
				t.Errorf("preview rune count = %d, want %d", n, previewRuneLimit+1)
			}
			if !strings.HasSuffix(got, "…") {
				t.Errorf("truncated preview should end with an ellipsis, got %q", got)
			}
			// The column is VARCHAR(255): even all-multi-byte content must fit.
			if len(got) > 255 {
				t.Errorf("preview is %d bytes, exceeds the column width", len(got))
			}
		})
	}
}
