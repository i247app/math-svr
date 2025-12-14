package ip

import (
	"context"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/driven-adapter/external/http_client"
	"math-ai.com/math-ai/internal/shared/utils/convert"
)

// IPGeolocationService handles IP-based geolocation operations using ip-api.com
type IPGeolocationService struct {
	client *http_client.Client
}

// NewIPGeolocationService creates a new IP geolocation service with configurable options
// Example usage:
//
//	service := NewIPGeolocationService(
//	  WithTimeout(30 * time.Second),
//	)
func NewIPGeolocationService(opts ...http_client.Option) *IPGeolocationService {
	// Add default options for IP geolocation service
	defaultOpts := []http_client.Option{
		http_client.WithBaseURL("http://ip-api.com"),
		http_client.WithContentType("application/json"),
		http_client.WithAccept("application/json"),
		http_client.WithUserAgent("Math-AI-IP-Geolocation-Client/1.0"),
		http_client.WithTimeout(30 * time.Second),
	}

	// Merge default options with provided options
	allOpts := append(defaultOpts, opts...)

	return &IPGeolocationService{
		client: http_client.NewClient(allOpts...),
	}
}

// GetLocationByIP gets location information for a specific IP address
func (s *IPGeolocationService) GetLocationByIP(ctx context.Context, req *IPLookupRequest) (*IPLocationInfo, error) {
	if req.IP == "" {
		return nil, fmt.Errorf("IP address is required")
	}

	if err := s.ValidateIP(req.IP); err != nil {
		return nil, err
	}

	// Build request options
	var reqOpts []http_client.RequestOption

	// Add fields parameter if specified
	if len(req.Fields) > 0 {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("fields", strings.Join(req.Fields, ",")))
	}

	// Add language parameter if specified
	if req.Language != "" {
		reqOpts = append(reqOpts, http_client.WithRequestQueryParam("lang", req.Language))
	}

	// Make request to /json/{ip}
	resp, err := s.client.Get(ctx, fmt.Sprintf("/json/%s", req.IP), reqOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP location: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("IP geolocation API error: %s", resp.String())
	}

	var locationInfo IPLocationInfo
	if err := resp.JSON(&locationInfo); err != nil {
		return nil, fmt.Errorf("failed to parse IP location response: %w", err)
	}

	if locationInfo.Status == "fail" {
		return nil, fmt.Errorf("IP lookup failed: %s", locationInfo.Message)
	}

	return &locationInfo, nil
}

// GetLocationByCurrentIP gets location information for the requester's IP (using empty IP)
func (s *IPGeolocationService) GetLocationByCurrentIP(ctx context.Context) (*IPLocationInfo, error) {
	// When no IP is specified, ip-api.com returns info for the requester's IP
	resp, err := s.client.Get(ctx, "/json/")
	if err != nil {
		return nil, fmt.Errorf("failed to get current IP location: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("IP geolocation API error: %s", resp.String())
	}

	var locationInfo IPLocationInfo
	if err := resp.JSON(&locationInfo); err != nil {
		return nil, fmt.Errorf("failed to parse IP location response: %w", err)
	}

	if locationInfo.Status == "fail" {
		return nil, fmt.Errorf("IP lookup failed: %s", locationInfo.Message)
	}

	return &locationInfo, nil
}

// BatchGetLocationByIPs gets location information for multiple IP addresses at once
func (s *IPGeolocationService) BatchGetLocationByIPs(ctx context.Context, req *BatchIPLookupRequest) ([]*IPLocationInfo, error) {
	if len(req.IPs) == 0 {
		return nil, fmt.Errorf("at least one IP address is required")
	}

	if len(req.IPs) > 100 {
		return nil, fmt.Errorf("batch lookup supports maximum 100 IPs, got %d", len(req.IPs))
	}

	// Validate all IPs
	for _, ip := range req.IPs {
		if err := s.ValidateIP(ip); err != nil {
			return nil, fmt.Errorf("invalid IP %s: %w", ip, err)
		}
	}

	// Build batch request payload
	type batchQuery struct {
		Query  string `json:"query"`
		Fields string `json:"fields,omitempty"`
		Lang   string `json:"lang,omitempty"`
	}

	var queries []batchQuery
	for _, ip := range req.IPs {
		query := batchQuery{Query: ip}
		if len(req.Fields) > 0 {
			query.Fields = strings.Join(req.Fields, ",")
		}
		if req.Language != "" {
			query.Lang = req.Language
		}
		queries = append(queries, query)
	}

	// Make batch request
	resp, err := s.client.Post(ctx, "/batch", queries)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get IP locations: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("batch IP geolocation API error: %s", resp.String())
	}

	var results []*IPLocationInfo
	if err := resp.JSON(&results); err != nil {
		return nil, fmt.Errorf("failed to parse batch IP location response: %w", err)
	}

	return results, nil
}

