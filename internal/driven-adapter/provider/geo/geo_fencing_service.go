package geo

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/driven-adapter/external/http_client"
	"math-ai.com/math-ai/internal/shared/utils/convert"
)

// GeoFencingService handles geo-location and geo-fencing operations using Google Maps API
type GeoFencingService struct {
	client *http_client.Client
	apiKey string
}

func NewGeoFencingService(apiKey string, opts ...http_client.Option) *GeoFencingService {
	// Add default options for geo-fencing service
	defaultOpts := []http_client.Option{
		http_client.WithBaseURL("https://maps.googleapis.com/maps/api"),
		http_client.WithContentType("application/json"),
		http_client.WithAccept("application/json"),
		http_client.WithUserAgent("Math-AI-GeoFencing-Client/1.0"),
		http_client.WithTimeout(30 * time.Second),
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &GeoFencingService{
		client: http_client.NewClient(allOpts...),
		apiKey: apiKey,
	}
}

// ReverseGeocode converts latitude/longitude to address information
func (s *GeoFencingService) ReverseGeocode(ctx context.Context, req *ReverseGeocodeRequest) (*LocationInfo, error) {
	if req.Lat < -90 || req.Lat > 90 {
		return nil, fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", req.Lat)
	}
	if req.Lng < -180 || req.Lng > 180 {
		return nil, fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", req.Lng)
	}

	// Build request options
	reqOpts := []http_client.RequestOption{
		http_client.WithRequestQueryParam("latlng", fmt.Sprintf("%f,%f", req.Lat, req.Lng)),
		http_client.WithRequestQueryParam("key", s.apiKey),
	}

	if len(req.ResultType) > 0 {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("result_type", strings.Join(req.ResultType, "|")))
	}

	if len(req.LocationType) > 0 {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("location_type", strings.Join(req.LocationType, "|")))
	}

	if req.Language != "" {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("language", req.Language))
	}

	resp, err := s.client.Get(ctx, "/geocode/json", reqOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse geocode: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("geocoding API error: %s", resp.String())
	}

	var geocodeResp GeocodingResponse
	if err := resp.JSON(&geocodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if geocodeResp.Status != "OK" {
		return nil, fmt.Errorf("geocoding failed: %s - %s", geocodeResp.Status, geocodeResp.ErrorMessage)
	}

	if len(geocodeResp.Results) == 0 {
		return nil, fmt.Errorf("no results found for coordinates: %f,%f", req.Lat, req.Lng)
	}

	// Parse the first result into LocationInfo
	return s.parseLocationInfo(&geocodeResp.Results[0]), nil
}

// ForwardGeocode converts address to latitude/longitude
func (s *GeoFencingService) ForwardGeocode(ctx context.Context, req *ForwardGeocodeRequest) (*LocationInfo, error) {
	if req.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	// Build request options
	reqOpts := []http_client.RequestOption{
		http_client.WithRequestQueryParam("address", req.Address),
		http_client.WithRequestQueryParam("key", s.apiKey),
	}

	if len(req.Components) > 0 {
		var components []string
		for k, v := range req.Components {
			components = append(components, fmt.Sprintf("%s:%s", k, v))
		}
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("components", strings.Join(components, "|")))
	}

	if req.Bounds != nil {
		bounds := fmt.Sprintf("%f,%f|%f,%f",
			req.Bounds.Southwest.Lat, req.Bounds.Southwest.Lng,
			req.Bounds.Northeast.Lat, req.Bounds.Northeast.Lng)
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("bounds", bounds))
	}

	if req.Language != "" {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("language", req.Language))
	}

	if req.Region != "" {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("region", req.Region))
	}

	resp, err := s.client.Get(ctx, "/geocode/json", reqOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to forward geocode: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("geocoding API error: %s", resp.String())
	}

	var geocodeResp GeocodingResponse
	if err := resp.JSON(&geocodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if geocodeResp.Status != "OK" {
		return nil, fmt.Errorf("geocoding failed: %s - %s", geocodeResp.Status, geocodeResp.ErrorMessage)
	}

	if len(geocodeResp.Results) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", req.Address)
	}

	return s.parseLocationInfo(&geocodeResp.Results[0]), nil
}

// IsLocationInGeoFence checks if a location is within a geo-fence
func (s *GeoFencingService) IsLocationInGeoFence(location LatLng, geoFence *GeoFence) bool {
	if geoFence.RadiusKM > 0 {
		// Circle-based geo-fence
		distance := s.CalculateDistance(location, geoFence.Center)
		return distance <= geoFence.RadiusKM
	}

	if len(geoFence.Polygon) >= 3 {
		// Polygon-based geo-fence
		return s.isPointInPolygon(location, geoFence.Polygon)
	}

	return false
}

