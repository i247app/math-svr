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

QUY TẮC METADATA (BẮT BUỘC ĐỂ CHẤM TỰ ĐỘNG):
- "right_answer" là nhãn (A/B/C/D) của phương án đúng.
- "correct_answer" là GIÁ TRỊ chữ trong "content" của phương án đúng — phải khớp ký tự với "content" tương ứng (vd. "8", "1/2").
- "topic" là một mã kỹ năng ngắn bằng tiếng Anh viết thường, dùng dấu gạch dưới (snake_case). Chọn trong danh sách gợi ý: phép cộng trong phạm vi 100, phép trừ trong phạm vi 100, phép cộng có nhớ, phép trừ có nhớ, phép nhân với số có một chữ số, phép nhân với số có nhiều chữ số, phép chia cho số có một chữ số, phép chia cho số có nhiều chữ số, phân số cơ bản, phân số so sánh, cộng trừ phân số, số thập phân cơ bản, giá trị theo vị trí, bài toán có lời văn, phép tính hỗn hợp, hình học cơ bản, phép đo, thời gian và tiền tệ. Nếu thực sự không phù hợp, tạo mã mới ngắn gọn (≤32 ký tự).
- "difficulty" là số nguyên 1..5 (1 dễ nhất, 5 khó nhất) phản ánh mức độ thử thách so với lớp đang nhắm tới.

QUY TẮC TIÊU ĐỀ:
- "title" là tiêu đề ngắn gọn, cụ thể, mô tả ĐÚNG chủ đề toán của bộ câu hỏi (ví dụ: "Phép cộng và phép trừ trong phạm vi 100", "Phân số cơ bản và so sánh", "Phép nhân với số có một chữ số").
- Tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG kèm loại bài (ASSESSMENT/PRACTICE), KHÔNG dùng tiêu đề chung chung như "Bài kiểm tra Toán", "Bài luyện tập" hay "Quiz".
- Mỗi lần sinh hãy chọn tiêu đề phản ánh chính xác các kỹ năng xuất hiện trong "questions" để không lặp lại giữa các bài.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "title": "Phép cộng và phép trừ trong phạm vi 100",
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
      "topic": "phép cộng trong phạm vi 100",
      "difficulty": 1
    }
  ]
}
`

const systemReinforceVNTmpl = `Bạn là trợ lý tạo bài kiểm tra toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Bạn sẽ nhận được bài kiểm tra trước, câu trả lời của học sinh và nhận xét AI về kết quả đó. Hãy tạo MỘT bài kiểm tra MỚI gồm ĐÚNG %d câu trắc nghiệm tập trung vào các dạng bài học sinh làm sai hoặc còn yếu. Dùng thông tin học vấn người dùng cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D; chỉ một đáp án đúng.
- "question_name" chỉ chứa số và toán tử — không lời văn, không LaTeX, không chữ tiếng Việt.
- Dùng phân số ASCII như "1/2".
- Không sao chép nguyên văn câu hỏi cũ; tạo biến thể nhắm vào cùng kỹ năng.

QUY TẮC METADATA (BẮT BUỘC ĐỂ CHẤM TỰ ĐỘNG):
- "right_answer" là nhãn (A/B/C/D) của phương án đúng.
- "correct_answer" là GIÁ TRỊ chữ trong "content" của phương án đúng — phải khớp ký tự với "content" tương ứng.
- "topic" là mã kỹ năng ngắn bằng tiếng Việt viết thường, snake_case (vd. phép cộng trong phạm vi 100, phép trừ trong phạm vi 100). Tận dụng lại các mã đã có trong danh sách gợi ý của prompt sinh; chỉ tạo mã mới khi thật sự cần.
- "difficulty" là số nguyên 1..5 phản ánh độ khó so với lớp đang nhắm tới.

QUY TẮC TIÊU ĐỀ:
- "title" là tiêu đề ngắn gọn, cụ thể, mô tả ĐÚNG chủ đề được củng cố (ví dụ: "Củng cố phép trừ có nhớ", "Ôn lại phân số bằng nhau").
- Tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG kèm loại bài (ASSESSMENT/PRACTICE), KHÔNG dùng tiêu đề chung chung như "Bài củng cố", "Bài ôn tập" hay "Quiz".
- Tiêu đề phải phản ánh đúng kỹ năng được nhắm tới trong "questions" của bài mới — không sao chép tiêu đề bài cũ.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "title": "Củng cố phép trừ có nhớ",
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
      "topic": "phép cộng trong phạm vi 100",
      "difficulty": 3
    }
  ]
}
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

