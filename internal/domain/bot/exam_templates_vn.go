package bot

import (
	"fmt"
	"regexp"
	"strings"
)

// Vietnamese exam-generation templates. Vietnamese only, and not by
// omission: the product serves Vietnamese children and ma_ai_exams has no
// language column, so a stored row could not record being anything else.
//
// JSON keys stay English — the parser, the DB columns and the mobile
// client all share one shape regardless of the prose language.

// systemExamVNTmpl carries three %d slots, all the same question count.
const systemExamVNTmpl = `Bạn là trợ lý tạo bài kiểm tra toán cho trẻ mẫu giáo và học sinh tiểu học Việt Nam (Mẫu giáo, Lớp 1-5).

Hãy tạo CHÍNH XÁC %d câu hỏi trắc nghiệm theo GRADE PROFILE và LEVEL PROFILE mà người dùng cung cấp.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D.
- Chỉ có một phương án đúng.
- Với câu ARITHMETIC, "question_name" chỉ chứa số và toán tử (+, -, *, /, ^, dấu ngoặc, "?") — không lời văn, không LaTeX, không chữ tiếng Việt. Các loại câu khác tuân theo VISUAL QUESTION RULES ở cuối prompt.
- Dùng phân số ASCII như "1/2", không dùng "½".
- Không lặp lại câu hỏi.

QUY TẮC METADATA (BẮT BUỘC ĐỂ CHẤM TỰ ĐỘNG):
- "right_answer_label" là nhãn (A/B/C/D) của phương án đúng.
- "right_answer_content" là GIÁ TRỊ chữ trong "content" của phương án đúng — phải khớp ký tự với "content" tương ứng (vd. "8", "1/2").
- "question_topic" là một mã kỹ năng ngắn bằng tiếng Việt viết thường, ví dụ: cộng trong phạm vi 100, trừ có nhớ, nhân số có một chữ số, chia số có một chữ số, phân số cơ bản, so sánh phân số, số thập phân cơ bản, giá trị theo vị trí, bài toán có lời, phép toán hỗn hợp, hình học cơ bản, đo lường, thời gian và tiền tệ, đếm. Nếu thực sự không phù hợp, tạo mã mới ngắn gọn (≤32 ký tự).
- "question_grade" là cấp lớp mà RIÊNG câu đó nhắm tới, dạng số nguyên (0 = mẫu giáo, 1-5 = lớp 1 đến lớp 5, 6 = trên lớp 5). Xem quy tắc cấp lớp trong phần người dùng cung cấp.
- "question_level" là bậc cường độ của câu đó, số nguyên 1..10, lấy đúng bằng bậc cường độ mà LEVEL PROFILE nêu.

QUY TẮC TITLE & SHORT_TEXT:
- "title" PHẢI đúng bằng chuỗi mà phần người dùng chỉ định, không thêm bớt ký tự nào. KHÔNG tự chọn cấp độ, KHÔNG đặt chủ đề toán vào "title".
- "short_text" là tiêu đề ngắn gọn, cụ thể, mô tả ĐÚNG chủ đề toán của bộ câu hỏi (ví dụ: "Phép cộng và phép trừ trong phạm vi 100", "Phân số cơ bản và so sánh").
- "short_text" tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG kèm loại bài, KHÔNG dùng cụm chung chung như "Bài kiểm tra Toán", "Bài luyện tập" hay "Quiz".

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "title": "Lớp 1 - Cấp độ 1",
  "short_text": "Phép cộng và phép trừ trong phạm vi 20",
  "questions":[
    {
      "question_number": 1,
      "question_type": "ARITHMETIC",
      "question_name": "5 + 3 = ?",
      "answers": [
        {"label": "A", "content": "8"},
        {"label": "B", "content": "9"},
        {"label": "C", "content": "10"},
        {"label": "D", "content": "7"}
      ],
      "right_answer_label": "A",
      "right_answer_content": "8",
      "question_topic": "phép cộng trong phạm vi 20",
      "question_grade": 1,
      "question_level": 1
    }
  ]
}
`

// examVocabulary rewrites the shared grade-profile block into the exam's
// field names.
//
// The block ends in a few-shot example question, and a few-shot example
// outweighs any prose rule — left alone it would teach the model the quiz
// spelling (right_answer, topic, difficulty) two paragraphs before the
// schema asks for the exam one. The alternative was a second exemplar per
// band, which is seven more strings to keep in step with the first seven.
//
// difficulty is dropped rather than renamed: the exam expresses hardness
// as question_level on the 1..10 scale, and an exemplar carrying a stale
// 1..5 value would just contradict the LEVEL PROFILE above it.
var examVocabularyRenames = strings.NewReplacer(
	`"right_answer"`, `"right_answer_label"`,
	`"correct_answer"`, `"right_answer_content"`,
	`"topic"`, `"question_topic"`,
)

var exemplarDifficultyRe = regexp.MustCompile(`,\s*"difficulty":\s*\d+`)

func examVocabulary(block string) string {
	return exemplarDifficultyRe.ReplaceAllString(examVocabularyRenames.Replace(block), "")
}

func buildSystemExamVN(n int) string {
	// The visual rules are shared with the quiz prompts and still name
	// "topic"; the exam vocabulary is applied to them for the same reason
	// it is applied to the exemplar.
	return fmt.Sprintf(systemExamVNTmpl, n, n, n) + examVocabulary(visualQuestionRulesVN)
}

// buildUserExamVN assembles the per-request brief. Order matters: the two
// authoritative profile blocks come first so they outrank the schema
// example the system prompt ended on, then the grade rule for individual
// questions, then the softer curriculum context.
func buildUserExamVN(in ExamPromptInput, n int) string {
	out := ""

	if block := gradeProfileBlockByLevel(QuizLanguageVietnamese, GradeLevel(in.Grade)); block != "" {
		out += examVocabulary(block) + "\n\n"
	}
	if lvl := clampLevel(GradeLevel(in.Grade), in.Level); in.Level > 0 {
		if block := levelProfileBlock(lvl); block != "" {
			out += block + "\n\n"
		}
	}

	out += fmt.Sprintf("Hãy tạo bài kiểm tra loại %s gồm %d câu.\n\n", in.ExamType, n)
	out += "QUY TẮC CẤP LỚP CHO TỪNG CÂU:\n" + examProbeBlockVN(in, n) + "\n"

	if ctx := examContextVN(in); ctx != "" {
		out += "\nThông tin chương trình (chỉ để chọn chủ đề, KHÔNG dùng để tăng hay giảm độ khó):\n" + ctx + "\n"
	}

	out += fmt.Sprintf("\n\"title\" phải đúng bằng: %s", examBandTitle(in.Grade, clampLevel(GradeLevel(in.Grade), in.Level)))
	return out
}
