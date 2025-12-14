package geo

// LatLng represents a geographic coordinate
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// GeocodingResult represents a geocoding result from Google Maps API
type GeocodingResult struct {
	AddressComponents []AddressComponent `json:"address_components"`
	FormattedAddress  string             `json:"formatted_address"`
	Geometry          Geometry           `json:"geometry"`
	PlaceID           string             `json:"place_id"`
	Types             []string           `json:"types"`
}

// AddressComponent represents a component of an address
type AddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

// Geometry represents geometric information
type Geometry struct {
	Location     LatLng    `json:"location"`
	LocationType string    `json:"location_type"`
	Viewport     Viewport  `json:"viewport"`
	Bounds       *Viewport `json:"bounds,omitempty"`
}

// Viewport represents a bounding box
type Viewport struct {
	Northeast LatLng `json:"northeast"`
	Southwest LatLng `json:"southwest"`
}

// GeocodingResponse represents the response from Google Maps Geocoding API
type GeocodingResponse struct {
	Results      []GeocodingResult `json:"results"`
	Status       string            `json:"status"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// ReverseGeocodeRequest represents a reverse geocoding request (lat/lng to address)
type ReverseGeocodeRequest struct {
	Lat          float64
	Lng          float64
	ResultType   []string // Filter by result type (e.g., "street_address", "locality")
	LocationType []string // Filter by location type (e.g., "ROOFTOP", "APPROXIMATE")
	Language     string   // Language code (e.g., "en", "vi")
}

// ForwardGeocodeRequest represents a forward geocoding request (address to lat/lng)
type ForwardGeocodeRequest struct {
	Address    string
	Components map[string]string // Filter by components (e.g., "country:US", "postal_code:10001")
	Bounds     *Viewport         // Bias results to viewport
	Language   string            // Language code
	Region     string            // Region code for biasing
}

// GeoFence represents a geographic boundary
type GeoFence struct {
	Name     string
	Center   LatLng
	RadiusKM float64
	Polygon  []LatLng // Alternative to radius - define polygon boundary
}

// LocationInfo represents detailed location information
type LocationInfo struct {
	Address          string
	City             string
	State            string
	Country          string
	PostalCode       string
	CountryCode      string
	FormattedAddress string
	Location         LatLng
	PlaceID          string
}