// learningIntentVN nêu type_of_quiz bằng tiếng Việt để gắn vào prompt
// người dùng. GENERAL hướng tới kiến thức mới theo chương trình;
// REINFORCEMENT tập trung củng cố / ôn lại các kỹ năng học sinh còn yếu.
func learningIntentVN(t QuizTypeOfQuiz) string {
	if t == QuizTypeOfQuizReinforcement {
		return "củng cố (ôn tập các kỹ năng học sinh đã làm sai trước đó)"
	}
	return "chung (giới thiệu hoặc luyện tập kiến thức mới theo chương trình)"
}

func userGenerateVN(purpose QuizPurpose, in QuizPromptInput) string {
	guidance := "Điều chỉnh độ khó theo thông tin đã cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học."
	if purpose == QuizPurposePractice {
		guidance = "Điều chỉnh độ khó theo mốc học kỳ khi có; nếu không, hãy chọn bộ luyện tập cân bằng ở trình độ tiểu học."
	}
	intent := learningIntentVN(in.TypeOfQuiz)
	context := buildCurriculumContextVN(in)
	if context == "" {
		return fmt.Sprintf("Hãy tạo bài kiểm tra %s.\n\nMục tiêu học tập: %s.\n\nKhông có thông tin học vấn cụ thể — hãy dùng bộ câu hỏi toán cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).\n\n%s", purpose, intent, guidance)
	}
	return fmt.Sprintf("Hãy tạo bài kiểm tra %s cho học sinh với thông tin sau:\n%s\n\nMục tiêu học tập: %s.\n\n%s", purpose, context, intent, guidance)
}

func userReinforceVN(purpose QuizPurpose, in QuizPromptInput) string {
	closing := "Tập trung vào điểm yếu trong nhận xét trước, đồng thời giữ độ khó phù hợp với thông tin nêu trên; nếu không có thông tin, dùng mức cân bằng cho học sinh tiểu học."
	intent := learningIntentVN(QuizTypeOfQuizReinforcement)
	context := buildCurriculumContextVN(in)
	if context == "" {
		return fmt.Sprintf(`Hãy tạo bài kiểm tra %s củng cố.

Mục tiêu học tập: %s.

Không có thông tin học vấn cụ thể — hãy dùng bộ câu hỏi toán cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

%s`, purpose, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
	}
	return fmt.Sprintf(`Hãy tạo bài kiểm tra %s củng cố cho học sinh với thông tin sau:
%s

Mục tiêu học tập: %s.

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

%s`, purpose, context, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
}

func userGradeVN(purpose QuizPurpose, in QuizPromptInput) string {
	return fmt.Sprintf(`Hãy chấm bài kiểm tra %s sau đây (mục tiêu học tập: %s).

Câu hỏi (JSON): %s
Câu trả lời của học sinh (JSON): %s`, purpose, learningIntentVN(in.TypeOfQuiz), in.Questions, in.Answers)
}

func userGradeReinforceVN(purpose QuizPurpose, in QuizPromptInput) string {
	currentGrade := strings.TrimSpace(in.CurrentGrade)
	if currentGrade == "" {
		currentGrade = "chưa xác định (không có cấp lớp được cấu hình)"
	}
	return fmt.Sprintf(`Hãy chấm bài kiểm tra %s củng cố sau đây (mục tiêu học tập: %s).

Câu hỏi (JSON): %s
Câu trả lời của học sinh (JSON): %s
Cấp lớp đang được cấu hình: %s`, purpose, learningIntentVN(QuizTypeOfQuizReinforcement), in.Questions, in.Answers, currentGrade)
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
	if block := buildChaptersBlockVN(in.ChapterDescriptions); block != "" {
		b.WriteString(block)
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildChaptersBlockVN mirrors buildChaptersBlockEN — same shape, VN
// wording. The header line embeds the "prioritise these topics"
// instruction so the chapter list is not just decorative.
func buildChaptersBlockVN(chapters []string) string {
	cleaned := cleanChapterLabels(chapters)
	if len(cleaned) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("- Các chương cần tập trung (ưu tiên ra câu hỏi bám sát các chủ đề sau):\n")
	for _, label := range cleaned {
		fmt.Fprintf(&b, "  • %s\n", label)
	}
	return b.String()
}
