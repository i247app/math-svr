package controller

import (
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type MiscController struct {
	appResource *resources.AppResource
}

func NewMiscController(appResource *resources.AppResource) *MiscController {
	return &MiscController{
		appResource: appResource,
	}
}

func (c *MiscController) HandleHealthCheck() string {
	return "OK"
}

func (c *MiscController) HandleSessionDump(w http.ResponseWriter, r *http.Request) {
	dumpedSession := session.Dump(c.appResource.SessionManager)
	response.WriteJson(w, r.Context(), dumpedSession, nil, status.OK)
}
