package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// Vietnamese exam-generation templates: the same rule sheet as
// exam_templates_en.go, in Vietnamese, so the two prompt languages can be
// compared against each other. Production sends the English one
// (examPromptLanguage in module/exam/bot_service.go) because English
// instructions tokenise shorter and a generation call pays for its prompt
// every time.
//
// Both templates are now the teaching team's own format — CASE KG /
// CASE NUM plus the 11 kindergarten question types — so a rule added to
// one belongs in the other. The GRADE PROFILE block this file used to put
// at the top of the user message went with that switch: the CASE blocks
// carry the per-band content rules now, and two authorities on the same
// question is how a prompt starts contradicting itself. grade_profile.go
// is still live — the exercise prompts render it, and GradeLabel reads the
// band names out of it.
//
// The ROUND is Vietnamese in both templates, and not by omission: the
// product serves Vietnamese children and ma_exam_pools has no language
// column, so a stored row could not record being anything else. JSON keys
// stay English — the parser, the DB columns and the mobile client all
// share one shape regardless of the prose language.
//
// The system prompt is sent on every generation call, so each line here is
// paid for on every round; the rules that survived are the ones the server
// cannot enforce after the fact. Everything the server DOES enforce —
// question_grade, question_level, the title — is stamped in code
// afterwards; the title is not even asked for.

// systemExamVNTmpl is the whole system prompt, filled by strings.Replacer:
// {{N}} is the question count; {{EMOJI}} is examAllowedEmoji;
// {{PROBE_RULE}}, {{KG_GRADE}} and {{NUM_GRADE}} are the three lines that
// name the probe positions (examProbeLinesVN), so the prompt and the
// server-side re-stamp always agree on which questions reach up a grade.
//
// The grade is addressed by its Vietnamese label throughout — the CASE
// switch keys on "Mẫu giáo" / "Lớp N" and question_grade is asked for as
// that label — so the user message binds {grade} to the same label. The
// parser maps it back to the stored int.
const systemExamVNTmpl = `VAI TRÒ: Soạn đề Toán tiếng Việt cho trẻ Mẫu giáo và Lớp 1–5.
Tạo CHÍNH XÁC {{N}} câu hỏi trắc nghiệm tiếng Việt cho {grade}.

QUY ĐỊNH CHUNG:
* Độ khó tăng dần từ Q1 đến Q{{N}}.
* {{PROBE_RULE}}
* Đúng 4 phương án (A–D), 1 đáp án đúng, các phương án sai phải hợp lý.
* Mọi câu phải có đủ dữ kiện để giải; không ra câu hỏi thiếu thông tin.
* Không lặp lại câu hỏi hay phép tính.
* PHÂN SỐ / HỖN SỐ — BẮT BUỘC: question_type PHẢI là "FRACTION", và mọi phân số ở đề bài và ở cả 4 phương án PHẢI viết bằng LaTeX — không viết thường, không dùng phân số Unicode (không 3/4, không ½): phân số "$\frac{tử}{mẫu}$" (ví dụ "$\frac{1}{2}$"); hỗn số "$phầnnguyên\frac{tử}{mẫu}$" (ví dụ "$2\frac{1}{3}$").
* Toàn bộ nội dung cho trẻ phải bằng tiếng Việt; không dùng từ tiếng Anh.

TRƯỜNG HỢP MG — nếu {grade} = "Mẫu giáo":
* {{KG_GRADE}}
* Soạn câu hỏi Mẫu giáo theo 11 dạng câu hỏi dưới đây.
* Dùng thật đa dạng icon trong danh sách cho phép. Tránh lặp lại cùng một icon quá nhiều.
* CHỈ được dùng icon trong danh sách sau: {{EMOJI}}

11 DẠNG CÂU HỎI MẪU GIÁO:
Định dạng: DẠNG | mẫu câu hỏi | ví dụ question_name | đáp án đúng
1. COUNTING | ICON×n = ? | 🍎🍎🍎 = ? | 3
2. NUM_MATCH | N = ? (icon tương ứng) | 3 = ? | 🥝🥝🥝
3. ADD | G + G = ? | 🦉🦉 + 🦉 = ? | 3
4. SUB | G − G = ? | 🐝🐝🐝 − 🐝 = ? | 2
5. COMP | G = G + ? | 🚁🚁🚁 = 🚁🚁 + ? | 1
6. CMP | G ? G → >, <, = | 🎁🎁🎁 ? ⚽️⚽️ | >
7. PATTERN | ICON×4 ? | 🔴🟡🔴🟡🔴 ? | 🟡
8. ODD_ONE | chọn vật khác loại | Chọn khác loại:  🍒 | 🍑 | 🍍| 🦋
9. NUM_SEQ | N N ? N | 1 2 ? 4 | 3
10. BEFORE_AFTER | N ? N | 2 ? 4 | 3
11.ORDER_NUM | N N N → sắp xếp tăng/giảm dần | Sắp xếp tăng dần: 321 | 123

TRƯỜNG HỢP SỐ — nếu {grade} = Lớp 1–5:
* {{NUM_GRADE}}
* KHÔNG dùng emoji.
* Câu hỏi phải theo chương trình của một trong các bộ sách: Chân Trời Sáng Tạo / Kết Nối Tri Thức / Cánh Diều.
* Câu hỏi rõ ràng, phù hợp độ tuổi và khó dần.
* Ưu tiên câu hỏi về số và phép tính, hạn chế chữ.

### KIỂM TRA ĐÁP ÁN — BẮT BUỘC:

Với MỖI câu hỏi:
1. GIẢI VÀ XÁC MINH TRƯỚC
* Giải câu hỏi rồi tính lại để chắc chắn đáp án đúng.
* TUYỆT ĐỐI không đoán.

2. TẠO 4 PHƯƠNG ÁN A, B, C, D
* Cả 4 phương án PHẢI khác nhau.
* Đáp án đúng PHẢI xuất hiện ĐÚNG MỘT LẦN.
* 3 phương án còn lại PHẢI sai.
* TUYỆT ĐỐI không lặp lại hay dùng đáp án đúng làm phương án gây nhiễu.

3. ĐỐI CHIẾU
* A.content ≠ B.content ≠ C.content ≠ D.content
* Đáp án đúng xuất hiện đúng một lần.
* Cả 3 phương án gây nhiễu đều sai.

4. GÁN NHÃN SAU CÙNG
* right_answer_content = đúng nội dung của phương án đúng.
* right_answer_label = nhãn A/B/C/D của phương án chứa "right_answer_content".
* TUYỆT ĐỐI không tự sinh "right_answer_label" độc lập.
* Nhãn PHẢI suy ra từ 4 phương án A–D thực tế.

KIỂM TRA CUỐI:
* Giải → Xác minh → Tạo A–D → Đối chiếu → Xác định phương án đúng → Gán nhãn của nó.
* Nếu đáp án hoặc nhãn không khớp, SỬA trước khi trả kết quả.

ĐẦU RA:
Chỉ trả về JSON hợp lệ. Không Markdown, không giải thích. Mọi khoá JSON phải bằng tiếng Anh.
MG: question_name và answers[].content chỉ chứa chữ số, ký hiệu toán học, icon được phép và cách sắp xếp trực quan.
Lớp 1–5: toàn bộ nội dung cho trẻ phải bằng tiếng Việt.
Khoá JSON luôn giữ nguyên tiếng Anh.
short_text — dòng mô tả bài hiển thị cho phụ huynh:
* Một cụm từ tiếng Việt, tối đa 80 ký tự, nêu 1–3 chủ đề toán chính mà các câu hỏi thực sự kiểm tra (lấy từ question_topic), chủ đề nhiều câu nhất trước.
* Ví dụ: "Phép cộng, phép trừ trong phạm vi 20" / "Đếm và so sánh số lượng trong phạm vi 5".
* KHÔNG ghi tên bộ sách, lớp, loại bài hay lời hướng dẫn kiểu "Chọn đáp án đúng".
CẤU TRÚC:
{
  "short_text": "...",
  "questions": [
    {
      "question_number": 1,
      "question_type": "...",
      "question_name": "...",
      "answers": [
        {
          "label": "A",
          "content": "..."
        }
      ],
      "right_answer_label": "...",
      "right_answer_content": "...",
      "question_topic": "...",
      "question_grade": "..."
    }
  ]
}
QUY TẮC NGÔN NGỮ QUAN TRỌNG:
Toàn bộ nội dung đề bài phải bằng tiếng Việt. Khoá JSON phải bằng tiếng Anh.`

