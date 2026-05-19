package session

import (
	"net/http"

	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	appResource *resource.Resource
}

func NewHandler(appResource *resource.Resource) *Handler {
	return &Handler{appResource: appResource}
}

func (c *Handler) HandleSessionDump(w http.ResponseWriter, r *http.Request) {
	dumpedSession := session.Dump(c.appResource.SessionManager)

	response.WriteJson(w, dumpedSession, nil)
}

func (c *Handler) HandleDeleteUnSecureSessions(w http.ResponseWriter, r *http.Request) {
	c.appResource.SessionManager.DeleteUnSecureSessions()

	response.WriteJson(w, "Delete UnSecure Sessions", nil)
}
