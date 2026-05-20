package bot

var PromptForGenerateQuizPracticeVN = `
Tạo 5 câu hỏi toán trắc nghiệm cho cấp Tiểu học Việt Nam (%s-%s) với độ khó theo 4 học kỳ và lập trình viên là %s.

QUAN TRỌNG: Chỉ trả về duy nhất một JSON array hợp lệ. Chỉ hiển thị đề bài. KHÔNG thêm văn bản, giải thích hoặc markdown.
Sử dụng cấu trúc mẫu câu hỏi trắc nghiệm sau:
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
QUY TẮC ĐỊNH DẠNG NGHIÊM NGẶT: KHÔNG dùng LATEX. Chỉ dùng văn bản đơn giản (ví dụ: "1/2", "x", "^"). "question_name" chỉ được chứa số và toán tử. Chính xác 5 câu hỏi và mỗi câu có 4 đáp án (A, B, C, D).
`

var PromptForSubmitQuizPracticeAnswerVN = `
Bạn là một trợ lý chấm điểm bài kiểm tra toán.
Dựa trên câu trả lời của người dùng và thông tin câu hỏi bên dưới, hãy phân tích và tính toán kết quả bài kiểm tra.

+ Thông tin câu hỏi: %s
+ Câu trả lời của người dùng: %s

Dựa trên câu trả lời của người dùng cho bài kiểm tra, hãy cung cấp các thông tin sau trong một JSON object:
- total_questions: Tổng số câu hỏi trong bài kiểm tra
- correct_number: Số câu trả lời đúng
- score_percentage: Điểm phần trăm của người dùng (correct_number / total_questions * 100)
- ai_review: Nhận xét ngắn gọn về kết quả làm bài, nêu điểm mạnh, điểm cần cải thiện và các dạng bài cần luyện thêm.

Chỉ trả về JSON object theo cấu trúc sau:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Làm rất tốt! Bạn có nền tảng toán cơ bản khá vững, nhưng kỹ năng phép trừ vẫn chưa đủ tốt. Hãy luyện tập thêm các bài toán phép trừ để cải thiện kỹ năng."
}

Yêu cầu:
- Chỉ trả về JSON object (không xuống dòng (\n), không cần khoảng trắng), không thêm bất kỳ nội dung nào khác
- Đảm bảo JSON được format hợp lệ với dấu ngoặc kép và dấu phẩy chính xác
- "ai_review" phải ngắn hơn 200 ký tự
`

var PromptForGenerateReinforceQuizPracticeVN = `
Tạo một bài kiểm tra toán mới gồm 5 câu hỏi trắc nghiệm dựa trên kết quả làm bài và đánh giá trước đó của người dùng dưới đây:

+ Thông tin câu hỏi: %s
+ Câu trả lời trước đó của người dùng: %s
+ Đánh giá AI về kết quả làm bài của người dùng: %s

QUAN TRỌNG: Chỉ trả về duy nhất một JSON array hợp lệ. Chỉ hiển thị đề bài. KHÔNG thêm văn bản, giải thích hoặc markdown.
Sử dụng cấu trúc mẫu câu hỏi trắc nghiệm sau:
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
QUY TẮC ĐỊNH DẠNG NGHIÊM NGẶT: KHÔNG dùng LATEX. Chỉ dùng văn bản đơn giản (ví dụ: "1/2", "x", "^"). "question_name" chỉ được chứa số và toán tử. Chính xác 5 câu hỏi và mỗi câu có 4 đáp án (A, B, C, D).
`
