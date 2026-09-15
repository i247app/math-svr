// Package placement turns a child's exam statistics into the one
// server-derived field on ma_user_exams: res_review.
//
// It used to derive res_grade as well — a single-attempt window that
// moved the child up or down a band on each submit. That rule is gone:
// the columns are now current_grade / current_level, stated by the
// client on each hand-out, and the server records them without judging
// them. What remains here is the review sentence, kept behind one
// function so the wording can change without touching the transaction,
// the repository or the response shape.
package placement

import "fmt"

// Input is everything the review reads. The lifetime figures are the
// totals AFTER the current submission has been added, so a caller must
// apply its delta before deriving — otherwise the review it stores will
// describe the child as they were one exam ago.
type Input struct {
	LifetimeTotal   int
	LifetimeCorrect int
	LifetimeSkipped int
}

// Result carries the derived field.
type Result struct {
	Review string
}

// Derive builds the review for a journey's own row.
func Derive(in Input) Result {
	return Result{Review: buildReview(in)}
}

// DerivePractice builds the review for a PRACTICE row. It is the same
// sentence over the practice totals; it stays a separate entry point so a
// practice-specific wording can land here without touching the command.
func DerivePractice(in Input) Result {
	return Result{Review: buildReview(in)}
}

// reviewMaxLen keeps the sentence inside a sane display width. The column
// is TEXT, so this is about the parent reading it, not about the schema.
const reviewMaxLen = 250

// buildReview writes the cumulative sentence in Vietnamese — the product's
// only language, and the one the parent reads. It describes the whole
// history rather than the last sitting, because that is what the column
// means: it is overwritten on every submit and there is one row per exam
// type, not one per attempt.
func buildReview(in Input) string {
	if in.LifetimeTotal <= 0 {
		return "Chưa có dữ liệu để đánh giá."
	}

	accuracy := in.LifetimeCorrect * 100 / in.LifetimeTotal
	body := fmt.Sprintf("Đã trả lời %d câu, đúng %d câu (%d%%).",
		in.LifetimeTotal, in.LifetimeCorrect, accuracy)

	if in.LifetimeSkipped > 0 {
		body += fmt.Sprintf(" Còn bỏ trống %d câu.", in.LifetimeSkipped)
	}

	if runes := []rune(body); len(runes) > reviewMaxLen {
		body = string(runes[:reviewMaxLen-1]) + "…"
	}
	return body
}
