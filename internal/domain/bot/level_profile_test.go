package bot

import (
	"strings"
	"testing"
)

// TestLevelProfilesComplete guards the same "one row per band" contract
// the grade profiles have: a half-filled row renders a LEVEL PROFILE block
// with blank sections, which silently drops a difficulty instruction
// instead of failing loudly.
func TestLevelProfilesComplete(t *testing.T) {
	for level := 1; level <= 10; level++ {
		p, ok := levelProfiles[level]
		if !ok {
			t.Errorf("levelProfiles[%d] is missing", level)
			continue
		}
		fields := map[string]string{
			"stepsVN": p.stepsVN, "stepsEN": p.stepsEN,
			"rangeSpotVN": p.rangeSpotVN, "rangeSpotEN": p.rangeSpotEN,
			"formVN": p.formVN, "formEN": p.formEN,
			"distractorVN": p.distractorVN, "distractorEN": p.distractorEN,
		}
		for name, value := range fields {
			if strings.TrimSpace(value) == "" {
				t.Errorf("levelProfiles[%d].%s is empty", level, name)
			}
		}
	}
	if len(levelProfiles) != 10 {
		t.Errorf("levelProfiles has %d entries, want exactly 10", len(levelProfiles))
	}
}

func TestClampLevel(t *testing.T) {
	tests := []struct {
		name  string
		grade GradeLevel
		level int
		want  int
	}{
		{"in range stays put", 3, 7, 7},
		{"below range floors to 1", 3, 0, 1},
		{"negative floors to 1", 3, -4, 1},
		{"above range caps at 10", 3, 42, 10},
		{"kindergarten caps at its ceiling", GradeKindergarten, 9, 4},
		{"kindergarten under ceiling untouched", GradeKindergarten, 2, 2},
		{"grade 5 uses the full scale", 5, 10, 10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampLevel(tc.grade, tc.level); got != tc.want {
				t.Errorf("clampLevel(%d, %d) = %d, want %d", tc.grade, tc.level, got, tc.want)
			}
		})
	}
}

func TestLevelProfileBlock(t *testing.T) {
	t.Run("unknown level yields no block", func(t *testing.T) {
		for _, lang := range []QuizLanguage{QuizLanguageVietnamese, QuizLanguageEnglish} {
			if got := levelProfileBlock(lang, 0); got != "" {
				t.Errorf("%s: expected empty block for level 0, got %q", lang, got)
			}
			if got := levelProfileBlock(lang, 11); got != "" {
				t.Errorf("%s: expected empty block for level 11, got %q", lang, got)
			}
		}
	})

	t.Run("block carries the band's own wording", func(t *testing.T) {
		block := levelProfileBlock(QuizLanguageVietnamese, 9)
		if !strings.Contains(block, levelProfiles[9].stepsVN) {
			t.Error("VN block does not contain the level's steps wording")
		}
		if !strings.Contains(block, levelProfiles[9].distractorVN) {
			t.Error("VN block does not contain the level's distractor wording")
		}
		block = levelProfileBlock(QuizLanguageEnglish, 9)
		if !strings.Contains(block, levelProfiles[9].stepsEN) {
			t.Error("EN block does not contain the level's steps wording")
		}
		if strings.Contains(block, levelProfiles[9].stepsVN) {
			t.Error("EN block leaks the Vietnamese wording")
		}
	})

	t.Run("block subordinates itself to the grade", func(t *testing.T) {
		// The whole point of the ranking sentence: without it the model
		// reaches into the next grade to satisfy a high level.
		for _, lang := range []QuizLanguage{QuizLanguageVietnamese, QuizLanguageEnglish} {
			if block := levelProfileBlock(lang, 10); !strings.Contains(block, "GRADE PROFILE") {
				t.Errorf("%s block never defers to GRADE PROFILE", lang)
			}
		}
	})
}

// TestClampLevelToGrade covers the exported entrance the module layer
// uses. It matters that this is the same clamp the prompt applies: the
// service stamps its result onto the cache tag and the stored row, so a
// second, different clamp deeper down would split identical exams across
// two cache keys.
func TestClampLevelToGrade(t *testing.T) {
	tests := []struct {
		name        string
		grade, want int
		level       int
	}{
		{"kindergarten ceiling applies", 0, 4, 9},
		{"kindergarten below ceiling untouched", 0, 3, 3},
		{"elementary uses the full scale", 2, 10, 10},
		{"out of range floors", 3, 1, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClampLevelToGrade(tc.grade, tc.level); got != tc.want {
				t.Errorf("ClampLevelToGrade(%d, %d) = %d, want %d", tc.grade, tc.level, got, tc.want)
			}
		})
	}
}
