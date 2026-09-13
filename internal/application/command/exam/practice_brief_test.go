package command_test

import (
	"reflect"
	"testing"

	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

func answered(topic, stem, right, picked string, correct bool) *exam.UserExamDetail {
	d := exam.NewUserExamDetail()
	d.SetQuestionTopic(&topic)
	d.SetQuestionName(&stem)
	d.SetRightAnswerContent(&right)
	d.SetSelectedContent(&picked)
	d.SetIsCorrect(correct)
	return d
}

func TestBuildPracticeBrief(t *testing.T) {
	tests := []struct {
		name       string
		log        []*exam.UserExamDetail
		wantMode   enum.PracticeMode
		wantWeak   []string
		wantStrong []string
		wantWrong  int
	}{
		{
			name: "one miss makes it a drill, topics ranked by misses",
			log: []*exam.UserExamDetail{
				answered("phép cộng", "1 + 1 = ?", "2", "2", true),
				answered("phép trừ", "5 - 2 = ?", "3", "4", false),
				answered("phép trừ", "7 - 4 = ?", "3", "3", true),
				answered("đếm", "🍎🍎🍎", "3", "2", false),
				answered("đếm", "🍎🍎", "2", "1", false),
			},
			wantMode:   enum.PracticeModeRetryWeak,
			wantWeak:   []string{"đếm", "phép trừ"}, // 2 misses before 1
			wantStrong: []string{"phép cộng"},
			wantWrong:  3,
		},
		{
			name: "a clean sheet pushes forward",
			log: []*exam.UserExamDetail{
				answered("phép cộng", "1 + 1 = ?", "2", "2", true),
				answered("đếm", "🍎🍎", "2", "2", true),
			},
			wantMode:   enum.PracticeModeAdvance,
			wantStrong: []string{"phép cộng", "đếm"},
		},
		{
			name: "topic case and spacing do not split a topic",
			log: []*exam.UserExamDetail{
				answered("Phép cộng ", "1 + 1 = ?", "2", "3", false),
				answered("phép cộng", "2 + 2 = ?", "4", "4", true),
			},
			wantMode:  enum.PracticeModeRetryWeak,
			wantWeak:  []string{"phép cộng"},
			wantWrong: 1,
		},
		{
			name: "a blank topic is never listed",
			log: []*exam.UserExamDetail{
				answered("", "1 + 1 = ?", "2", "3", false),
			},
			wantMode:  enum.PracticeModeRetryWeak,
			wantWrong: 1,
		},
		{
			name:     "nothing answered is a push forward with nothing to say",
			wantMode: enum.PracticeModeAdvance,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := command.BuildPracticeBrief(tc.log)
			if got.Mode != tc.wantMode {
				t.Fatalf("mode = %s, want %s", got.Mode, tc.wantMode)
			}
			if !reflect.DeepEqual(got.WeakTopics, tc.wantWeak) {
				t.Errorf("weak = %v, want %v", got.WeakTopics, tc.wantWeak)
			}
			if !reflect.DeepEqual(got.StrongTopics, tc.wantStrong) {
				t.Errorf("strong = %v, want %v", got.StrongTopics, tc.wantStrong)
			}
			if len(got.Wrong) != tc.wantWrong {
				t.Errorf("wrong = %d items, want %d", len(got.Wrong), tc.wantWrong)
			}
		})
	}
}

// TestBuildPracticeBriefCarriesTheMiss: what the model is shown for a
// wrong answer is the stem, the key and the pick — as content, not labels,
// since labels were shuffled per sitting and mean nothing on their own.
func TestBuildPracticeBriefCarriesTheMiss(t *testing.T) {
	got := command.BuildPracticeBrief([]*exam.UserExamDetail{
		answered("phép trừ", "5 - 2 = ?", "3", "4", false),
	})
	want := []string{"5 - 2 = ?", "phép trừ", "3", "4"}
	w := got.Wrong[0]
	if have := []string{w.Stem, w.Topic, w.RightAnswer, w.ChildAnswer}; !reflect.DeepEqual(have, want) {
		t.Fatalf("item = %v, want %v", have, want)
	}
}