// CheckLocationInGeoFence reverse geocodes a location and checks if it's in the geo-fence
func (s *GeoFencingService) CheckLocationInGeoFence(ctx context.Context, lat, lng float64, geoFence *GeoFence) (*LocationInfo, bool, error) {
	// Get location info
	locationInfo, err := s.ReverseGeocode(ctx, &ReverseGeocodeRequest{
		Lat: lat,
		Lng: lng,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to get location info: %w", err)
	}

	// Check if in geo-fence
	isInFence := s.IsLocationInGeoFence(LatLng{Lat: lat, Lng: lng}, geoFence)

	return locationInfo, isInFence, nil
}

// CalculateDistance calculates the distance between two points using Haversine formula (in kilometers)
func (s *GeoFencingService) CalculateDistance(point1, point2 LatLng) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := convert.DegreesToRadians(point1.Lat)
	lat2Rad := convert.DegreesToRadians(point2.Lat)
	deltaLat := convert.DegreesToRadians(point2.Lat - point1.Lat)
	deltaLng := convert.DegreesToRadians(point2.Lng - point1.Lng)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// ValidateLocation validates if coordinates are valid
func (s *GeoFencingService) ValidateLocation(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", lng)
	}
	return nil
}

// GetLocationByPlaceID gets location information by Google Place ID
func (s *GeoFencingService) GetLocationByPlaceID(ctx context.Context, placeID string) (*LocationInfo, error) {
	if placeID == "" {
		return nil, fmt.Errorf("place ID is required")
	}

	reqOpts := []http_client.RequestOption{
		http_client.WithRequestQueryParam("place_id", placeID),
		http_client.WithRequestQueryParam("key", s.apiKey),
	}

	resp, err := s.client.Get(ctx, "/geocode/json", reqOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get location by place ID: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("geocoding API error: %s", resp.String())
	}

	var geocodeResp GeocodingResponse
	if err := resp.JSON(&geocodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	if geocodeResp.Status != "OK" {
		return nil, fmt.Errorf("geocoding failed: %s - %s", geocodeResp.Status, geocodeResp.ErrorMessage)
	}

	if len(geocodeResp.Results) == 0 {
		return nil, fmt.Errorf("no results found for place ID: %s", placeID)
	}

	return s.parseLocationInfo(&geocodeResp.Results[0]), nil
}

// parseLocationInfo extracts location information from geocoding result
func (s *GeoFencingService) parseLocationInfo(result *GeocodingResult) *LocationInfo {
	info := &LocationInfo{
		FormattedAddress: result.FormattedAddress,
		Location:         result.Geometry.Location,
		PlaceID:          result.PlaceID,
	}

	// Extract address components
	for _, component := range result.AddressComponents {
		for _, componentType := range component.Types {
			switch componentType {
			case "locality":
				info.City = component.LongName
			case "administrative_area_level_2":
				info.City = component.LongName
			case "administrative_area_level_1":
				info.State = component.LongName
			case "country":
				info.Country = component.LongName
				info.CountryCode = component.ShortName
			case "postal_code":
				info.PostalCode = component.LongName
			case "street_address", "route":
				if info.Address == "" {
					info.Address = component.LongName
				}
			}
		}
	}

	return info
}

// isPointInPolygon checks if a point is inside a polygon using ray casting algorithm
func (s *GeoFencingService) isPointInPolygon(point LatLng, polygon []LatLng) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		xi, yi := polygon[i].Lng, polygon[i].Lat
		xj, yj := polygon[j].Lng, polygon[j].Lat

		intersect := ((yi > point.Lat) != (yj > point.Lat)) &&
			(point.Lng < (xj-xi)*(point.Lat-yi)/(yj-yi)+xi)

		if intersect {
			inside = !inside
		}

		j = i
	}

	return inside
}

// GetDistanceBetweenAddresses calculates distance between two addresses
func (s *GeoFencingService) GetDistanceBetweenAddresses(ctx context.Context, address1, address2 string) (float64, error) {
	// Geocode first address
	loc1, err := s.ForwardGeocode(ctx, &ForwardGeocodeRequest{Address: address1})
	if err != nil {
		return 0, fmt.Errorf("failed to geocode address1: %w", err)
	}

	// Geocode second address
	loc2, err := s.ForwardGeocode(ctx, &ForwardGeocodeRequest{Address: address2})
	if err != nil {
		return 0, fmt.Errorf("failed to geocode address2: %w", err)
	}

	// Calculate distance
	return s.CalculateDistance(loc1.Location, loc2.Location), nil
}

// BatchReverseGeocode reverse geocodes multiple locations
func (s *GeoFencingService) BatchReverseGeocode(ctx context.Context, locations []LatLng) ([]*LocationInfo, error) {
	results := make([]*LocationInfo, 0, len(locations))

	for _, loc := range locations {
		info, err := s.ReverseGeocode(ctx, &ReverseGeocodeRequest{
			Lat: loc.Lat,
			Lng: loc.Lng,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to reverse geocode %v: %w", loc, err)
		}
		results = append(results, info)
	}

	return results, nil
}

// FindNearestLocation finds the nearest location from a list
func (s *GeoFencingService) FindNearestLocation(target LatLng, candidates []LatLng) (LatLng, float64, int) {
	if len(candidates) == 0 {
		return LatLng{}, 0, -1
	}

	minDistance := math.MaxFloat64
	nearestIndex := 0
	var nearest LatLng

	for i, candidate := range candidates {
		distance := s.CalculateDistance(target, candidate)
		if distance < minDistance {
			minDistance = distance
			nearestIndex = i
			nearest = candidate
		}
	}

	return nearest, minDistance, nearestIndex
}
