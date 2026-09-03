package bot

import (
	"fmt"
	"strings"
)

// Exercise prompt templates — Vietnamese. JSON keys stay English; only
// the system text and the "short_text" value use Vietnamese. The teacher
// supplies the exercise title separately, so the model only emits
// "short_text" (the auto-generated topic description) here.

const systemExerciseGenerateVNTmpl = `Bạn là trợ lý tạo bài tập toán cho trẻ mẫu giáo và học sinh tiểu học Việt Nam (Mẫu giáo, Lớp 1-5).

Giáo viên đã chọn sẵn chủ đề theo chương và bài học. Hãy tạo CHÍNH XÁC %d câu hỏi trắc nghiệm bám sát đúng bài học đó; nếu có thông tin về cấp lớp / bộ sách, hãy dùng để hiệu chỉnh độ khó.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D.
- Chỉ có một phương án đúng.
- Với câu ARITHMETIC, "question_name" chỉ chứa số và toán tử (+, -, *, /, ^, dấu ngoặc, "?") — không lời văn, không LaTeX, không chữ tiếng Việt. Các loại câu khác tuân theo VISUAL QUESTION RULES ở cuối prompt.
- Dùng phân số ASCII như "1/2", không dùng "½".
- Không lặp lại câu hỏi.
- Tất cả câu hỏi phải nằm trong phạm vi chương + bài học giáo viên đã đặt; không lan sang chủ đề khác kể cả khi vẫn phù hợp với cấp lớp.

QUY TẮC METADATA (BẮT BUỘC ĐỂ CHẤM TỰ ĐỘNG):
- "right_answer" là nhãn (A/B/C/D) của phương án đúng.
- "correct_answer" là GIÁ TRỊ chữ trong "content" của phương án đúng — phải khớp ký tự với "content" tương ứng (vd. "8", "1/2").
- "topic" là mã kỹ năng ngắn bằng tiếng Việt viết thường, snake_case. Ưu tiên dùng các mã: phép cộng trong phạm vi 10, phép trừ trong phạm vi 10, phép nhân 1 chữ số, phép chia 1 chữ số, phép cộng trong phạm vi 100, phép trừ trong phạm vi 100, phép nhân nhiều chữ số, phép chia nhiều chữ số, phân số cơ bản, phân số so sánh, phân số cộng trừ, số thập phân cơ bản, giá trị chỗ, bài toán đố, phép toán hỗn hợp, hình học cơ bản, đo lường, thời gian tiền tệ. Nếu không phù hợp, tạo mã mới ngắn gọn (≤32 ký tự).
- "difficulty" là số nguyên 1..5 phản ánh mức độ thử thách so với bài học đang nhắm tới.

QUY TẮC SHORT_TEXT:
- "short_text" là mô tả ngắn gọn, cụ thể về kỹ năng toán của bộ câu hỏi (ví dụ: "Phép cộng trong phạm vi 10", "Phân số bằng nhau").
- Tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG dùng từ "bài tập" làm mô tả chung.
- "short_text" phải phản ánh đúng các kỹ năng xuất hiện trong "questions"; đây là mô tả nội dung tự sinh (tiêu đề do giáo viên tự đặt riêng).

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "short_text": "Phép cộng trong phạm vi 10",
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
      "right_answer": "A",
      "correct_answer": "8",
      "topic": "phép cộng trong phạm vi 10",
      "difficulty": 1
    }
  ]
}
`

// visualExerciseRulesVN mirrors visualExerciseRulesEN — tuned for
// teacher-scoped exercises: COUNT is gated on BOTH the grade icon policy
// (from the GRADE PROFILE) AND lesson fit. JSON keys stay English; the
// icon whitelist matches the quiz normalizer's geometryIconWhitelist.
const visualExerciseRulesVN = `
VISUAL QUESTION RULES:
- Mỗi câu có "question_type": ARITHMETIC (câu chữ thuần, mặc định) hoặc COUNT (đếm bằng icon). KHÔNG tạo bất kỳ question_type nào khác.
- CHỈ được dùng COUNT/icon khi CẢ HAI điều kiện đúng: (1) chính sách icon trong GRADE PROFILE cho phép — nếu là TẮT, hoặc không có GRADE PROFILE, dùng ARITHMETIC cho mọi câu và TUYỆT ĐỐI không phát sinh emoji hay token [icon:...]; VÀ (2) chương + bài học của giáo viên thực sự về đếm / số lượng / hình cơ bản. Ngược lại, dùng ARITHMETIC cho mọi câu.
- COUNT (chỉ khi cả hai điều kiện trên đúng): "question_name" hiển thị các vật để đếm hoặc cộng bằng icon kèm toán tử (+, "?"). Đáp án là số; giữ "topic" bám theo bài học. Icon có thể là emoji HOẶC token hình (xem ICONS), ví dụ "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?" hoặc "[icon:triangle] [icon:triangle] + [icon:triangle] = ?".

ICONS (chỉ dùng cho COUNT; TUYỆT ĐỐI không dùng trong ARITHMETIC):
- Emoji: với vật đếm được, dùng emoji phổ thông, thân thiện trẻ em, chèn trực tiếp và cách nhau bởi dấu cách, ví dụ 🏓 🍎 ⭐ 🐟 🎈 🚗 🌸 🍓 ⚽ 🐶. Mỗi câu chỉ dùng MỘT loại emoji.
- Hình học: CHỈ dùng token "[icon:NAME]" với NAME thuộc: triangle, square, rectangle, circle, star, diamond, oval, pentagon, hexagon, heart. Lặp token để biểu diễn nhiều hình, ví dụ "[icon:triangle] [icon:triangle] [icon:triangle]". Mỗi câu chỉ dùng MỘT loại hình. TUYỆT ĐỐI không tự đặt tên "[icon:...]" khác — nếu hình không có trong danh sách, hãy dùng emoji.
`