// examProbeLinesVN is examProbeLinesEN in Vietnamese: the GENERAL probe
// rule, the KG question_grade line and the NUM one. The positions come
// from ProbePositions so the prompt asks for exactly what the server
// enforces afterwards.
func examProbeLinesVN(probes []int, n int) (rule, kg, num string) {
	if len(probes) == 0 {
		return "Mọi câu đều thuộc nội dung lớp hiện tại.",
			"Mọi câu đều là Mẫu giáo.",
			"Mọi câu đều thuộc lớp hiện tại."
	}
	probeList, restList := probeAndRestLists(probes, n)
	rule = fmt.Sprintf("%s thuộc nội dung lớp trên. %s → đầu %s.",
		joinPositions(probes, " và "), GradeLabel(enum.ExamGradeMax), GradeLabel(int(GradeProbeCeiling)))
	kg = fmt.Sprintf("%s = Mẫu giáo; %s = Lớp 1.", restList, probeList)
	num = fmt.Sprintf("%s = lớp hiện tại; %s = lớp trên.", restList, probeList)
	return rule, kg, num
}

func buildSystemExamVN(in ExamPromptInput, n int) string {
	rule, kg, num := examProbeLinesVN(ProbePositions(in.ExamType, n), n)
	return strings.NewReplacer(
		"{{N}}", fmt.Sprint(n),
		"{{PROBE_RULE}}", rule,
		"{{KG_GRADE}}", kg,
		"{{NUM_GRADE}}", num,
		"{{EMOJI}}", examAllowedEmoji,
	).Replace(systemExamVNTmpl)
}

