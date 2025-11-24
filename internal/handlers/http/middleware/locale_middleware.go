package middleware

import (
	"net/http"

	appctx "math-ai.com/math-ai/internal/shared/utils/context"
)

// LocaleMiddleware sets the locale in the request context
// It uses the Accept-Language header to determine the locale
// If the Accept-Language header is not set, it uses the default locale

func LocaleMiddleware(defaultLocale string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			locale := r.Header.Get("Accept-Language")
			if locale == "" {
				locale = defaultLocale
			}

			ctx := appctx.SetLocale(r.Context(), locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
