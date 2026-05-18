package health

import (
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/health"
	"math-ai.com/math-ai/internal/shared/response"
)

type HealthHandler struct {
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// POST /ping
func (h *HealthHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	var res dto.PingRes
	res.Message = "Pong"
	response.WriteJson(w, res, nil)
}
