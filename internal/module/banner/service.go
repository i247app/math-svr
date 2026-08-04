package banner

import (
	"context"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/banner"
	dto "math-ai.com/math-ai/internal/application/dto/banner"
	query "math-ai.com/math-ai/internal/application/query/banner"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/banner"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const (
	// bannerFolder is the S3 prefix new banner media uploads land under.
	bannerFolder = "banner-media"
	// bannerUrlTTL bounds how long a generated banner media preview URL is
	// valid. Short enough that a stale link in the wild expires quickly;
	// long enough that a single screen render doesn't refresh mid-view.
	bannerUrlTTL = 1 * time.Hour
)

// Service is the banner module's public façade. It composes the CQRS
// handlers behind the validators and owns no I/O of its own — every write
// goes through the UoW-backed commands, every read through the repo-bound
// queries. storageProvider may be nil; in that case responses simply omit
// media_url.
type Service struct {
	getBannerQuery       *query.GetBannerByIdQueryHandler
	listBannersQuery     *query.ListBannersQueryHandler
	createBannerCmd      *command.CreateBannerCommandHandler
	updateBannerCmd      *command.UpdateBannerCommandHandler
	softDeleteBannerCmd  *command.SoftDeleteBannerCommandHandler
	forceDeleteBannerCmd *command.ForceDeleteBannerCommandHandler
	storageProvider      *storage.Adapter
}

func NewService(
	bannerRepo domain.IRepository,
	uow transaction.UnitOfWork,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getBannerQuery:       query.NewGetBannerByIdQueryHandler(bannerRepo),
		listBannersQuery:     query.NewListBannersQueryHandler(bannerRepo),
		createBannerCmd:      command.NewCreateBannerCommandHandler(uow),
		updateBannerCmd:      command.NewUpdateBannerCommandHandler(uow),
		softDeleteBannerCmd:  command.NewSoftDeleteBannerCommandHandler(uow),
		forceDeleteBannerCmd: command.NewForceDeleteBannerCommandHandler(uow),
		storageProvider:      storageProvider,
	}
}

func (s *Service) CreateBanner(ctx context.Context, req *dto.CreateBannerReq, actorID *int64) (*dto.CreateBannerRes, error) {
	log := logger.From(ctx)

	if err := ValidateCreateBanner(ctx, req); err != nil {
		return nil, err
	}

	var (
		mediaKey       string
		uploadedOnThis bool
	)
	switch {
	case strings.TrimSpace(req.Media) != "":
		key, err := s.normalizeBannerKey(ctx, req.Media, status.BANNER_MEDIA_INVALID_REFERENCE)
		if err != nil {
			return nil, err
		}
		mediaKey = key
	case req.MediaFile != nil:
		key, err := s.uploadBannerIfPresent(ctx, req)
		if err != nil {
			return nil, err
		}
		if key != nil {
			mediaKey = *key
			uploadedOnThis = true
		}
	}

	created, err := s.createBannerCmd.Handle(ctx, command.CreateBannerCommand{
		ActorID:       actorID,
		Title:         req.Title,
		ShortText:     req.ShortText,
		MediaType:     req.MediaType,
		MediaURLKey:   mediaKey,
		ButtonText:    req.ButtonText,
		ButtonLinkURL: req.ButtonLinkURL,
		Note:          req.Note,
		BannerStatus:  req.BannerStatus,
	})
	if err != nil {
		if uploadedOnThis && mediaKey != "" && s.storageProvider != nil {
			if delErr := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: mediaKey}); delErr != nil {
				log.Warnf("banner.create banner orphan cleanup failed key=%s err=%v", mediaKey, delErr)
			}
		}
		return nil, err
	}

	resp := dto.DomainToResponse(created)
	s.populateMediaUrl(ctx, resp)
	return &dto.CreateBannerRes{Banner: resp}, nil
}

