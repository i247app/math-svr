package input

import (
	"strconv"
	"strings"
)

// Placeholders an admin may write into a stored system prompt. They are
// the only text the server fills in; everything else is sent as written.
const (
	PlaceholderLevel = "{{LEVEL}}"       // the round's level, 0–9
	PlaceholderSeed  = "{{RANDOM_SEED}}" // a fresh number per call, to vary the set
)

// RenderSystemPrompt fills the placeholders of a stored system prompt.
// {{LEVEL}} is left as written when the round has no level, so the caller
// can see (HasUnfilledLevel) that the prompt asked for one it did not get.
func RenderSystemPrompt(text string, level *int, seed int64) string {
	pairs := []string{PlaceholderSeed, strconv.FormatInt(seed, 10)}
	if level != nil {
		pairs = append(pairs, PlaceholderLevel, strconv.Itoa(*level))
	}
	return strings.NewReplacer(pairs...).Replace(text)
}

// HasUnfilledLevel reports a rendered prompt that still names {{LEVEL}}.
func HasUnfilledLevel(rendered string) bool {
	return strings.Contains(rendered, PlaceholderLevel)
}
