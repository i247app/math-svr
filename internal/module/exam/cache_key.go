package exam

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// BuildCacheTag is the ONE function that builds ma_ai_exams.req_extras.
//
// The tag is a cache key, so its only real requirement is that identical
// requests produce an identical string, forever. That is why it lives
// alone here instead of being formatted at the call site: two call sites
// formatting "the same" string is how a cache starts missing silently,
// and a silent miss looks exactly like normal operation while costing a
// generation every time.
//
// Shape: TYPE-G<grade>-Q<questions>-S<semester>-P<program>
//
//	ASSESSMENT-G1-Q10-SHOC_KY_1-PCANH_DIEU
//
// There is no level segment. An earlier shape carried L<level>/LNA between
// grade and question count; it was removed with the level axis itself
// (the teaching team has no rule for it, so no request can vary by it and
// the segment could only ever have been the LNA constant). Tags written in
// the old shape simply never match again, which is correct — those exams
// were generated from a prompt that no longer exists.
func BuildCacheTag(examType enum.ExamType, grade int, level *int, numQues int, semester, program string) string {
	levelPart := "LNA"
	if level != nil {
		levelPart = fmt.Sprintf("L%d", *level)
	}
	return strings.Join([]string{
		normalizeTagPart(string(examType)),
		fmt.Sprintf("G%d", grade),
		levelPart,
		fmt.Sprintf("Q%d", numQues),
		"S" + normalizeTagPart(semester),
		"P" + normalizeTagPart(program),
	}, "-")
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
