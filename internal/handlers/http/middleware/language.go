package middleware

import (
	"net/http"
	"strings"

	"math-ai.com/math-ai/internal/shared/utils/language"
)

// LanguageMiddleware extracts language preference from HTTP headers and sets it in context
// Supports:
// - Accept-Language header (standard HTTP header)
// - X-Language header (custom header for explicit language selection)
//
// Language codes are normalized to uppercase 2-letter codes (e.g., "EN", "VN", "FR")
func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try custom header first (explicit language selection)
		lang := r.Header.Get("X-Language")

		// Fall back to Accept-Language header
		if lang == "" {
			lang = r.Header.Get("Accept-Language")
		}

		// Normalize and validate language code
		lang = normalizeLanguageCode(lang)

		// Set language in context
		ctx := language.SetLanguage(r.Context(), lang)

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// normalizeLanguageCode converts language codes to standardized 2-letter uppercase format
// Examples:
// - "en" -> "EN"
// - "en-US" -> "EN"
// - "vi" -> "VN"
// - "vi-VN" -> "VN"
// - "" -> "EN" (default)
func normalizeLanguageCode(lang string) string {
	if lang == "" {
		return language.DefaultLanguage
	}

	// Convert to uppercase
	lang = strings.ToUpper(lang)

	// Extract primary language code (before hyphen or underscore)
	if idx := strings.IndexAny(lang, "-_"); idx > 0 {
		lang = lang[:idx]
	}

	// Validate length
	if len(lang) < 2 {
		return language.DefaultLanguage
	}

	// Take first 2 characters
	lang = lang[:2]

	// Map some common codes to our internal codes
	switch lang {
	case "VI":
		return "VN" // Vietnamese
	case "EN":
		return "EN" // English
	case "FR":
		return "FR" // French
	default:
		return lang
	}
}