// GetSimpleLocation returns a simplified location summary
func (s *IPGeolocationService) GetSimpleLocation(ctx context.Context, ip string) (*IPLocationSummary, error) {
	locationInfo, err := s.GetLocationByIP(ctx, &IPLookupRequest{IP: ip})
	if err != nil {
		return nil, err
	}

	return &IPLocationSummary{
		IP:          locationInfo.Query,
		City:        locationInfo.City,
		Region:      locationInfo.RegionName,
		Country:     locationInfo.Country,
		CountryCode: locationInfo.CountryCode,
		Location:    LatLng{Lat: locationInfo.Lat, Lng: locationInfo.Lon},
		Timezone:    locationInfo.Timezone,
		ISP:         locationInfo.ISP,
	}, nil
}

// ValidateIP validates if an IP address is valid (IPv4 or IPv6)
func (s *IPGeolocationService) ValidateIP(ip string) error {
	if ip == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}

	// Check if it's a private/reserved IP
	if s.IsPrivateIP(ip) {
		return fmt.Errorf("private IP addresses are not supported: %s", ip)
	}

	return nil
}

// IsPrivateIP checks if an IP is private/reserved
func (s *IPGeolocationService) IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Define private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",     // Private network
		"172.16.0.0/12",  // Private network
		"192.168.0.0/16", // Private network
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 private
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range privateRanges {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if subnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// IsIPInCountry checks if an IP address is located in a specific country
func (s *IPGeolocationService) IsIPInCountry(ctx context.Context, ip, countryCode string) (bool, error) {
	locationInfo, err := s.GetLocationByIP(ctx, &IPLookupRequest{
		IP:     ip,
		Fields: []string{"status", "countryCode"},
	})
	if err != nil {
		return false, err
	}

	return strings.EqualFold(locationInfo.CountryCode, countryCode), nil
}

// IsIPInCity checks if an IP address is located in a specific city
func (s *IPGeolocationService) IsIPInCity(ctx context.Context, ip, city string) (bool, error) {
	locationInfo, err := s.GetLocationByIP(ctx, &IPLookupRequest{
		IP:     ip,
		Fields: []string{"status", "city"},
	})
	if err != nil {
		return false, err
	}

	return strings.EqualFold(locationInfo.City, city), nil
}

// GetIPsInCountry filters a list of IPs and returns only those in a specific country
func (s *IPGeolocationService) GetIPsInCountry(ctx context.Context, ips []string, countryCode string) ([]string, error) {
	results, err := s.BatchGetLocationByIPs(ctx, &BatchIPLookupRequest{
		IPs:    ips,
		Fields: []string{"status", "countryCode", "query"},
	})
	if err != nil {
		return nil, err
	}

	var filteredIPs []string
	for _, result := range results {
		if strings.EqualFold(result.CountryCode, countryCode) {
			filteredIPs = append(filteredIPs, result.Query)
		}
	}

	return filteredIPs, nil
}

// GetDistanceFromIP calculates distance between an IP's location and a coordinate
func (s *IPGeolocationService) GetDistanceFromIP(ctx context.Context, ip string, target LatLng) (float64, error) {
	locationInfo, err := s.GetLocationByIP(ctx, &IPLookupRequest{
		IP:     ip,
		Fields: []string{"status", "lat", "lon"},
	})
	if err != nil {
		return 0, err
	}

	ipLocation := LatLng{Lat: locationInfo.Lat, Lng: locationInfo.Lon}
	return calculateDistanceHaversine(ipLocation, target), nil
}

// CheckIPInGeoFence checks if an IP's location is within a geo-fence
func (s *IPGeolocationService) CheckIPInGeoFence(ctx context.Context, ip string, geoFence *GeoFence) (*IPLocationSummary, bool, error) {
	// Get IP location
	locationInfo, err := s.GetLocationByIP(ctx, &IPLookupRequest{IP: ip})
	if err != nil {
		return nil, false, fmt.Errorf("failed to get IP location: %w", err)
	}

	// Create summary
	summary := &IPLocationSummary{
		IP:          locationInfo.Query,
		City:        locationInfo.City,
		Region:      locationInfo.RegionName,
		Country:     locationInfo.Country,
		CountryCode: locationInfo.CountryCode,
		Location:    LatLng{Lat: locationInfo.Lat, Lng: locationInfo.Lon},
		Timezone:    locationInfo.Timezone,
		ISP:         locationInfo.ISP,
	}

	// Check if in geo-fence
	ipLocation := LatLng{Lat: locationInfo.Lat, Lng: locationInfo.Lon}
	var isInFence bool

	if geoFence.RadiusKM > 0 {
		// Circle-based geo-fence
		distance := calculateDistanceHaversine(ipLocation, geoFence.Center)
		isInFence = distance <= geoFence.RadiusKM
	} else if len(geoFence.Polygon) >= 3 {
		// Polygon-based geo-fence
		isInFence = isPointInPolygon(ipLocation, geoFence.Polygon)
	}

	return summary, isInFence, nil
}

// GetIPsByTimezone filters IPs by timezone
func (s *IPGeolocationService) GetIPsByTimezone(ctx context.Context, ips []string, timezone string) ([]*IPLocationInfo, error) {
	results, err := s.BatchGetLocationByIPs(ctx, &BatchIPLookupRequest{
		IPs:    ips,
		Fields: []string{"status", "timezone", "query", "city", "country"},
	})
	if err != nil {
		return nil, err
	}

	var filtered []*IPLocationInfo
	for _, result := range results {
		if strings.EqualFold(result.Timezone, timezone) {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// GetIPsByISP filters IPs by ISP
func (s *IPGeolocationService) GetIPsByISP(ctx context.Context, ips []string, ispName string) ([]*IPLocationInfo, error) {
	results, err := s.BatchGetLocationByIPs(ctx, &BatchIPLookupRequest{
		IPs:    ips,
		Fields: []string{"status", "isp", "query", "city", "country"},
	})
	if err != nil {
		return nil, err
	}

	var filtered []*IPLocationInfo
	for _, result := range results {
		if strings.Contains(strings.ToLower(result.ISP), strings.ToLower(ispName)) {
			filtered = append(filtered, result)
		}
	}

	return filtered, nil
}

// CompareIPLocations compares locations of two IP addresses
func (s *IPGeolocationService) CompareIPLocations(ctx context.Context, ip1, ip2 string) (float64, bool, error) {
	results, err := s.BatchGetLocationByIPs(ctx, &BatchIPLookupRequest{
		IPs:    []string{ip1, ip2},
		Fields: []string{"status", "lat", "lon", "countryCode", "city"},
	})
	if err != nil {
		return 0, false, err
	}

	if len(results) != 2 {
		return 0, false, fmt.Errorf("expected 2 results, got %d", len(results))
	}

	loc1 := LatLng{Lat: results[0].Lat, Lng: results[0].Lon}
	loc2 := LatLng{Lat: results[1].Lat, Lng: results[1].Lon}

	distance := calculateDistanceHaversine(loc1, loc2)
	sameCountry := results[0].CountryCode == results[1].CountryCode

	return distance, sameCountry, nil
}

// calculateDistanceHaversine calculates distance between two points (in kilometers)
func calculateDistanceHaversine(point1, point2 LatLng) float64 {
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

// isPointInPolygon checks if a point is inside a polygon
func isPointInPolygon(point LatLng, polygon []LatLng) bool {
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
