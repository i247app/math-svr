package bot

import "fmt"

// Level is the INTENSITY axis of an exam, 1..10, and it is a different
// question from grade. Grade decides WHAT content is allowed — the number
// range, the operations, whether icons are appropriate. Level decides HOW
// HARD the questions are inside that content: how many steps, where in the
// allowed range the numbers sit, how close the wrong options are.
//
// The two are ranked, not equal. When they disagree the grade wins, and
// the block below says so out loud, because the failure that actually
// happens is a model reaching into next year's syllabus to satisfy "level
// 9" instead of writing a genuinely hard question inside this year's.
//
// Each row is written twice, once per prompt language. Both describe the
// same intensity; which one is rendered follows the language of the
// prompt around it, not the child's — the round's content is always
// Vietnamese regardless.
type levelProfile struct {
	// steps — how many calculation steps a child must chain.
	stepsVN, stepsEN string
	// rangeSpot — where inside the grade's allowed numbers to sit.
	rangeSpotVN, rangeSpotEN string
	// form — the question shape typical of this intensity.
	formVN, formEN string
	// distractor — how close the three wrong options should be.
	distractorVN, distractorEN string
}

// levelProfiles is the draft calibration reviewed with the teaching team.
// Editing a row changes difficulty across every grade at once, since the
// same intensity is applied inside whatever range the grade allows.
var levelProfiles = map[int]levelProfile{
	1: {
		stepsVN: "1 bước tính", stepsEN: "1 step",
		rangeSpotVN: "nửa dưới của phạm vi, chọn những số nhỏ nhất", rangeSpotEN: "lower half of the range, the smallest numbers",
		formVN: "tính trực tiếp, dạng a + b = ?", formEN: "direct computation, a + b = ?",
		distractorVN: "lệch xa, sai rõ ràng để trẻ loại được ngay", distractorEN: "far off, obviously wrong so the child rules them out at once",
	},
	2: {
		stepsVN: "1 bước tính", stepsEN: "1 step",
		rangeSpotVN: "nửa dưới của phạm vi", rangeSpotEN: "lower half of the range",
		formVN: "tính trực tiếp; có thể xen câu so sánh nhiều hơn/ít hơn", formEN: "direct computation; may mix in more/fewer comparisons",
		distractorVN: "lệch xa", distractorEN: "far off",
	},
	3: {
		stepsVN: "1 bước tính", stepsEN: "1 step",
		rangeSpotVN: "khoảng giữa phạm vi", rangeSpotEN: "middle of the range",
		formVN: "thêm dạng điền khuyết, ví dụ a + ? = c", formEN: "add fill-in-the-blank, e.g. a + ? = c",
		distractorVN: "lệch 1 đến 2 đơn vị", distractorEN: "off by 1 to 2",
	},
	4: {
		stepsVN: "1 đến 2 bước tính", stepsEN: "1 to 2 steps",
		rangeSpotVN: "khoảng giữa phạm vi", rangeSpotEN: "middle of the range",
		formVN: "điền khuyết, so sánh, sắp thứ tự", formEN: "fill-in-the-blank, comparison, ordering",
		distractorVN: "lệch 1 đến 2 đơn vị", distractorEN: "off by 1 to 2",
	},
	5: {
		stepsVN: "2 bước tính", stepsEN: "2 steps",
		rangeSpotVN: "từ giữa lên phần trên của phạm vi", rangeSpotEN: "middle to upper part of the range",
		formVN: "ghép hai phép tính cùng loại trong một câu", formEN: "two operations of the same kind chained in one question",
		distractorVN: "trùng với kết quả của một lỗi tính sai điển hình", distractorEN: "equal to the result of a typical miscalculation",
	},
	6: {
		stepsVN: "2 bước tính", stepsEN: "2 steps",
		rangeSpotVN: "phần trên của phạm vi", rangeSpotEN: "upper part of the range",
		formVN: "toán đố một câu ngắn, chỉ khi cấp lớp đã đọc được chữ", formEN: "one-sentence word problem, only when the grade can read",
		distractorVN: "trùng với lỗi điển hình: quên nhớ, sai thứ tự phép tính", distractorEN: "equal to a typical error: forgotten carry, wrong operation order",
	},
	7: {
		stepsVN: "2 đến 3 bước tính", stepsEN: "2 to 3 steps",
		rangeSpotVN: "sát trần của phạm vi", rangeSpotEN: "near the ceiling of the range",
		formVN: "kết hợp hai kỹ năng khác nhau trong cùng một câu", formEN: "two different skills combined in one question",
		distractorVN: "sát, lệch đúng 1 đơn vị", distractorEN: "close, off by exactly 1",
	},
	8: {
		stepsVN: "2 đến 3 bước tính", stepsEN: "2 to 3 steps",
		rangeSpotVN: "sát trần của phạm vi", rangeSpotEN: "near the ceiling of the range",
		formVN: "kết hợp hai kỹ năng, kèm đổi đơn vị nếu cấp lớp đã học", formEN: "two skills combined, with unit conversion if the grade has learned it",
		distractorVN: "sát", distractorEN: "close",
	},
	9: {
		stepsVN: "từ 3 bước tính trở lên", stepsEN: "3 or more steps",
		rangeSpotVN: "sát trần của phạm vi", rangeSpotEN: "near the ceiling of the range",
		formVN: "toán đố nhiều bước hoặc suy luận ngược từ kết quả", formEN: "multi-step word problem, or reasoning backwards from a result",
		distractorVN: "sát nhất, mỗi phương án ứng với một bước làm sai", distractorEN: "closest; each option matches one wrong step",
	},
	10: {
		stepsVN: "từ 3 bước tính trở lên", stepsEN: "3 or more steps",
		rangeSpotVN: "kịch trần của phạm vi", rangeSpotEN: "at the ceiling of the range",
		formVN: "nhiều bước kết hợp suy luận ngược", formEN: "multi-step combined with backward reasoning",
		distractorVN: "sát nhất, mỗi phương án ứng với một bước làm sai", distractorEN: "closest; each option matches one wrong step",
	},
}