// examPracticeBlockVN tells the model what the child just did and how the
// round must respond to it. The two modes are spelled out as separate
// rules rather than left to the model's judgement: "harder" without a
// bound drifts into the next grade, and the grade is the one thing a
// practice round must not move.
//
// The probe positions are named outright rather than as "the probe rule":
// the system prompt never uses that phrase, it states the positions.
func examPracticeBlockVN(b *PracticeBrief, n int, probes []int) string {
	gradeRule := "giữ nguyên quy định question_grade ở trên"
	if len(probes) > 0 {
		gradeRule = fmt.Sprintf("giữ nguyên quy định question_grade ở trên (%s = lớp trên, các câu còn lại = current_grade)",
			joinPositions(probes, ", "))
	}

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
		fmt.Fprintf(&sb, "- YÊU CẦU: %s; cả %d câu tập trung vào các chủ đề đã sai (ưu tiên chủ đề sai nhiều hơn).\n", gradeRule, n)
		sb.WriteString("- Cùng kỹ năng với câu đã sai nhưng KHÔNG lặp lại nguyên văn: đổi số, đổi ngữ cảnh.\n")
		if len(b.StrongTopics) > 0 {
			fmt.Fprintf(&sb, "- Có thể xen 1-2 câu ở chủ đề học sinh đã làm đúng để củng cố: %s.", joinOr(b.StrongTopics, ""))
		}
	default: // ADVANCE
		fmt.Fprintf(&sb, "- Học sinh vừa làm ĐÚNG toàn bộ, ở các chủ đề: %s.\n", joinOr(b.StrongTopics, "(không rõ chủ đề)"))
		fmt.Fprintf(&sb, "- YÊU CẦU: %s; ngoài các câu dò trần đó, cả %d câu KHÔNG được vượt current_grade.\n", gradeRule, n)
		sb.WriteString("- Nhưng phải khó hơn TRONG current_grade: số lớn hơn trong phạm vi cho phép, nhiều bước tính hơn, có bài toán có lời;\n")
		sb.WriteString("  và/hoặc chuyển sang các chủ đề khác trong chương trình mà học sinh chưa được kiểm tra.")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// examAvoidBlock lists the stems the child has already met, one per
// line, under a heading in the prompt's language. Blank stems are
// skipped; an empty list renders nothing so the caller can append it
// unconditionally.
func examAvoidBlock(lang QuizLanguage, stems []string) string {
	var sb strings.Builder
	for _, stem := range stems {
		if stem = strings.TrimSpace(stem); stem != "" {
			sb.WriteString("- " + stem + "\n")
		}
	}
	if sb.Len() == 0 {
		return ""
	}
	head := "TRÁNH LẶP LẠI\nHọc sinh đã gặp các câu sau; KHÔNG dùng lại hay chỉ đổi nhẹ (đổi số, đổi vật, đổi cấu trúc):\n"
	if lang == QuizLanguageEnglish {
		head = "AVOID REPEATS\nThe child has already seen these questions; do not reuse or lightly reword them (change the numbers, objects, or structure):\n"
	}
	return head + strings.TrimRight(sb.String(), "\n")
}

// joinOr renders a topic list, or the fallback when there is none.
func joinOr(items []string, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return strings.Join(items, ", ")
}

// examContextVN renders only the curriculum lines that carry a value, so a
// request with no semester or program still produces a coherent brief.
func examContextVN(in ExamPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Học kỳ: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Bộ sách: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildUserExamVN binds {grade} to the Vietnamese band label the system
// prompt's CASE switch and question_grade vocabulary are written in, then
// adds the per-request blocks in the same order as buildUserExamEN.
func buildUserExamVN(in ExamPromptInput, n int) string {
	var out strings.Builder
	out.WriteString("current_grade: " + GradeLabel(in.Grade) + "\n")

	// The intensity block refines the grade it follows and must not be
	// read as licence to leave it.
	if in.ExamType == enum.ExamTypeGrade && in.Level != nil {
		if block := levelProfileBlock(QuizLanguageVietnamese, *in.Level); block != "" {
			out.WriteString("\n" + block + "\n")
		}
	}

	if in.ExamType == enum.ExamTypePractice && in.Practice != nil {
		out.WriteString("\n" + examPracticeBlockVN(in.Practice, n, ProbePositions(in.ExamType, n)) + "\n")
	}

	// CASE NUM lets the model follow whichever of the three textbooks it
	// likes; a stated one must win over that choice, or the request field
	// does nothing.
	if ctx := examContextVN(in); ctx != "" {
		out.WriteString("\nTHÔNG TIN CHƯƠNG TRÌNH (chỉ để chọn chủ đề — KHÔNG dùng để tăng hay giảm độ khó; nếu đã nêu bộ sách thì dùng bộ đó, không tự chọn):\n" + ctx + "\n")
	}
	if avoid := examAvoidBlock(QuizLanguageVietnamese, in.Avoid); avoid != "" {
		out.WriteString("\n" + avoid + "\n")
	}

	return strings.TrimRight(out.String(), "\n")
}
