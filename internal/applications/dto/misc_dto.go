package dto

type HealthCheckResponse struct {
	ServerPing   string `json:"server_ping"`
	DatabasePing string `json:"database_ping"`
}
