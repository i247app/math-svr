package repositories

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/shared/enum"
)

// userOwnedTables are the tables whose rows belong to a user through their
// `uid` column. A new table with a
// uid column goes here so /users/soft-delete and /users/force-delete cover it.
// ma_contact_us also has uid but no Go code at all (orphan schema).
var userOwnedTables = []string{
	userTable,
	aliasTable,
	profileTable,
	loginTable,
	deviceTable,
	loginLogTable,
	otpTable,
	notificationTable,
	presenceTable,
	chatParticipantTable,
	examLinkTable,
	examSessionTable,
}

// UserOwnedDataRepository implements user.IOwnedDataRepository.
type UserOwnedDataRepository struct {
	db database.Executor
}

func NewUserOwnedDataRepository(db database.Executor) *UserOwnedDataRepository {
	return &UserOwnedDataRepository{db: db}
}

// SoftDeleteByUid keeps modify_dt as it is (deleted_dt records the delete):
// bulk UPDATEs on ma_exam_sessions must not move its "latest journey" order.
func (r *UserOwnedDataRepository) SoftDeleteByUid(ctx context.Context, uid int64) error {
	now := mtime.Now().Time
	for _, table := range userOwnedTables {
		query := `UPDATE ` + table + ` SET status = ?, deleted_dt = ?, modify_dt = modify_dt
			WHERE uid = ? AND deleted_dt IS NULL`
		if _, err := r.db.Exec(ctx, query, enum.StatusInactive, now, uid); err != nil {
			return fmt.Errorf("user owned data repo soft delete %s: %w", table, err)
		}
	}
	return nil
}

// ForceDeleteByUid also deletes the answer lines of the user's attempts —
// ma_exam_session_lines has no uid, so it goes first, joined through
// ma_exam_links, while the attempts still exist.
func (r *UserOwnedDataRepository) ForceDeleteByUid(ctx context.Context, uid int64) error {
	linesQuery := `DELETE l FROM ` + examSessionLineTable + ` l
		JOIN ` + examLinkTable + ` k ON k.elink_id = l.elink_id
		WHERE k.uid = ?`
	if _, err := r.db.Exec(ctx, linesQuery, uid); err != nil {
		return fmt.Errorf("user owned data repo force delete %s: %w", examSessionLineTable, err)
	}
	for _, table := range userOwnedTables {
		query := `DELETE FROM ` + table + ` WHERE uid = ?`
		if _, err := r.db.Exec(ctx, query, uid); err != nil {
			return fmt.Errorf("user owned data repo force delete %s: %w", table, err)
		}
	}
	return nil
}
