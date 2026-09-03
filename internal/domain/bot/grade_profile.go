package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
	// IconEssential — icons are the PRIMARY question form (kindergarten,
	// where the child cannot yet read). Text-only questions are the
	// exception, not the rule.
	IconEssential
)

// GradeLevel is the ordinal position of a band in the learning ladder:
// kindergarten sits at 0, elementary grades at 1..5. It is the key of
// gradeProfiles, so the ordering is meaningful — a lower value is always
// an easier band.
//
// Note the deliberate split from the old "0 means unknown" convention:
// 0 is now a real band, so resolveGradeLevel reports unknown through a
// second (ok bool) return instead of a sentinel value.
type GradeLevel int

const (
	// GradeKindergarten is the pre-Grade-1 preparation year (mẫu giáo,
	// 5-6 years old). Children in this band are pre-literate, so the
	// profile leans on counting icons rather than written prose.
	GradeKindergarten GradeLevel = 0
	// GradeElementaryLast is the highest band the product serves today.
	GradeElementaryLast GradeLevel = 5
)

// gradeProfile is the authoritative, code-owned difficulty + visual
// contract for a single band. It is the SINGLE SOURCE OF TRUTH the
// prompt builders render into a high-salience GRADE PROFILE block, so
// question difficulty is anchored to band-appropriate content ranges
// instead of being inferred by the model from a bare grade label.
//
// exemplar is a language-neutral JSON question object (math symbols +
// numeric answers + English topic tag) shown as the band's few-shot
// anchor. It carries far more behavioural weight than any prose rule, so
// it must always sit at the band's true difficulty.
//
// floorEN / floorVN override the shared calibration sentence rendered
// under "Difficulty floor". Leave them empty to inherit the default,
// which only guards against going too EASY; set them on a band whose
// main risk runs the other way (kindergarten, where the model must also
// be stopped from drifting UP into Grade 1 material).
//
// To extend (new band, or later a new subject): add one row to
// gradeProfiles and one entry to the label matcher. No prompt-text edits
// are required.
type gradeProfile struct {
	nameEN     string
	nameVN     string
	icon       IconPolicy
	rangeEN    string
	rangeVN    string
	skillsEN   string
	skillsVN   string
	iconLineEN string
	iconLineVN string
	floorEN    string
	floorVN    string
	exemplar   string
}

