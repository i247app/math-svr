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
- Với câu ARITHMETIC, "question_name" chỉ chứa số và toán tử (+, -, *, /, ^, dấu ngoặc, "?") — không lời văn, không LaTeX, không chữ tiếng Việt. Các loại câu khác tuân theo VISUAL QUESTION RULES ở cuối prompt.
- Dùng phân số ASCII như "1/2", không dùng "½".
- Không lặp lại câu hỏi.

QUY TẮC METADATA (BẮT BUỘC ĐỂ CHẤM TỰ ĐỘNG):
- "right_answer" là nhãn (A/B/C/D) của phương án đúng.
- "correct_answer" là GIÁ TRỊ chữ trong "content" của phương án đúng — phải khớp ký tự với "content" tương ứng (vd. "8", "1/2").
- "topic" là một mã kỹ năng ngắn bằng tiếng Anh viết thường, dùng dấu gạch dưới (snake_case). Chọn trong danh sách gợi ý: phép cộng trong phạm vi 100, phép trừ trong phạm vi 100, phép cộng có nhớ, phép trừ có nhớ, phép nhân với số có một chữ số, phép nhân với số có nhiều chữ số, phép chia cho số có một chữ số, phép chia cho số có nhiều chữ số, phân số cơ bản, phân số so sánh, cộng trừ phân số, số thập phân cơ bản, giá trị theo vị trí, bài toán có lời văn, phép tính hỗn hợp, hình học cơ bản, phép đo, thời gian và tiền tệ. Nếu thực sự không phù hợp, tạo mã mới ngắn gọn (≤32 ký tự).
- "difficulty" là số nguyên 1..5 (1 dễ nhất, 5 khó nhất) phản ánh mức độ thử thách so với lớp đang nhắm tới.

QUY TẮC TITLE & SHORT_TEXT:
- "title" là nhãn cấp lớp/cấp độ của bài, theo định dạng "Grade <N> - Level <M>" (ví dụ: "Grade 1 - Level 1"). Lấy <N> từ cấp lớp trong thông tin học vấn người dùng cung cấp; chọn <M> (1..5) dựa trên độ khó tổng thể của các câu hỏi bạn tạo (1 = dễ nhất). KHÔNG đặt chủ đề toán vào "title".
- "short_text" là tiêu đề ngắn gọn, cụ thể, mô tả ĐÚNG chủ đề toán của bộ câu hỏi (ví dụ: "Phép cộng và phép trừ trong phạm vi 100", "Phân số cơ bản và so sánh", "Phép nhân với số có một chữ số").
- "short_text" tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG kèm loại bài (ASSESSMENT/PRACTICE), KHÔNG dùng cụm chung chung như "Bài kiểm tra Toán", "Bài luyện tập" hay "Quiz".
- Mỗi lần sinh hãy chọn "short_text" phản ánh chính xác các kỹ năng xuất hiện trong "questions" để không lặp lại giữa các bài.

QUY TẮC ASSESSMENT_GRADE:
- "assessment_grade" là cấp lớp mà bài kiểm tra này đánh giá, PHẢI là một trong: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Căn cứ vào cấp lớp trong thông tin học vấn và độ khó tổng thể của các câu hỏi bạn tạo.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không bình luận thêm.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "title": "Grade 1 - Level 1",
  "short_text": "Phép cộng và phép trừ trong phạm vi ...",
  "assessment_grade": "Grade 1",
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
      "topic": "phép cộng trong phạm vi ...",
      "difficulty": 1
    }
  ]
}
`

const systemReinforceVNTmpl = `Bạn là trợ lý tạo bài kiểm tra toán cho học sinh tiểu học Việt Nam (Lớp 1-5).

Bạn sẽ nhận được bài kiểm tra trước, câu trả lời của học sinh và nhận xét AI về kết quả đó. Hãy tạo MỘT bài kiểm tra MỚI gồm ĐÚNG %d câu trắc nghiệm tập trung vào các dạng bài học sinh làm sai hoặc còn yếu. Dùng thông tin học vấn người dùng cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học.

QUY TẮC NỘI DUNG:
- Mỗi câu có ĐÚNG 4 phương án A, B, C, D; chỉ một đáp án đúng.
- Với câu ARITHMETIC, "question_name" chỉ chứa số và toán tử — không lời văn, không LaTeX, không chữ tiếng Việt. Các loại câu khác tuân theo VISUAL QUESTION RULES ở cuối prompt.
- Dùng phân số ASCII như "1/2".
- Không sao chép nguyên văn câu hỏi cũ; tạo biến thể nhắm vào cùng kỹ năng.

