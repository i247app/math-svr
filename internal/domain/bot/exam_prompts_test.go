package bot

import (
	"strings"
	"testing"

	"math-ai.com/math-ai/internal/shared/enum"
)

func TestProbePositions(t *testing.T) {
	tests := []struct {
		name     string
		examType enum.ExamType
		numQues  int
		want     []int
	}{
		{"assessment of ten probes at 3 and 6", enum.ExamTypeAssessment, 10, []int{3, 6}},
		{"assessment shorter than the second slot", enum.ExamTypeAssessment, 5, []int{3}},
		{"assessment too short to probe at all", enum.ExamTypeAssessment, 2, nil},
		{"zero falls back to the default length", enum.ExamTypeAssessment, 0, []int{3, 6}},
		{"practice probes like an assessment", enum.ExamTypePractice, 10, []int{3, 6}},
		{"grade review probes too", enum.ExamTypeGrade, 10, []int{3, 6}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ProbePositions(tc.examType, tc.numQues)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// TestProbeGrade pins the ceiling. Without it an ASSESSMENT at Grade 5
// would ask for Grade 7 material the moment someone bumps the product's
// top band.
func TestProbeGrade(t *testing.T) {
	tests := []struct{ grade, want int }{
		{0, 1}, {1, 2}, {4, 5}, {5, 6},
	}
	for _, tc := range tests {
		if got := ProbeGrade(tc.grade); got != tc.want {
			t.Errorf("ProbeGrade(%d) = %d, want %d", tc.grade, got, tc.want)
		}
	}
}

func TestBuildExamPromptRejectsBadInput(t *testing.T) {
	tests := []struct {
		name string
		in   ExamPromptInput
	}{
		{"grade below range", ExamPromptInput{ExamType: enum.ExamTypePractice, Grade: -1}},
		{"grade above range", ExamPromptInput{ExamType: enum.ExamTypePractice, Grade: 6}},
		{"unknown exam type", ExamPromptInput{ExamType: enum.ExamType("HOMEWORK"), Grade: 1}},
		{"empty exam type", ExamPromptInput{Grade: 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := BuildExamPrompt(tc.in); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

func TestBuildExamPromptAssessment(t *testing.T) {
	system, user, err := BuildExamPrompt(ExamPromptInput{
		ExamType:     enum.ExamTypeAssessment,
		Grade:        1,
		NumQuestions: 10,
		Semester:     "Học kỳ 1",
		Program:      "Cánh diều",
	})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}

	// The probe rule is a fixed part of the system prompt, phrased against
	// current_grade; the user message binds that name to a number.
	for _, want := range []string{
		"Tạo CHÍNH XÁC 10 câu",
		"Q3 và Q6 là DÒ TRẦN",
		`Q3, Q6: question_grade = current_grade + 1`,
		`Các câu còn lại: question_grade = current_grade`,
		`"short_text": "Phép cộng và phép trừ trong phạm vi 20"`,
		`"questions":[`,
		"Nếu current_grade = 5, Q3/Q6 trong phạm vi lớp 6",
		"Không tiết lộ Q3/Q6",
		"đúng 10 câu, 4 đáp án/câu",
	} {
		if !strings.Contains(system, want) {
			t.Errorf("system prompt is missing %q", want)
		}
	}
	for _, want := range []string{
		"GRADE PROFILE",
		"current_grade: 1 (Lớp 1)",
		"Học kỳ 1",
		"Cánh diều",
	} {
		if !strings.Contains(user, want) {
			t.Errorf("user prompt is missing %q", want)
		}
	}
	// The title slot is in the schema but no rule pins its value: the
	// server stamps ExamTitle over whatever the model wrote.
	if strings.Contains(user, "title") {
		t.Error("the user message must not carry a title rule; the server stamps the title")
	}

	// The authoritative block must precede everything else, or the schema
	// example teaches its own difficulty by imitation.
	if strings.Index(user, "GRADE PROFILE") > strings.Index(user, "current_grade:") {
		t.Error("GRADE PROFILE must open the user message")
	}
}

// TestBuildExamPromptProbeRuleFollowsLength: the probe rule names only the
// positions that exist in a shorter round, and drops to a flat rule when
// none do — the same list NormalizeQuestionBands enforces afterwards.
func TestBuildExamPromptProbeRuleFollowsLength(t *testing.T) {
	system, _, err := BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeAssessment, Grade: 1, NumQuestions: 5})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	if !strings.Contains(system, "Q3 là DÒ TRẦN") || strings.Contains(system, "Q6") {
		t.Error("a 5-question round probes at Q3 only")
	}

	system, _, err = BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeAssessment, Grade: 1, NumQuestions: 2})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	if !strings.Contains(system, `Mọi câu: question_grade = current_grade`) || strings.Contains(system, "DÒ TRẦN") {
		t.Error("a round too short to probe must state the flat rule")
	}
}

// TestExamPromptUsesExamVocabulary guards the rename end to end. The
// few-shot exemplar is the trap: it outweighs prose, so a single stale key
// there teaches the model the quiz spelling and every generated question
// then fails to map onto the exam's columns.
func TestExamPromptUsesExamVocabulary(t *testing.T) {
	system, user, err := BuildExamPrompt(ExamPromptInput{
		ExamType: enum.ExamTypeAssessment, Grade: 1, NumQuestions: 10,
	})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	whole := system + "\n" + user

	for _, stale := range []string{`"right_answer"`, `"correct_answer"`, `"topic"`, `"difficulty"`} {
		if strings.Contains(whole, stale) {
			t.Errorf("prompt still names the quiz-era key %s", stale)
		}
	}
	for _, want := range []string{
		`"right_answer_label"`, `"right_answer_content"`,
		`"question_topic"`, `"question_grade"`,
	} {
		if !strings.Contains(whole, want) {
			t.Errorf("prompt never names %s", want)
		}
	}
}

// TestBuildExamPromptPracticeProbes: a PRACTICE round keeps the probe rule
// — the drill stretches the child the same way the exam did — and the
// practice block defers to it rather than contradicting it.
func TestBuildExamPromptPracticeProbes(t *testing.T) {
	system, user, err := BuildExamPrompt(ExamPromptInput{
		ExamType:     enum.ExamTypePractice,
		Grade:        4,
		NumQuestions: 10,
		Practice:     &PracticeBrief{Mode: enum.PracticeModeAdvance, StrongTopics: []string{"phép cộng"}},
	})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	if !strings.Contains(system, "Q3 và Q6 là DÒ TRẦN") {
		t.Error("a PRACTICE round must keep the probe rule")
	}
	for _, want := range []string{"current_grade: 4 (Lớp 4)", "kể cả câu dò trần"} {
		if !strings.Contains(user, want) {
			t.Errorf("PRACTICE prompt lacks %q", want)
		}
	}
	if strings.Contains(system, `Mọi câu: question_grade = current_grade`) {
		t.Error("a PRACTICE round must not flatten every question to the requested grade")
	}
}

// TestBuildExamPromptGradeReviewProbes: a GRADE review carries the probe
// rule like every other round, and with a level the two blocks coexist —
// the level shapes intensity inside the band, the probe reaches one band
// up at Q3/Q6.
func TestBuildExamPromptGradeReviewProbes(t *testing.T) {
	six := 6
	system, user, err := BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeGrade, Grade: 4, NumQuestions: 10, Level: &six})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	if !strings.Contains(system, "Q3 và Q6 là DÒ TRẦN") {
		t.Error("a GRADE review must keep the probe rule")
	}
	for _, want := range []string{"current_grade: 4 (Lớp 4)", "LEVEL PROFILE — cường độ bậc 6"} {
		if !strings.Contains(user, want) {
			t.Errorf("GRADE prompt lacks %q", want)
		}
	}
	if strings.Contains(system, `Mọi câu: question_grade = current_grade`) {
		t.Error("a GRADE review must not flatten every question to the requested grade")
	}
}

