package appctx

import (
	"context"
)

type localeKeyType string

const (
	localeKey = localeKeyType("locale")
)

func SetLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey, locale)
}

func GetLocale(ctx context.Context) string {
	locale := ctx.Value(localeKey)
	if locale == nil {
		return "en"
	}

	return locale.(string)
}