QUY TẮC METADATA (BẮT BUỘC ĐỂ CHẤM TỰ ĐỘNG):
- "right_answer" là nhãn (A/B/C/D) của phương án đúng.
- "correct_answer" là GIÁ TRỊ chữ trong "content" của phương án đúng — phải khớp ký tự với "content" tương ứng.
- "topic" là mã kỹ năng ngắn bằng tiếng Việt viết thường, snake_case (vd. phép cộng trong phạm vi 100, phép trừ trong phạm vi 100). Tận dụng lại các mã đã có trong danh sách gợi ý của prompt sinh; chỉ tạo mã mới khi thật sự cần.
- "difficulty" là số nguyên 1..5 phản ánh độ khó so với lớp đang nhắm tới.

QUY TẮC TITLE & SHORT_TEXT:
- "title" là nhãn cấp lớp/cấp độ của bài, theo định dạng "Grade <N> - Level <M>" (ví dụ: "Grade 1 - Level 1"). Lấy <N> từ cấp lớp trong thông tin học vấn người dùng cung cấp; chọn <M> (1..5) dựa trên độ khó tổng thể của các câu hỏi MỚI (1 = dễ nhất). KHÔNG đặt chủ đề toán vào "title".
- "short_text" là tiêu đề ngắn gọn, cụ thể, mô tả ĐÚNG chủ đề được củng cố (ví dụ: "Củng cố phép trừ có nhớ", "Ôn lại phân số bằng nhau").
- "short_text" tối đa 80 ký tự, viết bằng tiếng Việt, KHÔNG kèm cấp lớp, KHÔNG kèm loại bài (ASSESSMENT/PRACTICE), KHÔNG dùng cụm chung chung như "Bài củng cố", "Bài ôn tập" hay "Quiz".
- "short_text" phải phản ánh đúng kỹ năng được nhắm tới trong "questions" của bài mới — không sao chép short_text của bài cũ.

QUY TẮC ASSESSMENT_GRADE:
- "assessment_grade" là cấp lớp mà bài kiểm tra này đánh giá, PHẢI là một trong: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Căn cứ vào cấp lớp trong thông tin học vấn và độ khó tổng thể của các câu hỏi MỚI.

QUY TẮC ĐẦU RA:
- CHỈ trả về JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown.
- "questions" phải có đúng %d phần tử, "question_number" từ 1..%d theo thứ tự.

CẤU TRÚC:
{
  "title": "Grade 1 - Level 2",
  "short_text": "Củng cố phép trừ có nhớ",
  "assessment_grade": "Grade 1",
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
      "topic": "phép cộng trong phạm vi 100",
      "difficulty": 3
    }
  ]
}
`

// visualQuestionRulesVN mirrors visualQuestionRulesEN. JSON keys stay
// English; only answer text (e.g. shape names) is Vietnamese. Single
// source of truth for the icon contract, appended to both VN prompts.
// const visualQuestionRulesVN = `
// VISUAL QUESTION RULES:
// - Mỗi câu có "question_type": một trong ARITHMETIC, COUNT, PICK_BY_ICON, IDENTIFY_SHAPE. Mặc định là ARITHMETIC (câu chữ thuần).
// - Cân bằng theo LỚP: Lớp 1-2 NÊN xen COUNT / PICK_BY_ICON / IDENTIFY_SHAPE để tư duy trực quan; Lớp 3 dùng ít; Lớp 4-5 gần như toàn ARITHMETIC.
// - COUNT: "question_name" hiển thị các vật để đếm hoặc cộng bằng icon kèm toán tử (+, "?"), ví dụ "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?". Đáp án là số; "topic" là "phép đếm".
// - PICK_BY_ICON: "question_name" là câu hỏi ngắn chọn phương án đúng (ví dụ "Đáp án nào có 3 hình tam giác?"); mỗi "content" của đáp án là một nhóm icon; "topic" là "hình học cơ bản".
// - IDENTIFY_SHAPE: "question_name" là ĐÚNG MỘT token hình (ví dụ "[icon:triangle]"); đáp án là tên hình bằng tiếng Việt; "topic" là "hình học cơ bản".

