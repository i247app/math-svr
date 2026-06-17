package bot

import (
	"fmt"
	"strings"
)

// Exercise prompt templates — Vietnamese. JSON keys stay English; only
// the system text and the "short_text" value use Vietnamese. The teacher
// supplies the exercise title separately, so the model only emits
// "short_text" (the auto-generated topic description) here.

const systemExerciseGenerateVNTmpl = `Bạn là trợ lý tạo bài tập toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Giáo viên đã chọn sẵn chủ đề theo chương và bài học. Hãy tạo CHÍNH XÁC %d câu hỏi trắc nghiệm bám sát đúng bài học đó; nếu có thông tin về cấp lớp / bộ sách, hãy dùng để hiệu chỉnh độ khó.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D.
- Chỉ có một phương án đúng.
- "question_name" chỉ chứa số và toán tử (+, -, *, /, ^, dấu ngoặc, "?") — không lời văn, không LaTeX, không hình ảnh, không chữ tiếng Việt.
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

func buildSystemExerciseGenerateVN(n int) string {
	return fmt.Sprintf(systemExerciseGenerateVNTmpl, n, n, n)
}

func userExerciseGenerateVN(in ExercisePromptInput) string {
	scope := fmt.Sprintf("- Chương: %s\n- Bài học: %s",
		strings.TrimSpace(in.ChapterName), strings.TrimSpace(in.LessonName))
	context := buildExerciseContextVN(in)
	if context == "" {
		return fmt.Sprintf(`Hãy tạo bài tập trên lớp theo chủ đề giáo viên đã đặt:
%s

Không có thêm thông tin về cấp lớp / bộ sách — hãy chọn độ khó cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).

Mọi câu hỏi đều phải nằm trong phạm vi chương + bài học nêu trên.`, scope)
	}
	return fmt.Sprintf(`Hãy tạo bài tập trên lớp theo chủ đề giáo viên đã đặt:
%s

Thông tin bổ sung:
%s

Dùng các thông tin bổ sung phía trên để điều chỉnh độ khó, trọng tâm và cách diễn đạt — KHÔNG ghi đè phạm vi chương + bài học vốn bắt buộc. Hướng dẫn của giáo viên (nếu có) chỉ là gợi ý, không phải cớ để lan sang chủ đề khác.`, scope, context)
}

// Grading prompt — Vietnamese. Mirrors the quiz-grading output shape so
// the existing parseGradedQuiz helper handles the response unchanged.
const systemExerciseGradeVN = `Bạn là trợ lý chấm bài tập trắc nghiệm toán cho học sinh tiểu học Việt Nam.

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
