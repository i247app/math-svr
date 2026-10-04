package device

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/device"
	"math-ai.com/math-ai/internal/shared/response"
)

type DeviceHandler struct {
	service *Service
}

func NewDeviceHandler(service *Service) *DeviceHandler {
	return &DeviceHandler{service: service}
}

// POST /devices/detail
func (h *DeviceHandler) HandleGetDeviceById(w http.ResponseWriter, r *http.Request) {
	var req dto.GetDeviceByIdReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.GetDeviceById(r.Context(), &req)
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

	res, err := h.service.ListDevicesByUid(r.Context(), &req)
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

	res, err := h.service.UpdateDevice(r.Context(), &req)
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

	res, err := h.service.RevokeDevice(r.Context(), &req)
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

	res, err := h.service.SoftDeleteDevice(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /devices/force-delete
func (h *DeviceHandler) HandleForceDeleteDevice(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteDeviceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.ForceDeleteDevice(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
