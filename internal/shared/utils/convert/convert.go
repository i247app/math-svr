package convert

import "math"

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
