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
//
// The system prompt is deliberately terse. It is sent on every generation
// call, so each line here is paid for in tokens on every round; the rules
// that survived are the ones the server cannot enforce after the fact.
// Everything the server DOES enforce — question_grade, question_level,
// the title — is stamped in code afterwards regardless of what the model
// wrote.

// systemExamVNHead opens the system prompt. Slots, in order: question
// count, last question number, the probe rule (examProbeRuleVN).
const systemExamVNHead = `Bạn là AI tạo bài kiểm tra Toán cho trẻ Mẫu giáo và Lớp 1–5 Việt Nam.
Tạo CHÍNH XÁC %d câu trắc nghiệm theo GRADE PROFILE.

### RULES

- Mỗi câu có đúng 4 đáp án A, B, C, D, chỉ 1 đáp án đúng.
- Không lặp câu hỏi/phép tính.
- Độ khó tăng dần từ Q1 → Q%d.
- Mỗi câu có question_type, chọn type phù hợp và đa dạng theo GRADE PROFILE.
%s
- Chỉ được sử dụng các emoji sau: 🍎 🍊 🍐 🍌 🍉 🍇 🍓 🍒 🍑 🍍 🥝 🥕 🌽 🍅 🥦 🥒 🍭 🍬 🍪 🍩 🎂 🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨 🐯 🦁 🐮 🐷 🐸 🐵 🐔 🐧 🐦 🐤 🦆 🦉 🐟 🐠 🐡 🦋 🐝 🐞 🐢 🚗 🚕 🚌 🚎 🚲 🛵 🚂 ✈️ 🚁 🚢 ⭐️ 🎈 ⚽️ 🧸 📚 ✏️ 🖍️ 🎁 🔴 🟡 🟢 🔵 🟠 🟣 🟥 🟨 🟩 🟦 🟧 🟪
- Không dùng emoji để trang trí hoặc đặt ngẫu nhiên trong câu hỏi.
- Với câu hỏi số thuần túy như ARITHMETIC, SEQUENCE hoặc câu chỉ yêu cầu đọc/so sánh số, không dùng emoji.
- Nếu sử dụng emoji, mỗi câu chỉ dùng 1 loại emoji, có thể lặp lại emoji đó trong cùng câu.
- Phân số dùng ASCII (1/2, 3/4), không dùng Unicode/LaTeX.
- Câu tính toán trực tiếp: question_name chỉ chứa số, toán tử + - * / ^, dấu ngoặc và ?, không chữ, emoji hoặc LaTeX.
- Nội dung phải phù hợp độ tuổi và không vượt GRADE PROFILE.
- right_answer_label và right_answer_content phải khớp chính xác.

### OUTPUT

Chỉ trả về JSON hợp lệ, không Markdown, không giải thích, không text ngoài JSON.
CẤU TRÚC:
{
  "title": "Lớp 1",
  "short_text": "Phép cộng và phép trừ trong phạm vi 20",
  "questions":[
    {
      "question_number": 1,
      "question_type": "COUNT",
      "question_name": "🍎 🍎 🍎 + 🍎 🍎 = ?",
      "answers": [
        {"label": "A", "content": "5"},
        {"label": "B", "content": "4"},
        {"label": "C", "content": "6"},
        {"label": "D", "content": "3"}
      ],
      "right_answer_label": "A",
      "right_answer_content": "5",
      "question_topic": "phép cộng trong phạm vi 20",
      "question_grade": 1
    }
  ]
}
`

// systemExamVNTail closes the system prompt. One slot: question count.
const systemExamVNTail = `
Tự kiểm tra trước khi trả kết quả: đúng %d câu, 4 đáp án/câu, 1 đáp án đúng, đúng question_grade, không trùng, đúng độ khó và đúng GRADE PROFILE.`

