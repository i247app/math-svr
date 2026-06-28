package conversation

import (
	"encoding/json"
	"net/http"
	"strconv"

	dto "math-ai.com/math-ai/internal/application/dto/conversation"
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

// POST /ai/conversations/send
func (h *Handler) HandleSend(w http.ResponseWriter, r *http.Request) {
	var req dto.SendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = &uid

	res, err := h.svc.SendMessage(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /ai/conversations/list
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	var req dto.ListConversationsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = &uid

	res, err := h.svc.ListConversations(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /ai/conversations/{id}
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	res, err := h.svc.GetConversationById(r.Context(), &dto.GetConversationByIdReq{
		ConversationID: id,
		UserID:         &uid,
	})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /ai/conversations/soft-delete
func (h *Handler) HandleSoftDelete(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteConversationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := h.uidFromSession(w, r)
	if !ok {
		return
	}
	req.UserID = &uid

	if _, err := h.svc.SoftDeleteConversation(r.Context(), &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJsonNoContent(w)
}
