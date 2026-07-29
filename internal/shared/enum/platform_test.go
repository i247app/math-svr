package enum_test

import (
	"testing"

	"math-ai.com/math-ai/internal/shared/enum"
)

func TestParsePlatformType(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want enum.PlatformType
	}{
		{name: "lowercase ios", raw: "ios", want: enum.PlatformTypeIOS},
		{name: "lowercase android", raw: "android", want: enum.PlatformTypeAndroid},
		{name: "lowercase web", raw: "web", want: enum.PlatformTypeWeb},
		{name: "already uppercase", raw: "IOS", want: enum.PlatformTypeIOS},
		{name: "mixed case with whitespace", raw: "  AnDroid  ", want: enum.PlatformTypeAndroid},
		{name: "empty string falls back to unknown", raw: "", want: enum.PlatformTypeUnknown},
		{name: "unrecognized value falls back to unknown", raw: "harmonyos", want: enum.PlatformTypeUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := enum.ParsePlatformType(tc.raw)
			if got != tc.want {
				t.Fatalf("ParsePlatformType(%q) = %q, want %q", tc.raw, got, tc.want)
			}
			if !got.IsValid() {
				t.Fatalf("ParsePlatformType(%q) returned an invalid PlatformType: %q", tc.raw, got)
			}
		})
	}
}
