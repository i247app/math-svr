package repositories

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/shared/enum"
	"time"
)

// profileOwnedTables are the tables whose rows belong to a profile through
// their `profile_id` column. ma_profiles itself is not listed: its own
// SoftDelete / ForceDelete also set profile_status. A new table with a
// profile_id column goes here so /profiles/soft-delete and
// /profiles/force-delete cover it.
var profileOwnedTables = []string{
	classroomMemberTable,
	exerciseSubmissionTable,
	chatParticipantTable,
	examLinkTable,
	examSessionTable,
}

// ProfileOwnedDataRepository implements profile.IOwnedDataRepository.
type ProfileOwnedDataRepository struct {
	db database.Executor
}

func NewProfileOwnedDataRepository(db database.Executor) *ProfileOwnedDataRepository {
	return &ProfileOwnedDataRepository{db: db}
}

// SoftDeleteByProfileId keeps modify_dt as it is (deleted_dt records the
// delete): bulk UPDATEs on ma_exam_sessions must not move its "latest
// journey" order.
func (r *ProfileOwnedDataRepository) SoftDeleteByProfileId(ctx context.Context, profileId int64) error {
	now := time.Now().UTC()
	for _, table := range profileOwnedTables {
		query := `UPDATE ` + table + ` SET status = ?, deleted_dt = ?, modify_dt = modify_dt
			WHERE profile_id = ? AND deleted_dt IS NULL`
		if _, err := r.db.Exec(ctx, query, enum.StatusInactive, now, profileId); err != nil {
			return fmt.Errorf("profile owned data repo soft delete %s: %w", table, err)
		}
	}
	return nil
}

// ForceDeleteByProfileId also deletes the answer lines of the profile's
// attempts — ma_exam_session_lines has no profile_id, so it goes first,
// joined through ma_exam_links, while the attempts still exist.
func (r *ProfileOwnedDataRepository) ForceDeleteByProfileId(ctx context.Context, profileId int64) error {
	linesQuery := `DELETE l FROM ` + examSessionLineTable + ` l
		JOIN ` + examLinkTable + ` k ON k.elink_id = l.elink_id
		WHERE k.profile_id = ?`
	if _, err := r.db.Exec(ctx, linesQuery, profileId); err != nil {
		return fmt.Errorf("profile owned data repo force delete %s: %w", examSessionLineTable, err)
	}
	for _, table := range profileOwnedTables {
		query := `DELETE FROM ` + table + ` WHERE profile_id = ?`
		if _, err := r.db.Exec(ctx, query, profileId); err != nil {
			return fmt.Errorf("profile owned data repo force delete %s: %w", table, err)
		}
	}
	return nil
}
