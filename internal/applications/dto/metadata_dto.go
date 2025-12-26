package dto

type MetadataRequest struct {
	IpAddress string  `json:"ip_address"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type RequestEnvelope struct {
	Metadata *MetadataRequest `json:"metadata,omitempty"`
}
