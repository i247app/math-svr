package bot

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

// English exam-generation templates: the same rule sheet as
// exam_templates_vn.go, written in English because English instructions
// tokenise shorter, and a generation call pays for its prompt every time.
//
// Only the INSTRUCTIONS are English. The round itself — question text,
// answer content, question_topic, short_text — is Vietnamese in both
// templates, and this one says so explicitly, since a model reading an
// English prompt will otherwise answer in English.
//
// The user message here is only the grade. The GRADE PROFILE the rules
// refer to, the LEVEL PROFILE and the practice brief are not sent in this
// language — that is the shape the teaching team asked to trial, and the
// Vietnamese template still renders all of them for comparison.

// systemExamENHead is the whole system prompt. Slots, in order: question
// count, last question number (twice: the no-repeat rule and the
// difficulty ramp), the probe rule (examProbeRuleEN).
const systemExamENHead = `You generate Math tests for Vietnamese children: kindergarten and grades 1–5.
Create EXACTLY %d multiple-choice questions according to the GRADE PROFILE.

RULES
* Each question must have exactly 4 answer options: A, B, C, and D, with exactly 1 correct answer.
* Do not repeat questions, calculations, mathematical patterns, or identical question structures across the %d questions.
* Difficulty must increase progressively from Question 1 → Question %d.
* Each question must have a question_type; choose types that are appropriate and diverse according to the GRADE PROFILE.
%s
* Fractions must use ASCII format (e.g. 1/2, 3/4), never Unicode fractions or LaTeX.
* Direct calculation questions: question_name must contain only digits, the operators + - * / ^, parentheses, and ?, with no words, emojis, units, or LaTeX.
* Content must be age-appropriate and must not exceed the GRADE PROFILE.
* right_answer_label and right_answer_content must match exactly.
* All text that the child reads — question_name, answer content, question_topic, and short_text — must be written in Vietnamese. JSON keys must remain in English.

EMOJI RULES
* Use only emojis from the following allowed list:
  🍎 🍊 🍐 🍌 🍉 🍇 🍓 🍒 🍑 🍍 🥝 🥕 🌽 🍅 🥦 🥒 🍭 🍬 🍪 🍩 🎂
  🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨 🐯 🦁 🐮 🐷 🐸 🐵 🐔 🐧 🐦 🐤 🦆 🦉
  🐟 🐠 🐡 🦋 🐝 🐞 🐢
  🚗 🚕 🚌 🚎 🚲 🛵 🚂 ✈️ 🚁 🚢
  ⭐️ 🎈 ⚽️ 🧸 📚 ✏️ 🖍️ 🎁
  🔴 🟡 🟢 🔵 🟠 🟣 🟥 🟨 🟩 🟦 🟧 🟪
* Use emojis only when they directly represent necessary mathematical or visual information, including questions involving counting, quantity comparison, classification, colors, logic, patterns, and visual geometry.
* Do not use emojis for purely numerical questions such as ARITHMETIC, SEQUENCE, NUMBER_READING, NUMBER_ORDER, calculations, fractions, or purely numerical measurement questions.
* Never use emojis merely for decoration.

QUESTION NAME RULES
* question_name must be concise, natural, and immediately understandable by the child.
* For questions asking "how many", question_name should display only the emojis needed to represent the objects.
* For visual questions, question_name must contain all necessary visual information required to answer the question.
* Do not create a question that requires information not provided in question_name or the answer options.
* For direct calculation questions:
  * question_name must contain only digits, the operators + - * / ^, parentheses, and ?.
  * Do not include words, emojis, units, LaTeX, or other symbols.
  * Example: 25 + 17 = ?
* Never make question_name unnecessarily long or complicated.

ANSWER OPTION RULES
* Every question must have exactly 4 answer options: A, B, C, and D.
* Exactly ONE option must be mathematically correct.
* All four options must be plausible and relevant to the question.
* Do not include duplicate answer contents.
* Do not include two options that could both reasonably be interpreted as correct.
* Do not use trick answers caused by ambiguous wording.
* Answer options must be appropriate for the child's grade and the question type.
* For numeric questions, answer options must use clear numeric values.
* For visual questions, answer options must contain only the necessary visual content when appropriate.
* Do not introduce information in an answer option that changes the meaning of the question.

MATHEMATICAL ACCURACY RULES
* Generate questions and answers with the HIGHEST POSSIBLE MATHEMATICAL ACCURACY.
* Every question must have exactly one mathematically, logically, or visually correct answer.
* Solve every question independently before assigning the correct answer.
* Recalculate every arithmetic operation before returning the JSON.
* Verify addition, subtraction, multiplication, division, powers, fractions, comparisons, sequences, measurements, geometry, and logical relationships whenever applicable.
* For word problems, verify the complete reasoning chain from the given information to the final answer.
* For visual counting questions, independently count the represented objects and verify the count.
* For comparison questions, independently compare all relevant quantities before selecting the correct answer.
* For geometry questions, independently verify the geometric properties and calculations.
* For sequence questions, verify that the intended pattern is uniquely determined and that only one option can correctly continue the sequence.
* Never guess an answer.
* Never rely on an unverified calculation.
* Never create an answer first and then construct a question around it if this could introduce an error.
* If a generated question is uncertain, ambiguous, mathematically questionable, or difficult to verify, discard it and generate a new one.
* If any answer option is accidentally correct for a second reason, revise the question or the options.
* The final right_answer_label MUST identify the mathematically correct option.
* The final right_answer_content MUST exactly match the content of the corresponding answer option character-for-character.

OUTPUT

Return only valid JSON — no Markdown, no explanation, and no text outside the JSON.
The JSON keys must remain in English, while all child-facing content must be in Vietnamese.

STRUCTURE:
{
	"title": "Lớp 1",
	"short_text": "Phép cộng và phép trừ trong phạm vi 20",
	"questions": [
		{
		"question_number": 1,
		"question_type": "ARITHMETIC",
		"question_name": "5 + 3 = ?",
		"answers": [
			{ "label": "A", "content": "8" },
			{ "label": "B", "content": "9" },
			{ "label": "C", "content": "10" },
			{ "label": "D", "content": "7" }
		],
		"right_answer_label": "A",
		"right_answer_content": "8",
		"question_topic": "phép cộng trong phạm vi 20",
		"question_grade": 1
		}
	]
}`

