// Package placement turns a child's exam statistics into the three
// server-derived fields on ma_user_exams: res_grade, res_level and
// res_review.
//
// It lives on its own, away from the submit command, because all three
// are placeholders the teaching team intends to replace. Keeping them
// behind one function means the real formula lands as a change to this
// file and nothing else — no touching the transaction, the repository or
// the response shape.
package placement

import (
	"fmt"

	"math-ai.com/math-ai/internal/shared/enum"
)

// PromotionThreshold is the score, in percent, that separates moving up
// from moving down. Scoring exactly at the threshold moves nobody: the
// measurement is not precise enough to justify acting on a tie.
const PromotionThreshold = 50

// Input is everything the rules read. The lifetime figures are the totals
// AFTER the current submission has been added, so a caller must apply its
// delta before deriving — otherwise the review it stores will describe the
// child as they were one exam ago.
type Input struct {
	// CurrentGrade is res_grade before this submission; nil the first time.
	CurrentGrade *int
	// ExamGrade is the grade the attempt was actually taken at. It seeds
	// placement on the first submission, when there is no measurement yet.
	ExamGrade int
	// LastScorePercentage is THIS attempt's score. The window is N = 1
	// today, which is the whole reason the lifetime figures below do not
	// take part in the grade decision.
	LastScorePercentage int

	LifetimeTotal   int
	LifetimeCorrect int
	LifetimeSkipped int
}

// Result carries the three derived fields. Grade and Level are pointers
// because "not measured yet" is a real state that must reach the column
// as NULL rather than as zero — zero is kindergarten, a very different
// claim.
type Result struct {
	Grade  *int
	Level  *int
	Review string
}

// Derive applies the current placeholder rules.
//
// Grade: a single-attempt window. Above the threshold moves up one band,
// below moves down one, a tie holds. Clamped to the bands the product
// serves. The lifetime counters are deliberately NOT consulted — with a
// lifetime average, a child who has answered five hundred questions can no
// longer move the number enough to be promoted, which is the opposite of
// what a placement rule is for.
//
// Level: banded off the lifetime accuracy, as an interim measure so the
// intensity axis does something at all. It is the crudest of the three
// rules and the first that should be replaced.
func Derive(in Input) Result {
	grade := deriveGrade(in)
	level := deriveLevel(in)
	return Result{
		Grade:  grade,
		Level:  level,
		Review: buildReview(in, grade, level),
	}
}

func deriveGrade(in Input) *int {
	base := in.ExamGrade
	if in.CurrentGrade != nil {
		base = *in.CurrentGrade
	}

	switch {
	case in.LastScorePercentage > PromotionThreshold:
		base++
	case in.LastScorePercentage < PromotionThreshold:
		base--
	}

	if base < enum.ExamGradeMin {
		base = enum.ExamGradeMin
	}
	if base > enum.ExamGradeMax {
		base = enum.ExamGradeMax
	}
	return &base
}

// Level bands. Interim: the teaching team has not defined how level
// should move, so this maps lifetime accuracy onto the 1..10 scale
// coarsely and honestly rather than pinning every exam at level 1.
const (
	levelStart      = 1
	levelStruggling = 2
	levelSteady     = 5
	levelStrong     = 8

	levelSteadyFloor = 50
	levelStrongFloor = 80
)

func deriveLevel(in Input) *int {
	level := levelStart
	if in.LifetimeTotal > 0 {
		accuracy := in.LifetimeCorrect * 100 / in.LifetimeTotal
		switch {
		case accuracy > levelStrongFloor:
			level = levelStrong
		case accuracy >= levelSteadyFloor:
			level = levelSteady
		default:
			level = levelStruggling
		}
	}
	if level < enum.ExamLevelMin {
		level = enum.ExamLevelMin
	}
	if level > enum.ExamLevelMax {
		level = enum.ExamLevelMax
	}
	return &level
}

// reviewMaxLen keeps the sentence inside a sane display width. The column
// is TEXT, so this is about the parent reading it, not about the schema.
const reviewMaxLen = 250

// buildReview writes the cumulative sentence in Vietnamese — the product's
// only language, and the one the parent reads. It describes the whole
// history rather than the last sitting, because that is what the column
// means: it is overwritten on every submit and there is one row per exam
// type, not one per attempt.
func buildReview(in Input, grade, level *int) string {
	if in.LifetimeTotal <= 0 {
		return "Chưa có dữ liệu để đánh giá."
	}

	accuracy := in.LifetimeCorrect * 100 / in.LifetimeTotal
	body := fmt.Sprintf("Đã trả lời %d câu, đúng %d câu (%d%%).",
		in.LifetimeTotal, in.LifetimeCorrect, accuracy)

	if in.LifetimeSkipped > 0 {
		body += fmt.Sprintf(" Còn bỏ trống %d câu.", in.LifetimeSkipped)
	}
	if grade != nil {
		body += fmt.Sprintf(" Đang ở mức %s", gradeLabel(*grade))
		if level != nil {
			body += fmt.Sprintf(" - cấp độ %d", *level)
		}
		body += "."
	}

	if runes := []rune(body); len(runes) > reviewMaxLen {
		body = string(runes[:reviewMaxLen-1]) + "…"
	}
	return body
}

func gradeLabel(grade int) string {
	if grade == enum.ExamGradeMin {
		return "Mẫu giáo"
	}
	return fmt.Sprintf("Lớp %d", grade)
}
