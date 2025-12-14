package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/driven-adapter/provider/geo"
	"math-ai.com/math-ai/internal/driven-adapter/provider/ip"
	"math-ai.com/math-ai/internal/shared/constant/status"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
)

type MiscService struct {
	geoSvc *geo.GeoFencingService
	ipSvc  *ip.IPGeolocationService
}

func NewMiscService(geoSvc *geo.GeoFencingService, ipSvc *ip.IPGeolocationService) *MiscService {
	return &MiscService{
		geoSvc: geoSvc,
		ipSvc:  ipSvc,
	}
}

func (s *MiscService) Ping() (status.Code, error) {
	return status.SUCCESS, nil
}

func (s *MiscService) DetermineLocation(ctx context.Context, req *dto.LocationRequest) (status.Code, *dto.LocationResponse, error) {
	language := appctx.GetLocale(ctx)

	resp := &dto.LocationResponse{}

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

		resp.City = result.City
		resp.State = result.State
		resp.StateCode = result.StateCode
		resp.Country = result.Country
		resp.CountryCode = result.CountryCode
	case "ip_address":
		reqIP := &ip.IPLookupRequest{
			IP:       req.IpAddress,
			Language: language,
		}

		result, err := s.ipSvc.GetLocationByIP(ctx, reqIP)
		if err != nil {
			return status.FAIL, nil, err
		}

		resp.City = result.City
		resp.State = result.RegionName
		resp.StateCode = result.Region
		resp.Country = result.Country
		resp.CountryCode = result.CountryCode
	default:
		return status.FAIL, nil, nil
	}

	return status.SUCCESS, resp, nil
}
