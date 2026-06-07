package jobs

import "time"

var (
	HoChiMinhTimezone  = loadProjectTimezone("Asia/Ho_Chi_Minh")
	TokyoTimezone      = loadProjectTimezone("Asia/Tokyo")
	NewYorkTimezone    = loadProjectTimezone("America/New_York")
	LosAngelesTimezone = loadProjectTimezone("America/Los_Angeles")
)

func loadProjectTimezone(timezone string) *time.Location {
	if loc, err := time.LoadLocation(timezone); err == nil {
		return loc
	}
	return time.UTC
}
