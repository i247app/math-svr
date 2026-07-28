package bot

import (
	"fmt"
	"regexp"
	"strconv"
)

// IconPolicy is the grade-band decision on whether visual (COUNT / icon /
// emoji) questions are allowed. It is decided in code, not left to the
// model, so difficulty can never regress into "count the emoji" at a
// grade where that is below level.
type IconPolicy int

const (
	// IconOff — never use icons/emoji (upper grades). ARITHMETIC only.
	IconOff IconPolicy = iota
	// IconRare — icons only to visualise fractions/shapes, otherwise text.
	IconRare
	// IconLight — icons optional, sparing, only as a comprehension aid.
	IconLight
	// IconEncouraged — icons welcome as counting aids (early grades).
	IconEncouraged
)

// gradeProfile is the authoritative, code-owned difficulty + visual
// contract for a single elementary grade. It is the SINGLE SOURCE OF
// TRUTH the prompt builders render into a high-salience GRADE PROFILE
// block, so question difficulty is anchored to grade-appropriate content
// ranges instead of being inferred by the model from a bare grade label.
//
// exemplar is a language-neutral JSON question object (math symbols +
// numeric answers + English topic tag) shown as the grade's few-shot
// anchor. It carries far more behavioural weight than any prose rule, so
// it must always sit at the grade's true difficulty.
//
// To extend (new grade, or later a new subject): add one row to
// gradeProfiles. No prompt-text edits are required.
type gradeProfile struct {
	grade      int
	icon       IconPolicy
	rangeEN    string
	rangeVN    string
	skillsEN   string
	skillsVN   string
	iconLineEN string
	iconLineVN string
	exemplar   string
}

var gradeProfiles = map[int]gradeProfile{
	1: {
		grade:      1,
		icon:       IconEncouraged,
		rangeEN:    "whole numbers up to 20, then up to 100; addition and subtraction without regrouping.",
		rangeVN:    "số tự nhiên đến 20, rồi đến 100; cộng trừ không nhớ.",
		skillsEN:   "counting and cardinality, comparing quantities, simple addition/subtraction, recognising basic shapes.",
		skillsVN:   "đếm và số lượng, so sánh số lượng, cộng trừ đơn giản, nhận diện hình cơ bản.",
		iconLineEN: "ENCOURAGED — icons/emoji are welcome as counting aids for up to about 40 percent of questions; COUNT question_type is appropriate here.",
		iconLineVN: "KHUYẾN KHÍCH — icon/emoji rất phù hợp làm công cụ đếm, tối đa khoảng 40 phần trăm số câu; dùng question_type COUNT là hợp lý.",
		exemplar:   `{"question_number": 1, "question_type": "COUNT", "question_name": "🍎 🍎 🍎 + 🍎 🍎 = ?", "answers": [{"label":"A","content":"5"},{"label":"B","content":"4"},{"label":"C","content":"6"},{"label":"D","content":"3"}], "right_answer": "A", "correct_answer": "5", "topic": "counting", "difficulty": 1}`,
	},
	2: {
		grade:      2,
		icon:       IconLight,
		rangeEN:    "whole numbers up to 1000; addition and subtraction with regrouping; multiplication tables 2-5.",
		rangeVN:    "số đến 1000; cộng trừ có nhớ; bảng nhân 2-5.",
		skillsEN:   "regrouping, early multiplication, basic measurement (length, time).",
		skillsVN:   "cộng trừ có nhớ, nhân nhập môn, đo lường cơ bản (độ dài, thời gian).",
		iconLineEN: "LIGHT — icons are optional (about one in six questions at most), and only when they genuinely aid understanding.",
		iconLineVN: "NHẸ — icon tùy chọn (nhiều nhất khoảng 1/6 số câu), chỉ khi thực sự giúp hiểu bài.",
		exemplar:   `{"question_number": 1, "question_type": "ARITHMETIC", "question_name": "45 + 38 = ?", "answers": [{"label":"A","content":"83"},{"label":"B","content":"73"},{"label":"C","content":"84"},{"label":"D","content":"82"}], "right_answer": "A", "correct_answer": "83", "topic": "addition_regrouping", "difficulty": 2}`,
	},
	3: {
		grade:      3,
		icon:       IconRare,
		rangeEN:    "numbers up to 10,000; multiplication and division within the tables; introduction to fractions (1/2, 1/3, ...); perimeter.",
		rangeVN:    "số đến 10.000; nhân chia trong bảng; phân số nhập môn (1/2, 1/3, ...); chu vi.",
		skillsEN:   "multiplication/division fluency, unit fractions, perimeter, simple word problems.",
		skillsVN:   "thành thạo nhân chia, phân số đơn vị, chu vi, toán đố đơn giản.",
		iconLineEN: "RARE — use icons ONLY to visualise fractions or shapes; every other question is ARITHMETIC. Keep any COUNT question below one in five.",
		iconLineVN: "HIẾM — CHỈ dùng icon để minh hoạ phân số hoặc hình; mọi câu khác là ARITHMETIC. Giữ số câu COUNT dưới 1/5.",
		exemplar:   `{"question_number": 1, "question_type": "ARITHMETIC", "question_name": "7 * 8 = ?", "answers": [{"label":"A","content":"56"},{"label":"B","content":"54"},{"label":"C","content":"63"},{"label":"D","content":"48"}], "right_answer": "A", "correct_answer": "56", "topic": "multiplication_single_digit", "difficulty": 2}`,
	},
	4: {
		grade:      4,
		icon:       IconOff,
		rangeEN:    "large multi-digit numbers; multi-digit multiplication and division; fractions (equivalent, add/subtract like denominators); introduction to decimals; area.",
		rangeVN:    "số nhiều chữ số lớn; nhân chia nhiều chữ số; phân số (bằng nhau, cộng trừ cùng mẫu); nhập môn số thập phân; diện tích.",
		skillsEN:   "multi-digit arithmetic, fraction operations, decimals, area, multi-step problems.",
		skillsVN:   "số học nhiều chữ số, phép tính phân số, số thập phân, diện tích, toán nhiều bước.",
		iconLineEN: `OFF — do NOT use any emoji or [icon:...] tokens; use question_type "ARITHMETIC" for every question.`,
		iconLineVN: `TẮT — TUYỆT ĐỐI không dùng emoji hay token [icon:...]; dùng question_type "ARITHMETIC" cho mọi câu.`,
		exemplar:   `{"question_number": 1, "question_type": "ARITHMETIC", "question_name": "236 * 4 = ?", "answers": [{"label":"A","content":"944"},{"label":"B","content":"924"},{"label":"C","content":"844"},{"label":"D","content":"952"}], "right_answer": "A", "correct_answer": "944", "topic": "multiplication_multi_digit", "difficulty": 3}`,
	},
	5: {
		grade:      5,
		icon:       IconOff,
		rangeEN:    "operations on fractions and decimals; percentages and ratios; geometry (area and volume).",
		rangeVN:    "phép tính phân số và thập phân; tỉ lệ phần trăm và tỉ số; hình học (diện tích, thể tích).",
		skillsEN:   "fraction and decimal arithmetic, percentages, ratio, geometry, multi-step word problems.",
		skillsVN:   "phép tính phân số/thập phân, phần trăm, tỉ số, hình học, toán đố nhiều bước.",
		iconLineEN: `OFF — do NOT use any emoji or [icon:...] tokens; use question_type "ARITHMETIC" for every question.`,
		iconLineVN: `TẮT — TUYỆT ĐỐI không dùng emoji hay token [icon:...]; dùng question_type "ARITHMETIC" cho mọi câu.`,
		exemplar:   `{"question_number": 1, "question_type": "ARITHMETIC", "question_name": "3/4 + 2/5 = ?", "answers": [{"label":"A","content":"23/20"},{"label":"B","content":"5/9"},{"label":"C","content":"6/20"},{"label":"D","content":"1"}], "right_answer": "A", "correct_answer": "23/20", "topic": "fractions_add_sub", "difficulty": 4}`,
	},
}

