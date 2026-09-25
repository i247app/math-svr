package utils

import (
	crand "crypto/rand"
	"math"
	"math/big"
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

func RandomStringWithLenght(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[num.Int64()]
	}

	return string(result), nil
}
