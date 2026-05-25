package jobs

import "time"

var projectTimezone = loadProjectTimezone()

func loadProjectTimezone() *time.Location {
	if loc, err := time.LoadLocation("Asia/Ho_Chi_Minh"); err == nil {
		return loc
	}
	return time.UTC
}
