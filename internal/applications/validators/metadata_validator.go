package validators

import (
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
)

func MetadataRequestValidator(req dto.MetadataRequest) error {
	if req.IpAddress == "" {
		return fmt.Errorf("metadata.ip_address is required")
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		return fmt.Errorf("metadata.latitude is invalid")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return fmt.Errorf("metadata.longitude is invalid")
	}
	return nil
}
