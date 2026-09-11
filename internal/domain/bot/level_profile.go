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
// Vietnamese only, deliberately: the exam flow serves a Vietnamese
// product and ma_ai_exams carries no language column, so there is no
// second language for a stored row to be in. Add the English fields the
// day that changes, not before.
type levelProfile struct {
	// steps — how many calculation steps a child must chain.
	steps string
	// rangeSpot — where inside the grade's allowed numbers to sit.
	rangeSpot string
	// form — the question shape typical of this intensity.
	form string
	// distractor — how close the three wrong options should be.
	distractor string
}

// levelProfiles is the draft calibration reviewed with the teaching team.
// Editing a row changes difficulty across every grade at once, since the
// same intensity is applied inside whatever range the grade allows.
var levelProfiles = map[int]levelProfile{
	1: {
		steps:      "1 bước tính",
		rangeSpot:  "nửa dưới của phạm vi, chọn những số nhỏ nhất",
		form:       "tính trực tiếp, dạng a + b = ?",
		distractor: "lệch xa, sai rõ ràng để trẻ loại được ngay",
	},
	2: {
		steps:      "1 bước tính",
		rangeSpot:  "nửa dưới của phạm vi",
		form:       "tính trực tiếp; có thể xen câu so sánh nhiều hơn/ít hơn",
		distractor: "lệch xa",
	},
	3: {
		steps:      "1 bước tính",
		rangeSpot:  "khoảng giữa phạm vi",
		form:       "thêm dạng điền khuyết, ví dụ a + ? = c",
		distractor: "lệch 1 đến 2 đơn vị",
	},
	4: {
		steps:      "1 đến 2 bước tính",
		rangeSpot:  "khoảng giữa phạm vi",
		form:       "điền khuyết, so sánh, sắp thứ tự",
		distractor: "lệch 1 đến 2 đơn vị",
	},
	5: {
		steps:      "2 bước tính",
		rangeSpot:  "từ giữa lên phần trên của phạm vi",
		form:       "ghép hai phép tính cùng loại trong một câu",
		distractor: "trùng với kết quả của một lỗi tính sai điển hình",
	},
	6: {
		steps:      "2 bước tính",
		rangeSpot:  "phần trên của phạm vi",
		form:       "toán đố một câu ngắn, chỉ khi cấp lớp đã đọc được chữ",
		distractor: "trùng với lỗi điển hình: quên nhớ, sai thứ tự phép tính",
	},
	7: {
		steps:      "2 đến 3 bước tính",
		rangeSpot:  "sát trần của phạm vi",
		form:       "kết hợp hai kỹ năng khác nhau trong cùng một câu",
		distractor: "sát, lệch đúng 1 đơn vị",
	},
	8: {
		steps:      "2 đến 3 bước tính",
		rangeSpot:  "sát trần của phạm vi",
		form:       "kết hợp hai kỹ năng, kèm đổi đơn vị nếu cấp lớp đã học",
		distractor: "sát",
	},
	9: {
		steps:      "từ 3 bước tính trở lên",
		rangeSpot:  "sát trần của phạm vi",
		form:       "toán đố nhiều bước hoặc suy luận ngược từ kết quả",
		distractor: "sát nhất, mỗi phương án ứng với một bước làm sai",
	},
	10: {
		steps:      "từ 3 bước tính trở lên",
		rangeSpot:  "kịch trần của phạm vi",
		form:       "nhiều bước kết hợp suy luận ngược",
		distractor: "sát nhất, mỗi phương án ứng với một bước làm sai",
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

// levelProfileBlock renders the LEVEL PROFILE section. It returns "" for a
// level we do not describe, so the caller keeps its level-agnostic
// fallback instead of emitting a half-empty block.
func levelProfileBlock(level int) string {
	p, ok := levelProfiles[level]
	if !ok {
		return ""
	}
	return fmt.Sprintf(`LEVEL PROFILE — cường độ bậc %d trên 10.
- Số bước tính: %s
- Chọn số ở: %s
- Dạng câu: %s
- Phương án nhiễu: %s
- GRADE PROFILE ở trên quyết định NỘI DUNG. Bậc cường độ này chỉ được làm khó lên TRONG phạm vi đó; tuyệt đối không mượn kiến thức của lớp cao hơn để tăng độ khó.`,
		level, p.steps, p.rangeSpot, p.form, p.distractor)
}
