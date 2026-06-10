package enum

import "regexp"

// ProgressBucket selects the chart's x-axis granularity for the
// classroom-learning-progress endpoints. The validator normalises the
// request to one of these three values before any SQL is built — DAY
// groups submissions per calendar day, WEEK per ISO week (Monday
// start), MONTH per calendar month, all in the request's tz.
type ProgressBucket string

const (
	ProgressBucketDay   ProgressBucket = "DAY"
	ProgressBucketWeek  ProgressBucket = "WEEK"
	ProgressBucketMonth ProgressBucket = "MONTH"
)

func (b ProgressBucket) String() string {
	return string(b)
}

func (b ProgressBucket) IsValid() bool {
	switch b {
	case ProgressBucketDay,
		ProgressBucketWeek,
		ProgressBucketMonth:
		return true
	default:
		return false
	}
}

// ProgressComment classifies a single student's recent trend on the
// per-student card in /classrooms/progress/students. The rule table
// and evaluation order live in
// docs/features/001-class-learning-progress/FEATURE-SPEC.md §4 and are
// implemented by application/query/classroomprogress.Classify.
//
// NO_DATA and INSUFFICIENT are sentinel values for "not enough
// submissions to render a chip" — the client suppresses the colored
// pill for those rows but the row itself stays in the list
// (decision #18: always include all students).
type ProgressComment string

const (
	ProgressCommentNoData       ProgressComment = "NO_DATA"
	ProgressCommentInsufficient ProgressComment = "INSUFFICIENT"
	ProgressCommentNeedToTry    ProgressComment = "NEED_TO_TRY"
	ProgressCommentProgress     ProgressComment = "PROGRESS"
	ProgressCommentGoodProgress ProgressComment = "GOOD_PROGRESS"
)

func (c ProgressComment) String() string {
	return string(c)
}

func (c ProgressComment) IsValid() bool {
	switch c {
	case ProgressCommentNoData,
		ProgressCommentInsufficient,
		ProgressCommentNeedToTry,
		ProgressCommentProgress,
		ProgressCommentGoodProgress:
		return true
	default:
		return false
	}
}

// TrendSource labels which underlying series fed the comment chip on
// /classrooms/progress/students. The endpoint exposes both
// trend_series_by_day and trend_series_by_exercise per the 2026-06-10
// refinement; this enum tells the client which one was used for the
// Comment derivation. Default is LAST_5_DAYS (matches signed-off
// decision #1 and the screenshot footer "Xu hướng hiển thị dựa trên
// 5 ngày gần nhất").
type TrendSource string

const (
	TrendSourceLast5Days      TrendSource = "LAST_5_DAYS"
	TrendSourceLast5Exercises TrendSource = "LAST_5_EXERCISES"
)

func (t TrendSource) String() string {
	return string(t)
}

// DefaultProgressTz is the assumed-Vietnam offset applied when the
// request omits "tz". Matches the project default-Vietnamese posture
// — see .claude/rules/product.md §7. The value is passed verbatim
// into MySQL CONVERT_TZ.
const DefaultProgressTz = "+07:00"

// tzOffsetPattern restricts the request tz to a numeric offset like
// "+07:00" or "-05:30". IANA names ("Asia/Ho_Chi_Minh") are rejected
// because production MySQL is not guaranteed to have the tz tables
// loaded — keeping the format numeric lets CONVERT_TZ work everywhere.
const tzOffsetPattern = `^[+-]\d{2}:\d{2}$`

var tzOffsetRegex = regexp.MustCompile(tzOffsetPattern)

// IsValidTzOffset reports whether s is a CONVERT_TZ-compatible offset
// string. Empty strings are NOT valid here — callers should substitute
// DefaultProgressTz before validating.
func IsValidTzOffset(s string) bool {
	return tzOffsetRegex.MatchString(s)
}