func buildSystemExerciseGenerateVN(n int) string {
	return fmt.Sprintf(systemExerciseGenerateVNTmpl, n, n, n) + visualExerciseRulesVN
}

func userExerciseGenerateVN(in ExercisePromptInput) string {
	scope := fmt.Sprintf("- Chương: %s\n- Bài học: %s",
		strings.TrimSpace(in.ChapterName), strings.TrimSpace(in.LessonName))
	context := buildExerciseContextVN(in)
	gradeBlock := gradeProfileBlock(QuizLanguageVietnamese, in.Grade)

	// GRADE PROFILE quyết định mức độ khó; chương + bài học của giáo viên
	// quyết định chủ đề. Cả hai đều bắt buộc — grade không hạ dưới mức lớp,
	// bài học không cho câu hỏi lệch chủ đề.
	difficultyLine := "Hiệu chỉnh mọi câu hỏi đúng theo GRADE PROFILE ở trên, và giữ trong phạm vi chương + bài học — phạm vi này là bắt buộc."
	header := gradeBlock + "\n\n"
	if gradeBlock == "" {
		difficultyLine = "Chọn độ khó cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5), và giữ mọi câu trong phạm vi chương + bài học — phạm vi này là bắt buộc."
		header = ""
	}

	if context == "" {
		return fmt.Sprintf(`%sHãy tạo bài tập trên lớp theo chủ đề giáo viên đã đặt:
%s

%s`, header, scope, difficultyLine)
	}
	return fmt.Sprintf(`%sHãy tạo bài tập trên lớp theo chủ đề giáo viên đã đặt:
%s

Thông tin bổ sung:
%s

%s Hướng dẫn của giáo viên (nếu có) chỉ là gợi ý, không phải cớ để lan sang chủ đề khác.`, header, scope, context, difficultyLine)
}

// Grading prompt — Vietnamese. Mirrors the quiz-grading output shape so
// the existing parseGradedQuiz helper handles the response unchanged.
const systemExerciseGradeVN = `Bạn là trợ lý chấm bài tập trắc nghiệm toán cho trẻ mẫu giáo và học sinh tiểu học Việt Nam.

Đầu vào gồm:
- "questions": danh sách câu hỏi gốc kèm "right_answer" (nhãn đúng A/B/C/D).
- "answers": danh sách câu trả lời học sinh đã chọn theo "question_number" và "label".

NHIỆM VỤ:
- So khớp từng câu trả lời với "right_answer" của câu hỏi tương ứng để xác định đúng/sai.
- Câu hỏi không có câu trả lời tương ứng được tính là SAI.
- Tính tổng số câu (total_questions), số câu đúng (correct_number) và phần trăm điểm (score_percentage, làm tròn xuống).
- Viết "ai_review" ngắn gọn (2-4 câu) bằng tiếng Việt: ghi nhận điểm mạnh, lỗi sai phổ biến, và một gợi ý ôn tập cụ thể bám sát chương + bài học.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object đúng cấu trúc bên dưới, không lời dẫn, không markdown.

CẤU TRÚC:
{
  "total_questions": 10,
  "correct_number": 8,
  "score_percentage": 80,
  "ai_review": "..."
}
`

func userExerciseGradeVN(in ExercisePromptInput) string {
	return fmt.Sprintf(`Câu hỏi gốc (JSON):
%s

Câu trả lời học sinh (JSON):
%s

Hãy chấm bài theo cấu trúc đầu ra đã quy định.`, strings.TrimSpace(in.Questions), strings.TrimSpace(in.Answers))
}
