package session

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/session"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	appResource *resource.Resource
	service     *Service
}

func NewHandler(appResource *resource.Resource) *Handler {
	return &Handler{appResource: appResource, service: NewService(appResource.SessionManager)}
}

func (c *Handler) HandleSessionDump(w http.ResponseWriter, r *http.Request) {
	dumpedSession := session.Dump(c.appResource.SessionManager)

	response.WriteJson(w, dumpedSession, nil)
}

func (c *Handler) HandleDeleteUnSecureSessions(w http.ResponseWriter, r *http.Request) {
	c.appResource.SessionManager.DeleteUnSecureSessions()

	response.WriteJson(w, "Delete UnSecure Sessions", nil)
}

func (c *Handler) HandleDeleteAllSessions(w http.ResponseWriter, r *http.Request) {
	c.appResource.SessionManager.DeleteAll()

	response.WriteJson(w, "Delete All Sessions", nil)
}

// POST /sessions/delete — signs out the one session stored under the token.
func (c *Handler) HandleDeleteSession(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSessionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := c.service.DeleteByToken(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
