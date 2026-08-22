package chat

import (
	"encoding/json"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/chat"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	appResource *resource.Resource
	chatSvc     *Service
}

func NewHandler(appResource *resource.Resource, chatSvc *Service) *Handler {
	return &Handler{appResource: appResource, chatSvc: chatSvc}
}

// sessionUserID pulls the authenticated user off the session. Every chat route
// is auth-gated, so a zero here means the session was unreadable; the service
// treats 0 as "no ownership proof" and the profile check then fails closed.
func (h *Handler) sessionUserID(r *http.Request) int64 {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil || sess == nil {
		return 0
	}
	uid, ok := sess.UID()
	if !ok {
		return 0
	}
	return uid
}

func decode(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("invalid parameters")
	}
	return nil
}

// POST /chats/classroom-members
func (h *Handler) HandleListClassroomMembers(w http.ResponseWriter, r *http.Request) {
	var req dto.ListClassroomMembersReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.ListClassroomMembers(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}

// POST /chats/conversations/open
func (h *Handler) HandleOpenConversation(w http.ResponseWriter, r *http.Request) {
	var req dto.OpenConversationReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.OpenConversation(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}

// POST /chats/conversations/list
func (h *Handler) HandleListConversations(w http.ResponseWriter, r *http.Request) {
	var req dto.ListConversationsReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.ListConversations(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}

// POST /chats/messages/list
func (h *Handler) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	var req dto.ListMessagesReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.ListMessages(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}

// POST /chats/messages/send
func (h *Handler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	var req dto.SendMessageReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.SendMessage(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}

// POST /chats/messages/mark-read
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkReadReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.MarkRead(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}

// POST /chats/unread-count
func (h *Handler) HandleUnreadCount(w http.ResponseWriter, r *http.Request) {
	var req dto.UnreadCountReq
	if err := decode(r, &req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.chatSvc.UnreadCount(r.Context(), &req, h.sessionUserID(r))
	response.WriteJson(w, res, err)
}