// ICONS (chỉ dùng cho các loại visual trên; TUYỆT ĐỐI không dùng trong ARITHMETIC):
// - Emoji: với vật đếm được, dùng emoji phổ thông, thân thiện trẻ em, chèn trực tiếp và cách nhau bởi dấu cách, ví dụ 🏓 🍎 ⭐ 🐟 🎈 🚗 🌸 🍓 ⚽ 🐶. Mỗi câu chỉ dùng MỘT loại emoji.
// - Hình học: CHỈ dùng token "[icon:NAME]" với NAME thuộc: triangle, square, rectangle, circle, star, diamond, oval, pentagon, hexagon, heart. Lặp token để biểu diễn nhiều hình, ví dụ "[icon:triangle] [icon:triangle] [icon:triangle]". TUYỆT ĐỐI không tự đặt tên "[icon:...]" khác — nếu hình không có trong danh sách, hãy dùng emoji.

// VÍ DỤ VISUAL (từng object câu hỏi, cùng schema như trên):
// {"question_number": 2, "question_type": "COUNT", "question_name": "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?", "answers": [{"label":"A","content":"5"},{"label":"B","content":"6"},{"label":"C","content":"7"},{"label":"D","content":"4"}], "right_answer": "B", "correct_answer": "6", "topic": "phép đếm", "difficulty": 1}
// {"question_number": 3, "question_type": "IDENTIFY_SHAPE", "question_name": "[icon:triangle]", "answers": [{"label":"A","content":"Hình tam giác"},{"label":"B","content":"Hình tròn"},{"label":"C","content":"Hình vuông"},{"label":"D","content":"Hình chữ nhật"}], "right_answer": "A", "correct_answer": "Hình tam giác", "topic": "hình học cơ bản", "difficulty": 1}
// `

const visualQuestionRulesVN = `
VISUAL QUESTION RULES:
- Mỗi câu có "question_type": ARITHMETIC (câu chữ thuần, mặc định) hoặc COUNT (đếm bằng icon). KHÔNG tạo bất kỳ question_type nào khác.
- VIỆC dùng COUNT/icon hay không và với tần suất nào do CHÍNH SÁCH icon trong GRADE PROFILE quyết định HOÀN TOÀN. Nếu chính sách đó là TẮT (hoặc không có GRADE PROFILE), dùng ARITHMETIC cho mọi câu và TUYỆT ĐỐI không phát sinh emoji hay token [icon:...]. Chỉ dùng COUNT khi chính sách cho phép, và trong đúng tần suất nó nêu.
- COUNT (chỉ khi chính sách theo lớp cho phép): "question_name" hiển thị các vật để đếm hoặc cộng bằng emoji kèm toán tử (+, "?"), ví dụ "🏓 🏓 🏓 + 🏓 🏓 🏓 = ?". Đáp án là số; "topic" là "counting".

ICONS (chỉ dùng cho COUNT; TUYỆT ĐỐI không dùng trong ARITHMETIC):
- Emoji: với vật đếm được, dùng emoji phổ thông, thân thiện trẻ em, chèn trực tiếp và cách nhau bởi dấu cách, ví dụ 🏓 🍎 ⭐ 🐟 🎈 🚗 🌸 🍓 ⚽ 🐶. Mỗi câu chỉ dùng MỘT loại emoji.
`

func buildSystemGenerateVN(n int) string {
	return fmt.Sprintf(systemGenerateVNTmpl, n, n, n) + visualQuestionRulesVN
}

func buildSystemReinforceVN(n int) string {
	return fmt.Sprintf(systemReinforceVNTmpl, n, n, n) + visualQuestionRulesVN
}

