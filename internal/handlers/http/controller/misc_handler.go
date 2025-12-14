package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type MiscController struct {
	appResource *resources.AppResource
	service     di.IMiscService
}

func NewMiscController(appResource *resources.AppResource, service di.IMiscService) *MiscController {
	return &MiscController{
		appResource: appResource,
		service:     service,
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

func (c *MiscController) HandleDetermineLocation(w http.ResponseWriter, r *http.Request) {
	var req dto.LocationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.FAIL)
		return
	}

	statusCode, locationRes, err := c.service.DetermineLocation(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), locationRes, nil, statusCode)
}
