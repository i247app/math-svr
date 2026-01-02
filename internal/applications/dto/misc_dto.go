package dto

import "math-ai.com/math-ai/internal/shared/constant/enum"

type HealthCheckResponse struct {
	ServerPing   string `json:"server_ping"`
	DatabasePing string `json:"database_ping"`
}

type LocationResponse struct {
	City        string `json:"city"`
	State       string `json:"state"`
	StateCode   string `json:"state_code"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

type LocationDetectRequest struct {
	TypeDetect enum.EGeoCategory `json:"type_detect"`

	LatLng struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"lat_lng"`

	IpAddress string `json:"ip_address"`
}

type LocationDetectResponse struct {
	Location *LocationResponse `json:"location"`
}