// TestBuildExamPromptPracticeNeedsBrief: a PRACTICE round is always aimed
// at a previous sitting; without that brief the prompt would silently
// degrade into an ordinary round, so it is refused instead.
func TestBuildExamPromptPracticeNeedsBrief(t *testing.T) {
	_, _, err := BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypePractice, Grade: 1, NumQuestions: 10})
	if err == nil {
		t.Fatal("expected an error for a PRACTICE prompt without a brief")
	}
}

// TestBuildExamPromptPracticeModes pins what each mode tells the model:
// RETRY_WEAK lists the wrong answers and their topics; ADVANCE names the
// mastered topics and forbids moving up a grade.
func TestBuildExamPromptPracticeModes(t *testing.T) {
	retry := &PracticeBrief{
		Mode: enum.PracticeModeRetryWeak,
		Wrong: []PracticeItem{
			{Stem: "7 - 4 = ?", Topic: "phép trừ", RightAnswer: "3", ChildAnswer: "4"},
		},
		WeakTopics:   []string{"phép trừ"},
		StrongTopics: []string{"phép cộng"},
	}
	_, user, err := BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypePractice, Grade: 1, NumQuestions: 10, Practice: retry})
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	for _, want := range []string{"BÀI LUYỆN TẬP", "làm sai 1 câu", "phép trừ", "7 - 4 = ?", "đáp án đúng: 3", "học sinh chọn: 4", "KHÔNG lặp lại nguyên văn", "củng cố: phép cộng"} {
		if !strings.Contains(user, want) {
			t.Errorf("RETRY_WEAK prompt lacks %q", want)
		}
	}

	advance := &PracticeBrief{Mode: enum.PracticeModeAdvance, StrongTopics: []string{"phép cộng", "đếm"}}
	_, user, err = BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypePractice, Grade: 1, NumQuestions: 10, Practice: advance})
	if err != nil {
		t.Fatalf("advance: %v", err)
	}
	for _, want := range []string{"ĐÚNG toàn bộ", "phép cộng, đếm", "KHÔNG lên cấp lớp cao hơn", "khó hơn TRONG cấp lớp"} {
		if !strings.Contains(user, want) {
			t.Errorf("ADVANCE prompt lacks %q", want)
		}
	}
	if strings.Contains(user, "Các câu đã làm sai") {
		t.Error("ADVANCE prompt must not list wrong answers — there are none")
	}
}

