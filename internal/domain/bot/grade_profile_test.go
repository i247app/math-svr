package bot

import (
	"strings"
	"testing"
)

func TestResolveGradeLevel(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantLevel GradeLevel
		wantOK    bool
	}{
		{name: "english kindergarten", label: "Kindergarten", wantLevel: GradeKindergarten, wantOK: true},
		{name: "vietnamese kindergarten", label: "Mẫu giáo", wantLevel: GradeKindergarten, wantOK: true},
		{name: "kindergarten without diacritics", label: "Mau giao", wantLevel: GradeKindergarten, wantOK: true},
		// The regression this matcher exists for: an age in the label must
		// never be read as an elementary grade number.
		{name: "kindergarten label carrying an age", label: "Mẫu giáo 5 tuổi", wantLevel: GradeKindergarten, wantOK: true},
		{name: "lop la with age range", label: "Lớp Lá (5-6 tuổi)", wantLevel: GradeKindergarten, wantOK: true},
		{name: "preschool", label: "Preschool", wantLevel: GradeKindergarten, wantOK: true},
		{name: "mam non", label: "Mầm non", wantLevel: GradeKindergarten, wantOK: true},

		{name: "english grade", label: "Grade 1", wantLevel: 1, wantOK: true},
		{name: "vietnamese grade", label: "Lớp 3", wantLevel: 3, wantOK: true},
		{name: "bare number", label: "5", wantLevel: 5, wantOK: true},

		{name: "above the served range", label: "Grade 6", wantLevel: 0, wantOK: false},
		{name: "zero is not a grade number", label: "Grade 0", wantLevel: 0, wantOK: false},
		{name: "empty", label: "", wantLevel: 0, wantOK: false},
		{name: "no level in label", label: "Math", wantLevel: 0, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotLevel, gotOK := resolveGradeLevel(tc.label)
			if gotOK != tc.wantOK {
				t.Fatalf("resolveGradeLevel(%q) ok = %v, want %v", tc.label, gotOK, tc.wantOK)
			}
			if gotOK && gotLevel != tc.wantLevel {
				t.Fatalf("resolveGradeLevel(%q) level = %d, want %d", tc.label, gotLevel, tc.wantLevel)
			}
		})
	}
}

// TestGradeProfilesComplete guards the "add one row to extend" contract:
// a half-filled row would render a GRADE PROFILE block with blank
// sections, which silently degrades difficulty calibration.
func TestGradeProfilesComplete(t *testing.T) {
	for level, p := range gradeProfiles {
		fields := map[string]string{
			"nameEN":     p.nameEN,
			"nameVN":     p.nameVN,
			"rangeEN":    p.rangeEN,
			"rangeVN":    p.rangeVN,
			"skillsEN":   p.skillsEN,
			"skillsVN":   p.skillsVN,
			"iconLineEN": p.iconLineEN,
			"iconLineVN": p.iconLineVN,
			"exemplar":   p.exemplar,
		}
		for name, value := range fields {
			if strings.TrimSpace(value) == "" {
				t.Errorf("gradeProfiles[%d].%s is empty", level, name)
			}
		}
	}
}

func TestGradeProfileBlock(t *testing.T) {
	t.Run("unknown label yields no block", func(t *testing.T) {
		if got := gradeProfileBlock(QuizLanguageVietnamese, "Math"); got != "" {
			t.Fatalf("expected empty block for unknown label, got %q", got)
		}
	})

	t.Run("kindergarten renders its own ceiling and name", func(t *testing.T) {
		vn := gradeProfileBlock(QuizLanguageVietnamese, "Mẫu giáo")
		if !strings.Contains(vn, "Mẫu giáo") {
			t.Errorf("VN block missing the band name: %q", vn)
		}
		if !strings.Contains(vn, gradeProfiles[GradeKindergarten].floorVN) {
			t.Errorf("VN block did not use the kindergarten ceiling override: %q", vn)
		}
		if strings.Contains(vn, defaultFloorVN) {
			t.Errorf("VN block fell back to the default floor sentence: %q", vn)
		}

		en := gradeProfileBlock(QuizLanguageEnglish, "Kindergarten")
		if !strings.Contains(en, "Kindergarten") {
			t.Errorf("EN block missing the band name: %q", en)
		}
		if !strings.Contains(en, gradeProfiles[GradeKindergarten].floorEN) {
			t.Errorf("EN block did not use the kindergarten ceiling override: %q", en)
		}
	})

	t.Run("elementary grade inherits the default floor", func(t *testing.T) {
		en := gradeProfileBlock(QuizLanguageEnglish, "Grade 4")
		if !strings.Contains(en, "Grade 4") {
			t.Errorf("EN block missing the band name: %q", en)
		}
		if !strings.Contains(en, defaultFloorEN) {
			t.Errorf("EN block did not inherit the default floor: %q", en)
		}
	})
}
