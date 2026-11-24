package enum

// GsessStatusType is the lifecycle of one game round. A round is
// IN_PROGRESS from its start and only ever leaves that state by being
// submitted, which makes it COMPLETE; an abandoned one stays IN_PROGRESS.
type GsessStatusType string

const (
	GsessStatusInProgress GsessStatusType = "IN_PROGRESS"
	GsessStatusComplete   GsessStatusType = "COMPLETE"
	GsessStatusDeleted    GsessStatusType = "DELETED"
)

func (s GsessStatusType) String() string { return string(s) }

func (s GsessStatusType) IsValid() bool {
	switch s {
	case GsessStatusInProgress, GsessStatusComplete, GsessStatusDeleted:
		return true
	default:
		return false
	}
}

// Questions per game round: what a client that states no number gets, and
// the range it may ask for.
const (
	GameNumQuestionsDefault = 10
	GameNumQuestionsMin     = 1
	GameNumQuestionsMax     = 50
)