// TestBuildExamPromptHasNoLevelAxis pins the decision that the exam prompt
// carries no difficulty scale: no LEVEL PROFILE block, no question_level
// key, no "Cấp độ" suffix on the title. The columns exist and stay NULL
// until the teaching team defines a rule; until then the model must not be
// told about a scale nobody can explain.
func TestBuildExamPromptHasNoLevelAxis(t *testing.T) {
	system, user, err := BuildExamPrompt(ExamPromptInput{
		ExamType:     enum.ExamTypeAssessment,
		Grade:        0,
		NumQuestions: 10,
	})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	whole := system + "\n" + user

	for _, stale := range []string{"LEVEL PROFILE", `"question_level"`, "Cấp độ", "cường độ"} {
		if strings.Contains(whole, stale) {
			t.Errorf("prompt still carries the level axis: %q", stale)
		}
	}
	if !strings.Contains(user, "current_grade: 0 (Mẫu giáo)") {
		t.Error("kindergarten binds current_grade to 0 under its band name")
	}
}

// TestExamTitle: the stored title is the band name and nothing else — the
// server stamps it now that the model is no longer asked for one.
func TestExamTitle(t *testing.T) {
	tests := []struct {
		grade int
		want  string
	}{
		{0, "Mẫu giáo"},
		{1, "Lớp 1"},
		{5, "Lớp 5"},
	}
	for _, tc := range tests {
		if got := ExamTitle(tc.grade); got != tc.want {
			t.Errorf("ExamTitle(%d) = %q, want %q", tc.grade, got, tc.want)
		}
	}
}

