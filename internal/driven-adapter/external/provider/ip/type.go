package ip

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// GeoFence represents a geographic boundary
type GeoFence struct {
	Name     string
	Center   LatLng
	RadiusKM float64
	Polygon  []LatLng // Alternative to radius - define polygon boundary
}

// IPLocationInfo represents detailed location information from an IP address
type IPLocationInfo struct {
	Status      string  `json:"status"`      // success or fail
	Message     string  `json:"message"`     // Error message if status is fail
	Country     string  `json:"country"`     // Country name
	CountryCode string  `json:"countryCode"` // Country code (ISO 3166-1 alpha-2)
	Region      string  `json:"region"`      // Region/state code
	RegionName  string  `json:"regionName"`  // Region/state name
	City        string  `json:"city"`        // City name
	Zip         string  `json:"zip"`         // Zip/postal code
	Lat         float64 `json:"lat"`         // Latitude
	Lon         float64 `json:"lon"`         // Longitude
	Timezone    string  `json:"timezone"`    // Timezone (e.g., "Asia/Ho_Chi_Minh")
	ISP         string  `json:"isp"`         // ISP name
	Org         string  `json:"org"`         // Organization name
	AS          string  `json:"as"`          // Autonomous System
	Query       string  `json:"query"`       // IP address used for query
}

// IPLookupRequest represents a request for IP geolocation lookup
type IPLookupRequest struct {
	IP       string
	Fields   []string // Specific fields to return (optional)
	Language string   // Language code (en, de, es, pt-BR, fr, ja, zh-CN, ru)
}

// BatchIPLookupRequest represents a batch IP lookup request
type BatchIPLookupRequest struct {
	IPs      []string
	Fields   []string // Specific fields to return (optional)
	Language string   // Language code
}

// IPLocationSummary represents a simplified location summary
type IPLocationSummary struct {
	IP          string
	City        string
	Region      string
	Country     string
	CountryCode string
	Location    LatLng
	Timezone    string
	ISP         string
}