const systemGradeAssessmentVN = `Bạn là trợ lý chấm điểm bài kiểm tra toán cho học sinh tiểu học Việt Nam.

Bạn sẽ nhận được danh sách câu hỏi (JSON array) và câu trả lời của học sinh (JSON object có khoá là question_number). Hãy chấm điểm rồi dự đoán cấp lớp phù hợp nhất cho học sinh.

QUY TẮC CHẤM:
- Đối chiếu từng câu trả lời theo "question_number" với "right_answer" của câu hỏi.
- score_percentage = round(correct_number / total_questions * 100).
- "ai_review" viết bằng tiếng Việt, dài tối đa 200 ký tự, nêu một điểm mạnh và một điểm cần cải thiện cụ thể. Không xuống dòng.
- "assessment_grade" phải là một trong: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Căn cứ vào độ khó câu hỏi VÀ mức chính xác quan sát được, không chỉ dựa vào tỷ lệ đúng.

QUY TẮC ĐẦU RA:
- CHỈ trả về MỘT JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không xuống dòng trong giá trị chuỗi.

CẤU TRÚC:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Phép cộng cơ bản tốt; cần luyện thêm phép trừ có nhớ.",
  "assessment_grade": "Grade 3"
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
- "assessment_grade" phải là một trong: "Kindergarten", "Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5". Lấy cấp lớp hiện tại làm mốc; chỉ điều chỉnh khi kết quả thực sự rõ ràng.

QUY TẮC ĐẦU RA:
- CHỈ trả về MỘT JSON object theo cấu trúc bên dưới. Không lời dẫn, không khung markdown, không xuống dòng trong giá trị chuỗi.

CẤU TRÚC:
{
  "total_questions": 5,
  "correct_number": 4,
  "score_percentage": 80,
  "ai_review": "Phép trừ tốt hơn; cần luyện thêm cộng nhiều chữ số.",
  "assessment_grade": "Grade 3"
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
	intent := learningIntentVN(in.TypeOfQuiz)
	context := buildCurriculumContextVN(in)
	gradeBlock := gradeProfileBlock(QuizLanguageVietnamese, in.Grade)

	if context == "" {
		guidance := "Điều chỉnh độ khó theo thông tin đã cung cấp; nếu không có, hãy chọn bộ câu hỏi cân bằng ở trình độ tiểu học."
		if purpose == QuizPurposePractice {
			guidance = "Điều chỉnh độ khó theo mốc học kỳ khi có; nếu không, hãy chọn bộ luyện tập cân bằng ở trình độ tiểu học."
		}
		return fmt.Sprintf("Hãy tạo bài kiểm tra %s.\n\nMục tiêu học tập: %s.\n\nKhông có thông tin học vấn cụ thể — hãy dùng bộ câu hỏi toán cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).\n\n%s", purpose, intent, guidance)
	}

	// Khi đã có lớp, GRADE PROFILE là nguồn quyết định độ khó; context
	// chương trình chỉ tinh chỉnh chủ đề — bỏ cụm "cân bằng tiểu học" vốn
	// kéo độ khó về lớp 1.
	guidance := "Hiệu chỉnh mọi câu hỏi đúng theo GRADE PROFILE ở trên; thông tin chương trình chỉ tinh chỉnh chủ đề, không phải độ khó."
	header := ""
	if gradeBlock != "" {
		header = gradeBlock + "\n\n"
	} else {
		guidance = "Điều chỉnh độ khó theo thông tin học sinh ở trên."
	}
	return fmt.Sprintf("%sHãy tạo bài kiểm tra %s cho học sinh với thông tin sau:\n%s\n\nMục tiêu học tập: %s.\n\n%s", header, purpose, context, intent, guidance)
}

func userReinforceVN(purpose QuizPurpose, in QuizPromptInput) string {
	intent := learningIntentVN(QuizTypeOfQuizReinforcement)
	context := buildCurriculumContextVN(in)
	gradeBlock := gradeProfileBlock(QuizLanguageVietnamese, in.Grade)

	if context == "" {
		closing := "Tập trung vào điểm yếu trong nhận xét trước, đồng thời giữ độ khó ở mức cân bằng cho học sinh tiểu học."
		return fmt.Sprintf(`Hãy tạo bài kiểm tra %s củng cố.

Mục tiêu học tập: %s.

Không có thông tin học vấn cụ thể — hãy dùng bộ câu hỏi toán cân bằng cho học sinh tiểu học Việt Nam (Lớp 1-5).

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

%s`, purpose, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
	}

	closing := "Tập trung vào điểm yếu trong nhận xét trước, nhưng giữ MỌI câu hỏi đúng độ khó mà GRADE PROFILE ở trên quy định."
	header := ""
	if gradeBlock != "" {
		header = gradeBlock + "\n\n"
	} else {
		closing = "Tập trung vào điểm yếu trong nhận xét trước, đồng thời giữ độ khó phù hợp với thông tin nêu trên."
	}
	return fmt.Sprintf(`%sHãy tạo bài kiểm tra %s củng cố cho học sinh với thông tin sau:
%s

Mục tiêu học tập: %s.

Câu hỏi bài trước (JSON): %s
Câu trả lời của học sinh (JSON): %s
Nhận xét AI về kết quả trước: %s

%s`, header, purpose, context, intent, in.PreviousQuestions, in.PreviousAnswers, in.PreviousAIReview, closing)
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
