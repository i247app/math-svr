package appctx

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
)

type metadataKeyType string

const (
	metadataKey = metadataKeyType("metadata")
)

func SetMetadata(ctx context.Context, metadata dto.MetadataRequest) context.Context {
	return context.WithValue(ctx, metadataKey, metadata)
}

func GetMetadata(ctx context.Context) (dto.MetadataRequest, bool) {
	metadataStr := ctx.Value(metadataKey)
	meta, ok := metadataStr.(dto.MetadataRequest)
	return meta, ok
}
