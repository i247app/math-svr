package enum

import "regexp"

// ProgressComment classifies a single student's recent trend on the
// profile-learning-progress card (Xu hướng). The rule table and
// evaluation order live in
// docs/features/002-profile-learning-progress/FEATURE-SPEC.md §4 and are
// implemented by application/query/progress.Classify.
//
// NO_DATA and INSUFFICIENT are sentinel values for "not enough
// submissions to render a chip" — the client suppresses the colored
// pill for those rows.
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

// DefaultProgressTz is the assumed-Vietnam offset applied when the
// request omits "tz". Matches the project default-Vietnamese posture
// — see .claude/rules/product.md §7. The value is passed verbatim into
// MySQL CONVERT_TZ.
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
