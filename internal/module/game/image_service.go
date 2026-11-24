package game

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/provider/storage"
)

// imageURLTTL is how long a presigned image URL stays valid.
const imageURLTTL = time.Hour

// urlSigner is the part of the storage provider this module needs.
type urlSigner interface {
	CreatePresignedUrl(ctx context.Context, req *storage.CreatePresignedUrlRequest) (string, error)
}

// signImages fills the image URLs of a served round, in place. A bank
// question may carry image keys; with no storage, or on a signing failure,
// the URL is left empty and the request still succeeds.
func (s *Service) signImages(ctx context.Context, res *dto.GameSessionRes) {
	if s.signer == nil || res.GameSession == nil {
		return
	}
	for i := range res.GameSession.Questions {
		q := &res.GameSession.Questions[i]
		q.QuestionImageURL = s.imageURL(ctx, q.QuestionImageKey)
		for j := range q.Options {
			q.Options[j].ImageURL = s.imageURL(ctx, q.Options[j].ImageKey)
		}
	}
}

func (s *Service) imageURL(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	url, err := s.signer.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{Key: key, Expiration: imageURLTTL})
	if err != nil {
		logger.From(ctx).Warnf("game.image.presign_failed key=%s err=%v", key, err)
		return ""
	}
	return url
}
