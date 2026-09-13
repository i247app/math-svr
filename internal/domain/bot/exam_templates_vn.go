package bot

import (
	"fmt"
	"regexp"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Vietnamese exam-generation templates. Vietnamese only, and not by
// omission: the product serves Vietnamese children and ma_ai_exams has no
// language column, so a stored row could not record being anything else.
//
// JSON keys stay English — the parser, the DB columns and the mobile
// client all share one shape regardless of the prose language.

// systemExamVNTmpl carries three %d slots, all the same question count.
const systemExamVNTmpl = `Bạn là trợ lý tạo bài kiểm tra toán cho trẻ mẫu giáo và học sinh tiểu học Việt Nam (Mẫu giáo, Lớp 1-5).

Hãy tạo CHÍNH XÁC %d câu hỏi trắc nghiệm theo GRADE PROFILE mà người dùng cung cấp.

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

QUY TẮC TITLE & SHORT_TEXT:
- "title" PHẢI đúng bằng chuỗi mà phần người dùng chỉ định, không thêm bớt ký tự nào. KHÔNG tự thêm cấp độ hay mức độ khó, KHÔNG đặt chủ đề toán vào "title".
- "short_text" là tiêu đề ngắn gọn, cụ thể, mô tả ĐÚNG chủ đề toán của bộ câu hỏi (ví dụ: "Phép cộng và phép trừ trong phạm vi 100", "Phân số cơ bản và so sánh").
- "short_text" tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG kèm loại bài, KHÔNG dùng cụm chung chung như "Bài kiểm tra Toán", "Bài luyện tập" hay "Quiz".

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "title": "Lớp 1",
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
      "question_grade": 1
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
// difficulty is dropped rather than renamed: the exam schema has no
// per-question difficulty field at all (see the package note on the absent
// LEVEL axis in exam_prompts.go), and an exemplar carrying one would teach
// the model to emit a key nothing reads.
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

// examPracticeBlockVN tells the model what the child just did and how
// the round must respond to it. The two modes are spelled out as
// separate rules rather than left to the model's judgement: "harder"
// without a bound drifts into the next grade, and the grade is the one
// thing a practice round must not move.
func examPracticeBlockVN(b *PracticeBrief, n int) string {
	var sb strings.Builder
	sb.WriteString("BÀI LUYỆN TẬP (dựa trên bài học sinh vừa làm):\n")

	switch b.Mode {
	case enum.PracticeModeRetryWeak:
		fmt.Fprintf(&sb, "- Học sinh vừa làm sai %d câu, thuộc các chủ đề (xếp theo số câu sai giảm dần): %s.\n",
			len(b.Wrong), joinOr(b.WeakTopics, "(không rõ chủ đề)"))
		sb.WriteString("- Các câu đã làm sai:\n")
		for _, w := range b.Wrong {
			fmt.Fprintf(&sb, "  • %s — đáp án đúng: %s; học sinh chọn: %s\n",
				strings.TrimSpace(w.Stem), w.RightAnswer, w.ChildAnswer)
		}
		fmt.Fprintf(&sb, "- YÊU CẦU: giữ nguyên QUY TẮC CẤP LỚP ở trên (kể cả câu dò); cả %d câu tập trung vào các chủ đề đã sai (ưu tiên chủ đề sai nhiều hơn).\n", n)
		sb.WriteString("- Cùng kỹ năng với câu đã sai nhưng KHÔNG lặp lại nguyên văn: đổi số, đổi ngữ cảnh.\n")
		if len(b.StrongTopics) > 0 {
			fmt.Fprintf(&sb, "- Có thể xen 1-2 câu ở chủ đề học sinh đã làm đúng để củng cố: %s.", joinOr(b.StrongTopics, ""))
		}
	default: // ADVANCE
		fmt.Fprintf(&sb, "- Học sinh vừa làm ĐÚNG toàn bộ, ở các chủ đề: %s.\n", joinOr(b.StrongTopics, "(không rõ chủ đề)"))
		fmt.Fprintf(&sb, "- YÊU CẦU: giữ nguyên QUY TẮC CẤP LỚP ở trên (kể cả câu dò); ngoài câu dò, cả %d câu KHÔNG lên cấp lớp cao hơn.\n", n)
		sb.WriteString("- Nhưng phải khó hơn TRONG cấp lớp đó: số lớn hơn trong phạm vi cho phép, nhiều bước tính hơn, có bài toán có lời;\n")
		sb.WriteString("  và/hoặc chuyển sang các chủ đề khác trong chương trình mà học sinh chưa được kiểm tra.")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// joinOr renders a topic list, or the fallback when there is none.
func joinOr(items []string, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return strings.Join(items, ", ")
}

// buildUserExamVN assembles the per-request brief. Order matters: the
// authoritative grade profile comes first so it outranks the schema example
// the system prompt ended on, then the grade rule for individual questions,
// then the softer curriculum context.
func buildUserExamVN(in ExamPromptInput, n int) string {
	out := ""

	if block := gradeProfileBlockByLevel(QuizLanguageVietnamese, GradeLevel(in.Grade)); block != "" {
		out += examVocabulary(block) + "\n\n"
	}

	out += fmt.Sprintf("Hãy tạo bài kiểm tra loại %s gồm %d câu.\n\n", in.ExamType, n)
	out += "QUY TẮC CẤP LỚP CHO TỪNG CÂU:\n" + examProbeBlockVN(in, n) + "\n"

	if in.ExamType == enum.ExamTypePractice && in.Practice != nil {
		out += "\n" + examPracticeBlockVN(in.Practice, n) + "\n"
	}

	if ctx := examContextVN(in); ctx != "" {
		out += "\nThông tin chương trình (chỉ để chọn chủ đề, KHÔNG dùng để tăng hay giảm độ khó):\n" + ctx + "\n"
	}

	out += fmt.Sprintf("\n\"title\" phải đúng bằng: %s", examBandTitle(in.Grade))
	return out
}
