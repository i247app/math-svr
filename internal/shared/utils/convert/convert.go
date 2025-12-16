package convert

import (
	"math"
	"strings"
)

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func TrimSpace(str string) string {
	return strings.TrimSpace(str)
}