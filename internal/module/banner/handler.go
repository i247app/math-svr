package banner

import (
	"encoding/json"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/banner"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/response"
)

const (
	MaxImageUploadSize = 10 << 20 // 10 MB
)

type BannerHandler struct {
	appResource *resource.Resource
	bannerSvc   *Service
}

func NewBannerHandler(appResource *resource.Resource, bannerSvc *Service) *BannerHandler {
	return &BannerHandler{
		appResource: appResource,
		bannerSvc:   bannerSvc,
	}
}

// actorID pulls the authenticated user's id off the session so create/update
// can stamp create_id/modify_id. Returns nil when no session is bound (the
// route is auth-gated, so this is defensive only).
func (h *BannerHandler) actorID(r *http.Request) *int64 {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil || sess == nil {
		return nil
	}
	uid, ok := sess.UID()
	if !ok {
		return nil
	}
	return &uid
}

// POST /banners/create
func (h *BannerHandler) HandleCreateBanner(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBannerReq
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid parameters"))
			return
		}
	} else {
		if err := r.ParseMultipartForm(MaxImageUploadSize); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid form data"))
			return
		}

		title := r.FormValue("title")
		req.Title = &title

		shortText := r.FormValue("short_text")
		req.ShortText = &shortText

		req.MediaType = r.FormValue("media_type")

		buttonText := r.FormValue("button_text")
		req.ButtonText = &buttonText

		buttonLink := r.FormValue("button_link_url")
		req.ButtonLinkURL = &buttonLink

		note := r.FormValue("note")
		req.Note = &note

		// Only set banner_status when supplied; an empty value would fail
		// validation instead of falling through to the ACTIVE default.
		if bannerStatus := r.FormValue("banner_status"); bannerStatus != "" {
			req.BannerStatus = &bannerStatus
		}

		// Handle media file
		file, header, err := r.FormFile("media")
		if err == nil {
			defer file.Close()
			req.MediaFile = file
			req.MediaFilename = header.Filename
			req.MediaContentType = header.Header.Get("Content-Type")
		}
	}

	res, err := h.bannerSvc.CreateBanner(r.Context(), &req, h.actorID(r))
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /banners/update
func (h *BannerHandler) HandleUpdateBanner(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateBannerReq
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid parameters"))
			return
		}
	} else {
		if err := r.ParseMultipartForm(MaxImageUploadSize); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid form data"))
			return
		}

		title := r.FormValue("title")
		req.Title = &title

		shortText := r.FormValue("short_text")
		req.ShortText = &shortText

		mediaType := r.FormValue("media_type")
		req.MediaType = &mediaType

		buttonText := r.FormValue("button_text")
		req.ButtonText = &buttonText

		buttonLink := r.FormValue("button_link_url")
		req.ButtonLinkURL = &buttonLink

		note := r.FormValue("note")
		req.Note = &note

		// Only set banner_status when supplied; an empty value would fail
		// validation instead of leaving the field unchanged.
		if bannerStatus := r.FormValue("banner_status"); bannerStatus != "" {
			req.BannerStatus = &bannerStatus
		}

		// Handle media file
		file, header, err := r.FormFile("media")
		if err == nil {
			defer file.Close()
			req.MediaFile = file
			req.MediaFilename = header.Filename
			req.MediaContentType = header.Header.Get("Content-Type")
		}
	}
	res, err := h.bannerSvc.UpdateBanner(r.Context(), &req, h.actorID(r))
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /banners/soft-delete
func (h *BannerHandler) HandleSoftDeleteBanner(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteBannerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.bannerSvc.SoftDeleteBanner(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /banners/force-delete
func (h *BannerHandler) HandleForceDeleteBanner(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteBannerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.bannerSvc.ForceDeleteBanner(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /banners/detail
func (h *BannerHandler) HandleGetBanner(w http.ResponseWriter, r *http.Request) {
	var req dto.GetBannerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.bannerSvc.GetBanner(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /banners/list — also handles search via the optional `search` filter.
func (h *BannerHandler) HandleListBanners(w http.ResponseWriter, r *http.Request) {
	var req dto.ListBannersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.bannerSvc.ListBanners(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
