package user

import (
	"context"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	profileDto "math-ai.com/math-ai/internal/application/dto/profile"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const imageUrlTTL = 1 * time.Hour

func (s *Service) populateImageUrl(ctx context.Context, resp *profileDto.ProfileResponse) {
	if resp == nil || s.storageProvider == nil || resp.AvatarKey == nil || *resp.AvatarKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        *resp.AvatarKey,
		Expiration: imageUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("user.avatar presign failed user_id=%s err=%v", resp.UserID, err)
		return
	}
	resp.AvatarUrl = &url
}
