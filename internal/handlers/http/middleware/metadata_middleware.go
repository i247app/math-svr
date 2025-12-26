package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	"math-ai.com/math-ai/internal/shared/constant/status"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
)

func MetadataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "cannot read body", status.FAIL)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(rawBody))

			if len(bytes.TrimSpace(rawBody)) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			var env dto.RequestEnvelope
			if err := json.Unmarshal(rawBody, &env); err != nil {
				http.Error(w, "invalid JSON body", status.FAIL)
				return
			}

			if env.Metadata != nil {
				if err := validators.MetadataRequestValidator(*env.Metadata); err != nil {
					http.Error(w, err.Error(), status.FAIL)
					return
				}

				ctx := appctx.SetMetadata(r.Context(), *env.Metadata)
				next.ServeHTTP(w, r.WithContext(ctx))
			}

		})
	}
}
