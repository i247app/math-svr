package llmroute

import (
	"context"

	"math-ai.com/math-ai/internal/domain/shared/status"
	sctx "math-ai.com/math-ai/internal/shared/context"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// requireAdminGate fails closed when a route was registered without the
// admin middleware: a route decides which vendor every child's exam goes to.
func requireAdminGate(ctx context.Context) (sctx.AdminVia, error) {
	via := sctx.GetAdminVia(ctx)
	if via == "" {
		return "", errs.NewError(ctx, status.FORBIDDEN, nil, ErrAdminGateMissing)
	}
	return via, nil
}

// adminActor is the admin's uid for modify_id, or nil when the request came
// in with the X-Api-Admin-Key header and has no account behind it.
func adminActor(ctx context.Context) *int64 {
	if uid := sctx.UID(ctx); uid > 0 {
		return &uid
	}
	return nil
}