var gradeProfiles = map[GradeLevel]gradeProfile{
	GradeKindergarten: {
		nameEN:     "Kindergarten",
		nameVN:     "Mẫu giáo",
		icon:       IconEssential,
		rangeEN:    "whole numbers up to 10 (occasionally to 20 for counting only); adding and subtracting within 5, and never past 10.",
		rangeVN:    "số tự nhiên trong phạm vi 10 (thỉnh thoảng đến 20 nhưng chỉ để đếm); cộng trừ trong phạm vi 5, tuyệt đối không quá 10.",
		skillsEN:   "counting objects, recognising the digits 1-10, comparing more/fewer/equal, matching quantity to digit, simple grouping, recognising basic shapes.",
		skillsVN:   "đếm số lượng đồ vật, nhận biết chữ số 1-10, so sánh nhiều hơn/ít hơn/bằng nhau, nối số lượng với chữ số, gộp/tách đơn giản, nhận biết hình cơ bản.",
		iconLineEN: "ESSENTIAL — the child cannot read yet. Use COUNT with emoji for the MAJORITY of questions (at least 70 percent). Any remaining question is a bare digit sum such as \"2 + 1 = ?\"; never a written word problem.",
		iconLineVN: "BẮT BUỘC — trẻ chưa biết đọc. Dùng COUNT kèm emoji cho ĐA SỐ câu hỏi (ít nhất 70 phần trăm). Các câu còn lại chỉ là phép tính bằng chữ số như \"2 + 1 = ?\"; tuyệt đối không dùng toán đố bằng lời văn.",
		floorEN:    "Difficulty ceiling — this is the LOWEST band, so the risk runs upward, not downward: never exceed the number range above, never use regrouping, multiplication, division, fractions, money, time or multi-step reasoning, and never write a question that requires reading a sentence. Answer options must all be single digits or very short words.",
		floorVN:    "Trần độ khó — đây là bậc THẤP NHẤT nên rủi ro là ra đề quá khó, không phải quá dễ: không vượt phạm vi số nêu trên, không cộng trừ có nhớ, không nhân, chia, phân số, tiền tệ, thời gian hay suy luận nhiều bước, và tuyệt đối không ra câu hỏi buộc trẻ phải đọc một câu văn. Các phương án trả lời phải là chữ số đơn hoặc từ rất ngắn.",
		exemplar:   `{"question_number": 1, "question_type": "COUNT", "question_name": "🐟 🐟 + 🐟 = ?", "answers": [{"label":"A","content":"3"},{"label":"B","content":"2"},{"label":"C","content":"4"},{"label":"D","content":"1"}], "right_answer": "A", "correct_answer": "3", "topic": "counting", "difficulty": 1}`,
	},
	1: {
		nameEN:     "Grade 1",
		nameVN:     "Lớp 1",
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
		nameEN:     "Grade 2",
		nameVN:     "Lớp 2",
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
		nameEN:     "Grade 3",
		nameVN:     "Lớp 3",
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
		nameEN:     "Grade 4",
		nameVN:     "Lớp 4",
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
		nameEN:     "Grade 5",
		nameVN:     "Lớp 5",
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

// defaultFloorEN / defaultFloorVN are the calibration sentence used by
// every band that does not override it. They guard the downward drift
// the elementary grades actually suffer from — the model reaching for
// "count the apples" at Grade 4. A band whose risk runs the other way
// sets floorEN / floorVN instead.
const (
	defaultFloorEN = "Difficulty floor: never drop below this level — no trivial single-digit sums and no object-counting unless the Visual/icon policy above allows it."
	defaultFloorVN = "Sàn độ khó: không được thấp hơn bậc này — không dùng phép cộng một chữ số tầm thường, không đếm vật trừ khi chính sách icon ở trên cho phép."
)

// kindergartenLabelMarkers are the substrings that identify the
// pre-Grade-1 band in a free-form label. Both the accented and the
// unaccented Vietnamese spellings are listed because grade labels reach
// us from three places (the ma_grades seed, an admin edit, and the
// optional `grade_label` request override) and only the first is under
// our control.
//
// These are matched BEFORE any digit is extracted, which is what stops
// "Mẫu giáo 5 tuổi" from being read as Grade 5.
var kindergartenLabelMarkers = []string{
	"kindergarten",
	"preschool",
	"pre-school",
	"pre school",
	"mẫu giáo",
	"mau giao",
	"mầm non",
	"mam non",
	"lớp lá",
	"lop la",
	"nhà trẻ",
	"nha tre",
}

var gradeNumberRe = regexp.MustCompile(`\d+`)

// resolveGradeLevel maps a free-form label such as "Grade 5", "Lớp 5",
// "5", "Kindergarten" or "Mẫu giáo 5 tuổi" onto its band. The second
// return is false when the label names no band we serve, so callers keep
// their grade-agnostic fallback rather than guessing.
//
// Kindergarten is checked first on purpose: its labels routinely carry a
// digit (an age, or "lớp lá 5-6 tuổi") that a digit-first matcher would
// misread as an elementary grade — the most damaging failure available
// here, since it would hand a five-year-old a Grade 5 quiz.
func resolveGradeLevel(label string) (GradeLevel, bool) {
	normalized := strings.ToLower(strings.TrimSpace(label))
	if normalized == "" {
		return 0, false
	}
	for _, marker := range kindergartenLabelMarkers {
		if strings.Contains(normalized, marker) {
			return GradeKindergarten, true
		}
	}
	m := gradeNumberRe.FindString(normalized)
	if m == "" {
		return 0, false
	}
	n, err := strconv.Atoi(m)
	if err != nil || n < 1 || n > int(GradeElementaryLast) {
		return 0, false
	}
	return GradeLevel(n), true
}

// gradeProfileBlock renders the high-salience GRADE PROFILE block for the
// band parsed from gradeLabel, in the requested language. It returns ""
// when the band is unknown so the caller keeps its existing
// grade-agnostic fallback. This block is the authority on difficulty and
// icon policy; the prompt builders place it at the TOP of the user
// message so it outranks the schema's illustrative example.
func gradeProfileBlock(lang QuizLanguage, gradeLabel string) string {
	level, ok := resolveGradeLevel(gradeLabel)
	if !ok {
		return ""
	}
	p, ok := gradeProfiles[level]
	if !ok {
		return ""
	}
	if lang == QuizLanguageEnglish {
		floor := p.floorEN
		if floor == "" {
			floor = defaultFloorEN
		}
		return fmt.Sprintf(`GRADE PROFILE — AUTHORITATIVE. Calibrate EVERY question to %s; this OVERRIDES the difficulty of any example or schema below.
- Number range & operations: %s
- Expected competencies: %s
- Visual/icon policy: %s
- %s
- The JSON schema shown later illustrates FORMAT ONLY — never copy its difficulty.
- Example question at the CORRECT difficulty for %s:
%s`, p.nameEN, p.rangeEN, p.skillsEN, p.iconLineEN, floor, p.nameEN, p.exemplar)
	}
	floor := p.floorVN
	if floor == "" {
		floor = defaultFloorVN
	}
	return fmt.Sprintf(`GRADE PROFILE — BẮT BUỘC. Hiệu chỉnh MỌI câu hỏi đúng theo %s; điều này ƯU TIÊN CAO HƠN độ khó của mọi ví dụ hay schema bên dưới.
- Phạm vi số & phép tính: %s
- Năng lực cần đạt: %s
- Chính sách icon/hình: %s
- %s
- Schema JSON bên dưới CHỈ minh hoạ ĐỊNH DẠNG — tuyệt đối không sao chép độ khó của nó.
- Ví dụ câu hỏi ĐÚNG độ khó cho %s:
%s`, p.nameVN, p.rangeVN, p.skillsVN, p.iconLineVN, floor, p.nameVN, p.exemplar)
}
