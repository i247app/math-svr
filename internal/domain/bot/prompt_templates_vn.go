package bot

import "fmt"

// Vietnamese prompt templates. ai_review is emitted in Vietnamese; all
// JSON keys remain English so persistence and grading code can use a
// single struct shape regardless of QuizLanguage.

const systemGenerateVN = `Bạn là trợ lý tạo bài kiểm tra toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Hãy tạo CHÍNH XÁC 5 câu hỏi trắc nghiệm phù hợp với lớp, học kỳ và chương trình học được yêu cầu.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D.
- Chỉ có một phương án đúng.
- "question_name" chỉ chứa số và toán tử (+, -, *, /, ^, dấu ngoặc, "?") — không lời văn, không LaTeX, không hình ảnh, không chữ tiếng Việt.
- Dùng phân số ASCII như "1/2", không dùng "½".
- Không lặp lại câu hỏi.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON array theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- Mảng phải có đúng 5 phần tử, "question_number" từ 1..5 theo thứ tự.

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

const systemReinforceVN = `Bạn là trợ lý tạo bài kiểm tra toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Bạn sẽ nhận được bài kiểm tra trước, câu trả lời của học sinh và nhận xét AI về kết quả đó. Hãy tạo MỘT bài kiểm tra MỚI gồm ĐÚNG 5 câu trắc nghiệm tập trung vào các dạng bài học sinh làm sai hoặc còn yếu, đồng thời giữ độ khó phù hợp với lớp, học kỳ và chương trình.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D; chỉ một đáp án đúng.
- "question_name" chỉ chứa số và toán tử — không lời văn, không LaTeX, không chữ tiếng Việt.
- Dùng phân số ASCII như "1/2".
- Không sao chép nguyên văn câu hỏi cũ; tạo biến thể nhắm vào cùng kỹ năng.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON array theo cấu trúc bên dưới. Không lời dẫn, không khung markdown.
- Mảng phải có đúng 5 phần tử, "question_number" từ 1..5 theo thứ tự.

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
	guidance := "Điều chỉnh độ khó phù hợp với lớp và học kỳ."
	if typ == QuizTypePractice {
		guidance = "Điều chỉnh độ khó theo mốc học kỳ; đây là vòng luyện tập thông thường."
	}
	return fmt.Sprintf(`Hãy tạo bài kiểm tra %s cho học sinh với thông tin sau:
- Lớp: %s
- Học kỳ: %s
- Chương trình học: %s

%s`, typ, in.Grade, in.Semester, in.Program, guidance)
}

func userReinforceVN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Hãy tạo bài kiểm tra %s củng cố cho học sinh với thông tin sau:
- Lớp: %s
- Học kỳ: %s
- Chương trình học: %s

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

Tập trung vào điểm yếu trong nhận xét trước, đồng thời giữ độ khó phù hợp với lớp và học kỳ đang cấu hình.`, typ, in.Grade, in.Semester, in.Program,
		in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview)
}

func userGradeVN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Hãy chấm bài kiểm tra %s sau đây.

Câu hỏi (JSON): %s
Câu trả lời của học sinh (JSON): %s`, typ, in.Questions, in.Answers)
}

func userGradeReinforceVN(typ QuizType, in QuizPromptInput) string {
	return fmt.Sprintf(`Hãy chấm bài kiểm tra %s củng cố sau đây.

Câu hỏi (JSON): %s
Câu trả lời của học sinh (JSON): %s
Cấp lớp đang được cấu hình: %s`, typ, in.Questions, in.Answers, in.CurrentGrade)
}