func (s *Service) UpdateBanner(ctx context.Context, req *dto.UpdateBannerReq, actorID *int64) (*dto.UpdateBannerRes, error) {
	log := logger.From(ctx)

	if err := ValidateUpdateBanner(ctx, req); err != nil {
		return nil, err
	}

	existBanner, err := s.getBannerQuery.Handle(ctx, query.GetBannerByIdQuery{BannerID: req.BannerID})
	if err != nil {
		return nil, err
	}
	if existBanner == nil {
		return nil, errs.NewError(ctx, status.BANNER_NOT_FOUND, nil, ErrBannerNotFound)
	}

	var (
		mediaKey       *string
		uploadedOnThis bool
	)
	switch {
	case strings.TrimSpace(req.Media) != "":
		key, err := s.normalizeBannerKey(ctx, req.Media, status.BANNER_MEDIA_INVALID_REFERENCE)
		if err != nil {
			return nil, err
		}
		mediaKey = &key
	case req.MediaFile != nil:
		key, err := s.updateBannerIfPresent(ctx, req)
		if err != nil {
			return nil, err
		}
		mediaKey = key
		uploadedOnThis = true
	}

	updated, err := s.updateBannerCmd.Handle(ctx, command.UpdateBannerCommand{
		ActorID:       actorID,
		BannerID:      req.BannerID,
		Title:         req.Title,
		ShortText:     req.ShortText,
		MediaType:     req.MediaType,
		MediaURLKey:   mediaKey,
		ButtonText:    req.ButtonText,
		ButtonLinkURL: req.ButtonLinkURL,
		Note:          req.Note,
		BannerStatus:  req.BannerStatus,
	})
	if err != nil {
		if uploadedOnThis && mediaKey != nil && s.storageProvider != nil {
			if delErr := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{Key: *mediaKey}); delErr != nil {
				log.Warnf("banner.update banner orphan cleanup failed key=%s err=%v", *mediaKey, delErr)
			}
		}
		return nil, err
	}

	// Delete the old media object once it has been superseded — best-effort:
	// the DB update already committed, so a failed S3 delete (or disabled
	// storage) must not fail the request, only leave an orphan we log.
	if s.storageProvider != nil && existBanner.MediaURLKey() != "" &&
		mediaKey != nil && existBanner.MediaURLKey() != *mediaKey {
		if err := s.storageProvider.HandleDelete(ctx, &storage.DeleteFileRequest{
			Key: existBanner.MediaURLKey(),
		}); err != nil {
			log.Warnf("banner.update old media cleanup failed key=%s err=%v", existBanner.MediaURLKey(), err)
		}
	}

	resp := dto.DomainToResponse(updated)
	s.populateMediaUrl(ctx, resp)
	return &dto.UpdateBannerRes{Banner: resp}, nil
}

func (s *Service) SoftDeleteBanner(ctx context.Context, req *dto.DeleteBannerReq) (*dto.DeleteBannerRes, error) {
	if err := ValidateDeleteBanner(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteBannerCmd.Handle(ctx, command.SoftDeleteBannerCommand{BannerID: req.BannerID}); err != nil {
		return nil, err
	}
	return &dto.DeleteBannerRes{}, nil
}

func (s *Service) ForceDeleteBanner(ctx context.Context, req *dto.DeleteBannerReq) (*dto.DeleteBannerRes, error) {
	if err := ValidateDeleteBanner(ctx, req); err != nil {
		return nil, err
	}
	if err := s.forceDeleteBannerCmd.Handle(ctx, command.ForceDeleteBannerCommand{BannerID: req.BannerID}); err != nil {
		return nil, err
	}
	return &dto.DeleteBannerRes{}, nil
}

func (s *Service) GetBanner(ctx context.Context, req *dto.GetBannerReq) (*dto.GetBannerRes, error) {
	if err := ValidateGetBanner(ctx, req); err != nil {
		return nil, err
	}

	found, err := s.getBannerQuery.Handle(ctx, query.GetBannerByIdQuery{BannerID: req.BannerID})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if found == nil {
		return nil, errs.NewError(ctx, status.BANNER_NOT_FOUND, nil, ErrBannerNotFound)
	}

	resp := dto.DomainToResponse(found)
	s.populateMediaUrl(ctx, resp)
	return &dto.GetBannerRes{Banner: resp}, nil
}

func (s *Service) ListBanners(ctx context.Context, req *dto.ListBannersReq) (*dto.ListBannersRes, error) {
	if err := ValidateListBanners(ctx, req); err != nil {
		return nil, err
	}

	banners, pg, err := s.listBannersQuery.Handle(ctx, query.ListBannersQuery{
		Search:       req.Search,
		MediaType:    req.MediaType,
		BannerStatus: req.BannerStatus,
		BannerIDs:    req.BannerIDs,
		Page:         req.Page,
		Limit:        req.Size,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	responses := dto.DomainListToResponse(banners)
	for _, r := range responses {
		s.populateMediaUrl(ctx, r)
	}
	return &dto.ListBannersRes{
		Banners:    responses,
		Pagination: pg,
	}, nil
}

// populateMediaUrl mutates resp in place to add a short-lived presigned URL
// when the banner carries a media_url_key. No-op if storage is disabled or
// the banner has no media (e.g. a TEXT banner).
func (s *Service) populateMediaUrl(ctx context.Context, resp *dto.BannerResponse) {
	if resp == nil || s.storageProvider == nil || resp.MediaURLKey == "" {
		return
	}
	url, err := s.storageProvider.CreatePresignedUrl(ctx, &storage.CreatePresignedUrlRequest{
		Key:        resp.MediaURLKey,
		Expiration: bannerUrlTTL,
	})
	if err != nil {
		logger.From(ctx).Warnf("banner.media presign failed banner_id=%d err=%v", resp.BannerID, err)
		return
	}
	resp.MediaUrl = &url
}
