package device

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/device"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

type DeviceHandler struct {
	deviceSvc *Service
}

func NewDeviceHandler(deviceSvc *Service) *DeviceHandler {
	return &DeviceHandler{deviceSvc: deviceSvc}
}

// GET /devices/{id}
func (h *DeviceHandler) HandleGetDeviceById(w http.ResponseWriter, r *http.Request) {
	deviceID, err := utils.StringToUUID(r.PathValue("id"))
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.deviceSvc.GetDeviceById(r.Context(), &dto.GetDeviceByIdReq{DeviceID: deviceID})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /devices/list
func (h *DeviceHandler) HandleListDevices(w http.ResponseWriter, r *http.Request) {
	var req dto.ListDevicesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.deviceSvc.ListDevicesByUserId(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /devices/update
func (h *DeviceHandler) HandleUpdateDevice(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateDeviceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.deviceSvc.UpdateDevice(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /devices/revoke
func (h *DeviceHandler) HandleRevokeDevice(w http.ResponseWriter, r *http.Request) {
	var req dto.RevokeDeviceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.deviceSvc.RevokeDevice(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /devices/soft-delete
func (h *DeviceHandler) HandleSoftDeleteDevice(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteDeviceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.deviceSvc.SoftDeleteDevice(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