// levelCeilingByGrade caps how hard a band can be pushed. Kindergarten is
// the case that forced this: its content ceiling is fixed by the fact that
// the child cannot read yet, so "level 10 kindergarten" has nowhere to go
// and asking for it only tempts the model past the band. Every other grade
// uses the full scale until the teaching team says otherwise.
var levelCeilingByGrade = map[GradeLevel]int{
	GradeKindergarten: 4,
}

// ClampLevelToGrade is the exported entrance, used by the module layer so
// the level that reaches the cache tag, the stored req_level and the
// prompt is one and the same number.
//
// Clamping only inside the prompt builder would have been enough to make
// the QUESTIONS right, and wrong for everything else: a kindergarten
// request for level 9 would be stored as 9, tagged as 9, and cached apart
// from the identical level-4 exam it actually produced.
func ClampLevelToGrade(grade, level int) int {
	return clampLevel(GradeLevel(grade), level)
}

// clampLevel brings a level inside [1,10] and then inside whatever ceiling
// the grade imposes. Callers pass the request's raw level; the prompt
// renders the clamped one, so an out-of-range request degrades to the
// nearest sane band rather than being rejected.
func clampLevel(grade GradeLevel, level int) int {
	if level < 1 {
		level = 1
	}
	if level > 10 {
		level = 10
	}
	if ceiling, ok := levelCeilingByGrade[grade]; ok && level > ceiling {
		return ceiling
	}
	return level
}

// levelProfileBlock renders the LEVEL PROFILE section in the prompt's
// language. It returns "" for a level we do not describe, so the caller
// keeps its level-agnostic fallback instead of emitting a half-empty block.
func levelProfileBlock(lang QuizLanguage, level int) string {
	p, ok := levelProfiles[level]
	if !ok {
		return ""
	}
	if lang == QuizLanguageEnglish {
		return fmt.Sprintf(`LEVEL PROFILE — intensity %d of 10.
- Steps: %s
- Numbers from: %s
- Question form: %s
- Distractors: %s
- The GRADE PROFILE above decides the CONTENT. This intensity may only raise difficulty INSIDE that range; never borrow a higher grade's material to make a question harder.`,
			level, p.stepsEN, p.rangeSpotEN, p.formEN, p.distractorEN)
	}
	return fmt.Sprintf(`LEVEL PROFILE — cường độ bậc %d trên 10.
- Số bước tính: %s
- Chọn số ở: %s
- Dạng câu: %s
- Phương án nhiễu: %s
- GRADE PROFILE ở trên quyết định NỘI DUNG. Bậc cường độ này chỉ được làm khó lên TRONG phạm vi đó; tuyệt đối không mượn kiến thức của lớp cao hơn để tăng độ khó.`,
		level, p.stepsVN, p.rangeSpotVN, p.formVN, p.distractorVN)
}