// examProbeRuleEN renders the probe rule in terms of current_grade, which
// the user message then binds to a number. The positions come from
// ProbePositions so the prompt and the server-side re-stamp always name
// the same questions: the prompt asks, the server enforces.
func examProbeRuleEN(in ExamPromptInput, n int) string {
	positions := ProbePositions(in.ExamType, n)
	if len(positions) == 0 {
		return `* Every question: question_grade = current_grade.`
	}

	long := make([]string, 0, len(positions))
	short := make([]string, 0, len(positions))
	for _, p := range positions {
		long = append(long, fmt.Sprintf("Question %d", p))
		short = append(short, fmt.Sprintf("Q%d", p))
	}

	return fmt.Sprintf(`* %s are ability ceiling-probe questions, exactly 1 grade harder than current_grade:
  * %s: question_grade = current_grade + 1
  * All other questions: question_grade = current_grade
  * If current_grade = %d, %s must still remain within Grade %d scope.`,
		strings.Join(long, " and "), strings.Join(short, ", "),
		enum.ExamGradeMax, strings.Join(short, "/"), GradeProbeCeiling)
}

func buildSystemExamEN(in ExamPromptInput, n int) string {
	return fmt.Sprintf(systemExamENHead, n, n, n, examProbeRuleEN(in, n))
}

// buildUserExamEN binds current_grade, which every rule in the system
// prompt is phrased against. It is rendered as the number the probe rule
// does arithmetic on, with the band name beside it — the name alone would
// leave "current_grade + 1" for the model to work out from a label, and
// question_grade is an integer in the schema.
func buildUserExamEN(in ExamPromptInput, _ int) string {
	return fmt.Sprintf("CURRENT GRADE\n\ncurrent_grade: %d (%s)", in.Grade, ExamTitle(in.Grade))
}