// examProbeRuleVN renders the probe rule in terms of current_grade, which
// the user message then binds to a number. The positions come from
// ProbePositions so the prompt and the server-side re-stamp always name
// the same questions: the prompt asks, the server enforces.
func examProbeRuleVN(in ExamPromptInput, n int) string {
	positions := ProbePositions(in.ExamType, n)
	if len(positions) == 0 {
		return `- Mọi câu: question_grade = current_grade.`
	}

	labels := make([]string, 0, len(positions))
	for _, p := range positions {
		labels = append(labels, fmt.Sprintf("Q%d", p))
	}
	and := strings.Join(labels, " và ")
	list := strings.Join(labels, ", ")
	slash := strings.Join(labels, "/")

	return fmt.Sprintf(`- %s là DÒ TRẦN, khó hơn current_grade đúng 1 grade:
  - %s: question_grade = current_grade + 1
  - Các câu còn lại: question_grade = current_grade
  - Nếu current_grade = %d, %s trong phạm vi lớp %d.
- Không tiết lộ %s là câu dò trần.`,
		and, list, enum.ExamGradeMax, slash, GradeProbeCeiling, slash)
}

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
// per-question difficulty field, and an exemplar carrying one would teach
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

func buildSystemExamVN(in ExamPromptInput, n int) string {
	return fmt.Sprintf(systemExamVNHead, n, n, examProbeRuleVN(in, n)) +
		fmt.Sprintf(systemExamVNTail, n)
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
		fmt.Fprintf(&sb, "- YÊU CẦU: giữ nguyên RULES về question_grade (kể cả câu dò trần); cả %d câu tập trung vào các chủ đề đã sai (ưu tiên chủ đề sai nhiều hơn).\n", n)
		sb.WriteString("- Cùng kỹ năng với câu đã sai nhưng KHÔNG lặp lại nguyên văn: đổi số, đổi ngữ cảnh.\n")
		if len(b.StrongTopics) > 0 {
			fmt.Fprintf(&sb, "- Có thể xen 1-2 câu ở chủ đề học sinh đã làm đúng để củng cố: %s.", joinOr(b.StrongTopics, ""))
		}
	default: // ADVANCE
		fmt.Fprintf(&sb, "- Học sinh vừa làm ĐÚNG toàn bộ, ở các chủ đề: %s.\n", joinOr(b.StrongTopics, "(không rõ chủ đề)"))
		fmt.Fprintf(&sb, "- YÊU CẦU: giữ nguyên RULES về question_grade (kể cả câu dò trần); ngoài câu dò, cả %d câu KHÔNG lên cấp lớp cao hơn.\n", n)
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

// examContextVN renders only the curriculum lines that carry a value, so a
// request with no semester or program still produces a coherent brief.
func examContextVN(in ExamPromptInput) string {
	var b strings.Builder
	if v := strings.TrimSpace(in.Semester); v != "" {
		fmt.Fprintf(&b, "- Học kỳ: %s\n", v)
	}
	if v := strings.TrimSpace(in.Program); v != "" {
		fmt.Fprintf(&b, "- Chương trình học: %s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

// buildUserExamVN assembles the per-request brief. The GRADE PROFILE
// comes first because it is the authority every rule defers to; then
// current_grade, which binds the probe rule in the system prompt to a
// number; then the optional refinements (intensity, practice brief) and
// the curriculum context.
func buildUserExamVN(in ExamPromptInput, n int) string {
	var out strings.Builder

	if block := gradeProfileBlockByLevel(QuizLanguageVietnamese, GradeLevel(in.Grade)); block != "" {
		out.WriteString(examVocabulary(block) + "\n\n")
	}
	fmt.Fprintf(&out, "current_grade: %d (%s)\n", in.Grade, ExamTitle(in.Grade))

	// The intensity block refines the grade profile it follows and must
	// not be read as licence to leave the grade.
	if in.ExamType == enum.ExamTypeGrade && in.Level != nil {
		if block := levelProfileBlock(*in.Level); block != "" {
			out.WriteString("\n" + block + "\n")
		}
	}

	if in.ExamType == enum.ExamTypePractice && in.Practice != nil {
		out.WriteString("\n" + examPracticeBlockVN(in.Practice, n) + "\n")
	}

	if ctx := examContextVN(in); ctx != "" {
		out.WriteString("\nThông tin chương trình (chỉ để chọn chủ đề, KHÔNG dùng để tăng hay giảm độ khó):\n" + ctx + "\n")
	}

	return strings.TrimRight(out.String(), "\n")
}

func buildUserExamVN_V2(in ExamPromptInput, n int) string {
	var out strings.Builder

	fmt.Fprintf(&out, "current_grade: %d (%s)\n", in.Grade, ExamTitle(in.Grade))

	return strings.TrimRight(out.String(), "\n")
}
