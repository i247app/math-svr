package controller

import (
	"net/http"
	"time"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
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

func (c *MiscController) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	var res dto.HealthCheckResponse

	res.ServerPing = "Go live " + time.Now().Format(time.DateTime)

	err := c.appResource.Db.PingContext(r.Context())
	if err != nil {
		res.DatabasePing = "can not connect database: " + err.Error()
	} else {
		res.DatabasePing = "Database live " + time.Now().Format(time.DateTime)
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}

func (c *MiscController) HandleSessionDump(w http.ResponseWriter, r *http.Request) {
	dumpedSession := session.Dump(c.appResource.SessionManager)
	response.WriteJson(w, r.Context(), dumpedSession, nil, status.OK)
}
