package bot

import (
	"fmt"
	"strings"
)

// Vietnamese prompt templates. ai_review is emitted in Vietnamese; all
// JSON keys remain English so persistence and grading code can use a
// single struct shape regardless of QuizLanguage.

// systemGenerateVNTmpl mirrors the EN template: three %d slots all
// filled with the same target question count.
const systemGenerateVNTmpl = `Bạn là trợ lý tạo bài kiểm tra toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Hãy tạo CHÍNH XÁC %d câu hỏi trắc nghiệm phù hợp với thông tin học vấn người dùng cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D.
- Chỉ có một phương án đúng.
- "question_name" chỉ chứa số và toán tử (+, -, *, /, ^, dấu ngoặc, "?") — không lời văn, không LaTeX, không hình ảnh, không chữ tiếng Việt.
- Dùng phân số ASCII như "1/2", không dùng "½".
- Không lặp lại câu hỏi.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON array theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- Mảng phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
[
  {
    "question_number": 1,
    "question_name": "5 + 3 = ?",
    "answers": [
      {"label": "A", "content": "8"},
      {"label": "B", "content": "9"},
      {"label": "C", "content": "10"},
      {"label": "D", "content": "7"}
    ],
    "right_answer": "A"
  }
]
`

const systemReinforceVNTmpl = `Bạn là trợ lý tạo bài kiểm tra toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Bạn sẽ nhận được bài kiểm tra trước, câu trả lời của học sinh và nhận xét AI về kết quả đó. Hãy tạo MỘT bài kiểm tra MỚI gồm ĐÚNG %d câu trắc nghiệm tập trung vào các dạng bài học sinh làm sai hoặc còn yếu. Dùng thông tin học vấn người dùng cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D; chỉ một đáp án đúng.
- "question_name" chỉ chứa số và toán tử — không lời văn, không LaTeX, không chữ tiếng Việt.
- Dùng phân số ASCII như "1/2".
- Không sao chép nguyên văn câu hỏi cũ; tạo biến thể nhắm vào cùng kỹ năng.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON array theo cấu trúc bên dưới. Không lời dẫn, không khung markdown.
- Mảng phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
[
  {
    "question_number": 1,
    "question_name": "5 + 3 = ?",
    "answers": [
      {"label": "A", "content": "8"},
      {"label": "B", "content": "9"},
      {"label": "C", "content": "10"},
      {"label": "D", "content": "7"}
    ],
    "right_answer": "A"
  }
]
`

func buildSystemGenerateVN(n int) string {
	return fmt.Sprintf(systemGenerateVNTmpl, n, n, n)
}

func buildSystemReinforceVN(n int) string {
	return fmt.Sprintf(systemReinforceVNTmpl, n, n, n)
}

const systemGradeAssessmentVN = `Bạn là trợ lý chấm điểm bài kiểm tra toán cho học sinh tiểu học Việt Nam.

Bạn sẽ nhận được danh sách câu hỏi (JSON array) và câu trả lời của học sinh (JSON object có khoá là question_number). Hãy chấm điểm rồi dự đoán cấp lớp phù hợp nhất cho học sinh.

QUY TẮC CHẤM:
- Đối chiếu từng câu trả lời theo "question_number" với "right_answer" của câu hỏi.
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" viết bằng tiếng Việt, dài tối đa 200 ký tự, nêu một điểm mạnh và một điểm cần cải thiện cụ thể. Không xuống dòng.
- "ai_detect_grade" phải là một trong: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Căn cứ vào độ khó câu hỏi VÀ mức chính xác quan sát được, không chỉ dựa vào tỷ lệ đúng.

QUY TẮC ĐẦU RA:
- CHỈ trả về MỘT JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không xuống dòng trong giá trị chuỗi.

CẤU TRÚC:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Phép cộng cơ bản tốt; cần luyện thêm phép trừ có nhớ.",
  "ai_detect_grade": "Grade 3"
}
`

const systemGradePracticeVN = `Bạn là trợ lý chấm điểm bài kiểm tra toán cho học sinh tiểu học Việt Nam.

Bạn sẽ nhận được danh sách câu hỏi (JSON array) và câu trả lời của học sinh (JSON object có khoá là question_number). Hãy chấm điểm và trả về nhận xét ngắn gọn.

QUY TẮC CHẤM:
- Đối chiếu từng câu trả lời theo "question_number" với "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" viết bằng tiếng Việt, dài tối đa 200 ký tự, nêu một điểm mạnh và một điểm cần cải thiện cụ thể. Không xuống dòng.

QUY TẮC ĐẦU RA:
- CHỈ trả về MỘT JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không xuống dòng trong giá trị chuỗi.

CẤU TRÚC:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Phép cộng cơ bản tốt; cần luyện thêm phép trừ có nhớ."
}
`