var gradeNumberRe = regexp.MustCompile(`\d+`)

// parseGradeNumber extracts the elementary grade (1..5) from a free-form
// label such as "Grade 5", "Lớp 5", or "5". Returns 0 when no in-range
// number is found, so callers fall back to grade-agnostic guidance.
func parseGradeNumber(label string) int {
	m := gradeNumberRe.FindString(label)
	if m == "" {
		return 0
	}
	n, err := strconv.Atoi(m)
	if err != nil || n < 1 || n > 5 {
		return 0
	}
	return n
}

// gradeProfileBlock renders the high-salience GRADE PROFILE block for the
// grade parsed from gradeLabel, in the requested language. It returns ""
// when the grade is unknown (grade 0) so the caller keeps its existing
// grade-agnostic fallback. This block is the authority on difficulty and
// icon policy; the prompt builders place it at the TOP of the user
// message so it outranks the schema's illustrative example.
func gradeProfileBlock(lang QuizLanguage, gradeLabel string) string {
	p, ok := gradeProfiles[parseGradeNumber(gradeLabel)]
	if !ok {
		return ""
	}
	if lang == QuizLanguageEnglish {
		return fmt.Sprintf(`GRADE PROFILE — AUTHORITATIVE. Calibrate EVERY question to grade %d; this OVERRIDES the difficulty of any example or schema below.
- Number range & operations: %s
- Expected competencies: %s
- Visual/icon policy: %s
- Difficulty floor: never drop below this grade — no trivial single-digit sums and no object-counting unless the icon policy above is "ENCOURAGED".
- The JSON schema shown later illustrates FORMAT ONLY — never copy its difficulty.
- Example question at the CORRECT difficulty for grade %d:
%s`, p.grade, p.rangeEN, p.skillsEN, p.iconLineEN, p.grade, p.exemplar)
	}
	return fmt.Sprintf(`GRADE PROFILE — BẮT BUỘC. Hiệu chỉnh MỌI câu hỏi đúng theo lớp %d; điều này ƯU TIÊN CAO HƠN độ khó của mọi ví dụ hay schema bên dưới.
- Phạm vi số & phép tính: %s
- Năng lực cần đạt: %s
- Chính sách icon/hình: %s
- Sàn độ khó: không được thấp hơn lớp này — không dùng phép cộng một chữ số tầm thường, không đếm vật trừ khi chính sách icon ở trên là "KHUYẾN KHÍCH".
- Schema JSON bên dưới CHỈ minh hoạ ĐỊNH DẠNG — tuyệt đối không sao chép độ khó của nó.
- Ví dụ câu hỏi ĐÚNG độ khó cho lớp %d:
%s`, p.grade, p.rangeVN, p.skillsVN, p.iconLineVN, p.grade, p.exemplar)
}
