package notification

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/notification"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	appResource *resource.Resource
	svc         *Service
}

func NewHandler(appResource *resource.Resource, svc *Service) *Handler {
	return &Handler{appResource: appResource, svc: svc}
}

// uidFromSession extracts the authenticated user id, or writes an
// UNAUTHORIZED response and reports ok=false.
func (h *Handler) uidFromSession(w http.ResponseWriter, r *http.Request) (int64, bool) {
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return 0, false
	}
	uid, ok := session.UID()
	if !ok {
		response.WriteJson(w, nil, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession))
		return 0, false
	}
	return uid, true
}

// POST /notifications/ping
func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	var req dto.PingNotificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.svc.Ping(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /notifications/send — create + push a notification to a recipient.
func (h *Handler) HandleSend(w http.ResponseWriter, r *http.Request) {
	var req dto.SendNotificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.CreatorUID = &uid

	res, err := h.svc.SendNotification(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /notifications/list — the caller's notifications, most-recent first.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	var req dto.ListNotificationsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = &uid

	res, err := h.svc.ListNotifications(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /notifications/unread-count — the caller's unread count.
func (h *Handler) HandleUnreadCount(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	res, err := h.svc.UnreadCount(r.Context(), &dto.UnreadCountReq{UserID: &uid})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /notifications/mark-read — mark one owned notification read.
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkReadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = &uid

	if err := h.svc.MarkRead(r.Context(), &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJsonNoContent(w)
}

// POST /notifications/mark-all-read — mark every unread notification read.
func (h *Handler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	if err := h.svc.MarkAllRead(r.Context(), &dto.MarkAllReadReq{UserID: &uid}); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJsonNoContent(w)
}

// POST /notifications/soft-delete — soft-delete an owned notification.
func (h *Handler) HandleSoftDelete(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteNotificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = &uid

	if err := h.svc.SoftDelete(r.Context(), &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJsonNoContent(w)
}
