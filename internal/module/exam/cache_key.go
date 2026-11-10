package exam

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// BuildCacheTag is the ONE function that builds ma_exam_pools.req_extras.
//
// The tag is a cache key, so its only real requirement is that identical
// requests produce an identical string, forever. That is why it lives
// alone here instead of being formatted at the call site: two call sites
// formatting "the same" string is how a cache starts missing silently,
// and a silent miss looks exactly like normal operation while costing a
// generation every time.
//
// Shape: TYPE-G<grade>-L<level>-Q<questions>-S<semester>-P<program>-V<prompt_version>
//
//	GRADE-G1-L3-Q10-SHỌC_KỲ_1-PCÁNH_DIỀU-V1
//	ASSESSMENT-G2-LNA-Q10-SNA-PNA-V4
//
// L is the stated level, LNA when none: a GRADE paper is written at it, and
// GRADE / PRACTICE stamp it as question_level (input.AllowedLevels), so a
// set at one level must never be served for another. An ASSESSMENT ignores
// the stated level — its levels come from a fixed per-position table — so
// its part is assessmentLevelPart whatever the request says; bump it when
// that table changes, so sets stamped under the old one are not served. V is ma_prompts.prompt_version of
// the row the set was generated from (per type, and per grade for ASSESSMENT
// and GRADE): it moves on every admin rewrite of that prompt, so
// a set generated from the old text is never served again — the next
// request misses and generates from the new one. Tags written before V
// existed never match either, which is correct for the same reason: those
// sets were generated from a prompt that no longer exists.
//
// SL<sub_level> follows L only when a request states a sub-level
// (GRADE-G1-L3-SL2-Q10-…): the part is absent otherwise, so every tag
// written before sub_level existed still matches its request.
// assessmentLevelPart names the ASSESSMENT level table a set was stamped
// under (input.assessmentSlotLevels); T1 = the table of 2026-11-10.
const assessmentLevelPart = "LT1"

func BuildCacheTag(examType enum.ExamType, grade int, level, subLevel *int, numQues int, semester, program string, promptVersion int) string {
	levelPart := "LNA"
	switch {
	case examType == enum.ExamTypeAssessment:
		levelPart = assessmentLevelPart
	case level != nil:
		levelPart = fmt.Sprintf("L%d", *level)
	}
	parts := []string{
		normalizeTagPart(string(examType)),
		fmt.Sprintf("G%d", grade),
		levelPart,
	}
	if subLevel != nil {
		parts = append(parts, fmt.Sprintf("SL%d", *subLevel))
	}
	return strings.Join(append(parts,
		fmt.Sprintf("Q%d", numQues),
		"S"+normalizeTagPart(semester),
		"P"+normalizeTagPart(program),
		fmt.Sprintf("V%d", promptVersion),
	), "-")
}

// normalizeTagPart makes a free-text field safe to sit inside the tag:
// upper-cased for case-insensitive equality, whitespace collapsed, and
// the separator itself replaced so a program named "Chân trời - sáng tạo"
// cannot invent an extra field.
//
// Diacritics are deliberately KEPT. Stripping them would need a mapping
// table to maintain, and would fold genuinely different Vietnamese words
// onto one key; the column is utf8mb4 and an exact-match index does not
// care how the bytes look.
func normalizeTagPart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "NA"
	}
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "-", "_")
	return strings.Join(strings.Fields(s), "_")
}
