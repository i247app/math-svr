package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/driven-adapter/provider/geo"
	"math-ai.com/math-ai/internal/shared/constant/status"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
)

type MiscService struct {
	geoSvc *geo.GeoFencingService
}

func NewMiscService(geoSvc *geo.GeoFencingService) *MiscService {
	return &MiscService{
		geoSvc: geoSvc,
	}
}

func (s *MiscService) Ping() (status.Code, error) {
	return status.SUCCESS, nil
}

func (s *MiscService) DetermineLocation(ctx context.Context, req *dto.LocationRequest) (status.Code, *dto.LocationResponse, error) {
	language := appctx.GetLocale(ctx)

	var location *geo.LocationInfo
	switch req.TypeDetect {
	case "lat_lng":
		reqGeo := &geo.ReverseGeocodeRequest{
			Language: language,
			Lat:      req.LatLng.Lat,
			Lng:      req.LatLng.Lng,
		}

		result, err := s.geoSvc.ReverseGeocode(ctx, reqGeo)
		if err != nil {
			return status.FAIL, nil, err
		}

		location = result
	case "ip_address":
	default:
		return status.FAIL, nil, nil
	}

	resp := &dto.LocationResponse{
		City:    location.City,
		State:   location.State,
		Country: location.Country,
	}

	return status.SUCCESS, resp, nil
}
