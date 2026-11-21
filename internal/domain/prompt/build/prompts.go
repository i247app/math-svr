package input

import (
	"fmt"
	"strings"
)

// requireFields takes alternating (value, name) pairs and returns the
// first missing-field error encountered. Keeps the per-kind validation
// readable at the call site.
func requireFields(pairs ...string) error {
	if len(pairs)%2 != 0 {
		return fmt.Errorf("prompt: requireFields expects (value, name) pairs")
	}
	for i := 0; i < len(pairs); i += 2 {
		if strings.TrimSpace(pairs[i]) == "" {
			return fmt.Errorf("prompt: %s is required", pairs[i+1])
		}
	}
	return nil
}

// defaultNumQuestions is the fallback used when a caller leaves the
// question count at zero. Kept in sync with the module validators'
// DefaultNumQuestions (10), but duplicated here so the domain builders
// remain usable without going through the module layer.
const defaultNumQuestions = 10

func resolveNumQuestions(n int) int {
	if n <= 0 {
		return defaultNumQuestions
	}
	return n
}
