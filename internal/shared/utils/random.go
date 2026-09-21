package utils

import (
	"math"
	"math/rand"
)

func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func RandomIntWithLength(length int) int {
	max := int(math.Pow10(length))
	min := int(math.Pow10(length - 1))
	return rand.Intn(max-min) + min
}
