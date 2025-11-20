package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type LoginController struct {
	appResources *resources.AppResource
	service      di.ILoginService
}

func NewLoginController(appResources *resources.AppResource, service di.ILoginService) *LoginController {
	return &LoginController{
		appResources: appResources,
		service:      service,
	}
}

// POST - /login
func (c *LoginController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}
	defer r.Body.Close()
	req.IpAddress = r.RemoteAddr
	req.DeviceUUID = r.Header.Get("Device-UUID")
	req.DeviceName = r.Header.Get("Device-Name")

	ctx := r.Context()

	// sess, err := c.appResources.GetRequestSession(r)
	// if err != nil {
	// 	response.WriteJson(w, ctx, nil, fmt.Errorf("failed to get session %w", err), status.INTERNAL)
	// 	return
	// }

	// Perform login
	statusCode, res, err := c.service.Login(ctx, nil, &req)
	if err != nil {
		response.WriteJson(w, ctx, nil, err, statusCode)
		return
	}

	response.WriteJson(w, ctx, res, nil, statusCode)
}
