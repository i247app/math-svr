package language

import (
	"context"
)

type contextKey string

const (
	languageContextKey contextKey = "language"
	DefaultLanguage    string     = "EN"
)

// SetLanguage sets the language in the context
func SetLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, languageContextKey, language)
}

// GetLanguage retrieves the language from the context, defaults to "EN"
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(languageContextKey).(string); ok && lang != "" {
		return lang
	}
	return DefaultLanguage
}

// GetLanguageWithFallback retrieves the language from context with a custom fallback
func GetLanguageWithFallback(ctx context.Context, fallback string) string {
	if lang, ok := ctx.Value(languageContextKey).(string); ok && lang != "" {
		return lang
	}
	return fallback
}
