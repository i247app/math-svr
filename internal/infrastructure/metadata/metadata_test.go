package metadata

import (
	"encoding/json"
	"testing"

	"math-ai.com/math-ai/internal/shared/enum"
)

func TestClientLanguage_ToEnumLanguage(t *testing.T) {
	cases := []struct {
		in   ClientLanguage
		want enum.LanguageType
	}{
		{"vi", enum.LanguageTypeVietnamese},
		{"vi-VN", enum.LanguageTypeVietnamese},
		{"en", enum.LanguageTypeEnglish},
		{"en-EN", enum.LanguageTypeEnglish},
		{"", enum.LanguageTypeVietnamese},      // default → VN (project default)
		{"fr-FR", enum.LanguageTypeVietnamese}, // unknown → VN
	}
	for _, c := range cases {
		if got := c.in.ToEnumLanguage(); got != c.want {
			t.Errorf("ClientLanguage(%q).ToEnumLanguage() = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestRequestMetadata_UnmarshalClientPayload locks the struct's JSON tags to
// the mobile client's actual `metadata` payload shape.
func TestRequestMetadata_UnmarshalClientPayload(t *testing.T) {
	payload := `{
		"device_uuid": "android-device-id",
		"device_name": "Pixel 8",
		"device_push_token": "fcm-push-token",
		"model_name": "google",
		"platform": "android",
		"system_version": "35",
		"timestamp": "2026-07-10T08:30:00.000Z",
		"version": "1.0.0",
		"build": "12",
		"app_version": "Version: 1.0.0 + 12",
		"language": "vi",
		"accept_language": "vi",
		"accept": "application/json",
		"content_type": "application/json",
		"authorization": "Bearer jwt-value",
		"ip_address": "192.168.1.10"
	}`

	var m RequestMetadata
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	checks := map[string]struct{ got, want string }{
		"device_uuid":       {m.DeviceID, "android-device-id"},
		"device_name":       {m.DeviceName, "Pixel 8"},
		"device_push_token": {m.DevicePushToken, "fcm-push-token"},
		"model_name":        {m.ModelName, "google"},
		"platform":          {m.Platform, "android"},
		"system_version":    {m.OSVersion, "35"},
		"version":           {m.Version, "1.0.0"},
		"build":             {m.Build, "12"},
		"app_version":       {m.AppVersion, "Version: 1.0.0 + 12"},
		"language":          {string(m.Language), "vi"},
		"accept_language":   {m.AcceptLanguage, "vi"},
		"content_type":      {m.ContentType, "application/json"},
		"authorization":     {m.Authorization, "Bearer jwt-value"},
		"ip_address":        {m.IPAddress, "192.168.1.10"},
	}
	for field, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", field, c.got, c.want)
		}
	}

	if m.Language.ToEnumLanguage() != enum.LanguageTypeVietnamese {
		t.Errorf("language %q did not resolve to Vietnamese", m.Language)
	}
}
