// Package placement turns a child's exam statistics into the two
// server-derived fields on ma_user_exams: res_grade and res_review.
//
// It lives on its own, away from the submit command, because both are
// placeholders the teaching team intends to replace. Keeping them behind
// one function means the real formula lands as a change to this file and
// nothing else — no touching the transaction, the repository or the
// response shape.
//
// res_level is NOT derived here, or anywhere. The column exists and stays
// NULL: there is no agreed rule for what a level is, and an interim
// formula would have been a rule nobody signed off on, written into every
// child's row. When the teaching team defines one, it lands as a third
// field on Result and nothing else changes.
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

// Result carries the derived fields. Grade is a pointer because "not
// measured yet" is a real state that must reach the column as NULL rather
// than as zero — zero is kindergarten, a very different claim.
type Result struct {
	Grade  *int
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
func Derive(in Input) Result {
	grade := deriveGrade(in)
	return Result{
		Grade:  grade,
		Review: buildReview(in, grade),
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

// reviewMaxLen keeps the sentence inside a sane display width. The column
// is TEXT, so this is about the parent reading it, not about the schema.
const reviewMaxLen = 250

// buildReview writes the cumulative sentence in Vietnamese — the product's
// only language, and the one the parent reads. It describes the whole
// history rather than the last sitting, because that is what the column
// means: it is overwritten on every submit and there is one row per exam
// type, not one per attempt.
func buildReview(in Input, grade *int) string {
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
		body += fmt.Sprintf(" Đang ở mức %s.", gradeLabel(*grade))
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