// TestBuildExamPromptLevelIsForGradeReviewOnly: a stated level renders
// the LEVEL PROFILE block under a GRADE review — and only there. An
// ASSESSMENT measures at the band and must not be told an intensity.
func TestBuildExamPromptLevelIsForGradeReviewOnly(t *testing.T) {
	seven := 7
	_, user, err := BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeGrade, Grade: 3, NumQuestions: 10, Level: &seven})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	for _, want := range []string{"LEVEL PROFILE — cường độ bậc 7 trên 10", "GRADE PROFILE ở trên quyết định NỘI DUNG"} {
		if !strings.Contains(user, want) {
			t.Errorf("GRADE prompt with level lacks %q", want)
		}
	}
	if i, j := strings.Index(user, "current_grade:"), strings.Index(user, "LEVEL PROFILE"); i > j {
		t.Error("the level block must follow the grade it refines")
	}

	_, user, err = BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeAssessment, Grade: 3, NumQuestions: 10, Level: &seven})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	if strings.Contains(user, "LEVEL PROFILE") {
		t.Error("an ASSESSMENT must not carry a level block")
	}

	_, user, err = BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeGrade, Grade: 3, NumQuestions: 10})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	if strings.Contains(user, "LEVEL PROFILE") {
		t.Error("a GRADE review with no stated level must not carry a level block")
	}
}

// TestBuildExamPromptEnglish: the English template is the same rule sheet
// in a cheaper tokenisation. It must keep every server-facing name, bind
// current_grade the same way, and — the one rule the Vietnamese template
// does not need — tell the model that the round itself is Vietnamese.
func TestBuildExamPromptEnglish(t *testing.T) {
	six := 6
	system, user, err := BuildExamPrompt(ExamPromptInput{
		Language:     QuizLanguageEnglish,
		ExamType:     enum.ExamTypeGrade,
		Grade:        2,
		NumQuestions: 10,
		Semester:     "Học kỳ 1",
		Program:      "Cánh diều",
		Level:        &six,
	})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}

	for _, want := range []string{
		"Create EXACTLY 10 multiple-choice questions",
		"Q3 and Q6 are CEILING PROBES",
		"Q3, Q6: question_grade = current_grade + 1",
		"If current_grade = 5, Q3/Q6 stay within grade 6",
		"is written in Vietnamese",
		`"short_text": "Phép cộng và phép trừ trong phạm vi 20"`,
		"Self-check before answering: exactly 10 questions",
	} {
		if !strings.Contains(system, want) {
			t.Errorf("EN system prompt is missing %q", want)
		}
	}
	for _, want := range []string{
		"GRADE PROFILE — AUTHORITATIVE",
		"current_grade: 2 (Lớp 2)",
		"LEVEL PROFILE — intensity 6 of 10",
		"Semester: Học kỳ 1",
		"Curriculum: Cánh diều",
	} {
		if !strings.Contains(user, want) {
			t.Errorf("EN user prompt is missing %q", want)
		}
	}
	for _, stale := range []string{"DÒ TRẦN", "BẮT BUỘC", "cường độ", "Học kỳ:"} {
		if strings.Contains(system+user, stale) {
			t.Errorf("EN prompt leaks Vietnamese instruction text %q", stale)
		}
	}

	// A PRACTICE brief renders in English but quotes the child's answers
	// as served.
	_, user, err = BuildExamPrompt(ExamPromptInput{
		Language: QuizLanguageEnglish, ExamType: enum.ExamTypePractice, Grade: 1, NumQuestions: 10,
		Practice: &PracticeBrief{
			Mode:       enum.PracticeModeRetryWeak,
			Wrong:      []PracticeItem{{Stem: "7 - 4 = ?", Topic: "phép trừ", RightAnswer: "3", ChildAnswer: "4"}},
			WeakTopics: []string{"phép trừ"},
		},
	})
	if err != nil {
		t.Fatalf("BuildExamPrompt: %v", err)
	}
	for _, want := range []string{"PRACTICE ROUND", "got 1 questions wrong", "7 - 4 = ?", "correct: 3; child chose: 4", "NEVER a verbatim repeat"} {
		if !strings.Contains(user, want) {
			t.Errorf("EN RETRY_WEAK prompt lacks %q", want)
		}
	}

	// Empty language keeps the Vietnamese template; garbage is refused.
	system, _, err = BuildExamPrompt(ExamPromptInput{ExamType: enum.ExamTypeAssessment, Grade: 1})
	if err != nil || !strings.Contains(system, "Tạo CHÍNH XÁC") {
		t.Errorf("empty Language must select the Vietnamese template (err=%v)", err)
	}
	if _, _, err = BuildExamPrompt(ExamPromptInput{Language: "fr", ExamType: enum.ExamTypeAssessment, Grade: 1}); err == nil {
		t.Error("an unsupported prompt language must be refused")
	}
}
