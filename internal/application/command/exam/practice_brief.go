package command

import (
	"sort"
	"strings"

	"math-ai.com/math-ai/internal/domain/bot"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// BuildPracticeBrief turns the answer log of one sitting into the brief a
// PRACTICE prompt is aimed with.
//
// The mode is a fixed rule, not a judgement call: one wrong answer is
// enough to make the round a drill (RETRY_WEAK); a clean sheet makes it
// a push forward (ADVANCE). The rule is deliberately this blunt so that
// what the child gets next is explainable in one sentence to a parent.
//
// Only ANSWERED questions reach here — a skipped question has no detail
// row — so a sitting with three answers, all right, is ADVANCE. That is
// accepted: a skip says nothing about the skill either way, and guessing
// at what it meant would be worse than ignoring it.
//
// Topics are ranked by how many answers went wrong in them, ties broken
// alphabetically so the same log always yields the same brief. A topic
// counts as strong only when every answer in it was right; a topic with
// one wrong out of five is weak, because the drill is about the miss.
func BuildPracticeBrief(details []*exam.UserExamDetail) bot.PracticeBrief {
	wrongByTopic := map[string]int{}
	seenTopic := map[string]bool{}
	var wrong []bot.PracticeItem

	for _, d := range details {
		topic := topicOf(d)
		seenTopic[topic] = true
		if d.IsCorrect() {
			continue
		}
		wrongByTopic[topic]++
		wrong = append(wrong, bot.PracticeItem{
			Stem:        utils.DerefString(d.QuestionName()),
			Topic:       topic,
			RightAnswer: utils.DerefString(d.RightAnswerContent()),
			ChildAnswer: utils.DerefString(d.SelectedContent()),
		})
	}

	var weak, strong []string
	for topic := range seenTopic {
		if topic == "" {
			continue
		}
		if wrongByTopic[topic] > 0 {
			weak = append(weak, topic)
		} else {
			strong = append(strong, topic)
		}
	}
	sort.Slice(weak, func(i, j int) bool {
		if wrongByTopic[weak[i]] != wrongByTopic[weak[j]] {
			return wrongByTopic[weak[i]] > wrongByTopic[weak[j]]
		}
		return weak[i] < weak[j]
	})
	sort.Strings(strong)

	mode := enum.PracticeModeAdvance
	if len(wrong) > 0 {
		mode = enum.PracticeModeRetryWeak
	}
	return bot.PracticeBrief{
		Mode:         mode,
		Wrong:        wrong,
		WeakTopics:   weak,
		StrongTopics: strong,
	}
}

// topicOf normalises the stored topic for grouping. Topics come from the
// model and arrive in whatever case and spacing it chose; without this
// "Phép cộng" and "phép cộng" would rank as two topics.
func topicOf(d *exam.UserExamDetail) string {
	return strings.ToLower(strings.TrimSpace(utils.DerefString(d.QuestionTopic())))
}