const systemGradeReinforceAssessmentVN = `Bạn là trợ lý chấm điểm bài kiểm tra toán cho học sinh tiểu học Việt Nam.

Bạn sẽ nhận được câu hỏi của bài kiểm tra củng cố (JSON array), câu trả lời của học sinh và cấp lớp đang được cấu hình. Hãy chấm điểm rồi dự đoán xem học sinh đã tiến bộ đủ để lên lớp, giữ nguyên hay cần lùi lại.

QUY TẮC CHẤM:
- Đối chiếu từng câu trả lời theo "question_number" với "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" viết bằng tiếng Việt, tối đa 200 ký tự; có so sánh tiến bộ so với bài kiểm tra trước. Không xuống dòng.
- "ai_detect_grade" phải là một trong: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Lấy cấp lớp hiện tại làm mốc; chỉ điều chỉnh khi kết quả thực sự rõ ràng.

QUY TẮC ĐẦU RA:
- CHỈ trả về MỘT JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không xuống dòng trong giá trị chuỗi.

CẤU TRÚC:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Phép trừ tốt hơn; cần luyện thêm cộng nhiều chữ số.",
  "ai_detect_grade": "Grade 3"
}
`

const systemGradeReinforcePracticeVN = `Bạn là trợ lý chấm điểm bài kiểm tra toán cho học sinh tiểu học Việt Nam.

Bạn sẽ nhận được câu hỏi của bài luyện tập củng cố (JSON array), câu trả lời của học sinh và cấp lớp hiện tại. Hãy chấm điểm và đưa nhận xét ngắn về việc đợt luyện tập có thu hẹp được khoảng cách hay không.

QUY TẮC CHẤM:
- Đối chiếu từng câu trả lời theo "question_number" với "right_answer".
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" viết bằng tiếng Việt, tối đa 200 ký tự; có so sánh với đợt luyện trước. Không xuống dòng.

QUY TẮC ĐẦU RA:
- CHỈ trả về MỘT JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không xuống dòng trong giá trị chuỗi.

CẤU TRÚC:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Phép trừ tốt hơn; cần luyện thêm cộng nhiều chữ số."
}
`

func userGenerateVN(typ QuizType, in QuizPromptInput) string {
	guidance := "Điều chỉnh độ khó theo thông tin đã cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học."
	if typ == QuizTypePractice {
		guidance = "Điều chỉnh độ khó theo mốc học kỳ khi có; nếu không, hãy chọn bộ luyện tập cân bằng ở trình độ tiểu học."
	}
	context := buildCurriculumContextVN(in)
	if context == "" {
		return fmt.Sprintf("Hãy tạo bài kiểm tra %s.\n\nKhông có thông tin học vấn cụ thể — hãy dùng bộ câu hỏi toán cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).\n\n%s", typ, guidance)
	}
	return fmt.Sprintf("Hãy tạo bài kiểm tra %s cho học sinh với thông tin sau:\n%s\n\n%s", typ, context, guidance)
}

func userReinforceVN(typ QuizType, in QuizPromptInput) string {
	closing := "Tập trung vào điểm yếu trong nhận xét trước, đồng thời giữ độ khó phù hợp với thông tin nêu trên; nếu không có thông tin, dùng mức cân bằng cho học sinh tiểu học."
	context := buildCurriculumContextVN(in)
	if context == "" {
		return fmt.Sprintf(`Hãy tạo bài kiểm tra %s củng cố.

Không có thông tin học vấn cụ thể — hãy dùng bộ câu hỏi toán cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

%s`, typ, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
	}
	return fmt.Sprintf(`Hãy tạo bài kiểm tra %s củng cố cho học sinh với thông tin sau:
%s

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

%s`, typ, context, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
}

func userGradeVN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Hãy chấm bài kiểm tra %s sau đây.

Câu hỏi (JSON): %s
Câu trả lời của học sinh (JSON): %s`, typ, in.Questions, in.Answers)
}

func userGradeReinforceVN(typ QuizType, in QuizPromptInput) string {
	currentGrade := strings.TrimSpace(in.CurrentGrade)
	if currentGrade == "" {
		currentGrade = "chưa xác định (không có cấp lớp được cấu hình)"
	}
	return fmt.Sprintf(`Hãy chấm bài kiểm tra %s củng cố sau đây.

Câu hỏi (JSON): %s
Câu trả lời của học sinh (JSON): %s
Cấp lớp đang được cấu hình: %s`, typ, in.Questions, in.Answers, currentGrade)
}

// buildCurriculumContextVN mirrors the EN helper: emits only the lines
// whose value is non-empty, returns "" when nothing was provided.
func buildCurriculumContextVN(in QuizPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Grade); v != "" {
		fmt.Fprintf(&b, "- Lớp: %s\n", v)
	}
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Học kỳ: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Chương trình học: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}
