package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/domain/seq"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/shared/enum"
)

// clearTarget pairs a wipeable table with the ma_seqs counter that mints its
// external ids, plus that external-id column. It is the single source of truth
// for what the clear-data endpoints may touch — both the full wipe and the
// table-scoped variant — and, in the keep-users path, lets each surviving
// table's sequence be reset to MAX(external id) instead of 0.
type clearTarget struct {
	table    string
	seq      string
	extIDCol string // the seq-minted external id column (e.g. "uid", "profile_id")
}

// clearDataTargets lists the user-generated tables wiped by ClearData, each
// paired with the sequence reset alongside it. Reference / seed tables
// (programs, grades, semesters, schools) are intentionally excluded so the
// curriculum data seeded outside the app survives. Mirrors sql/clear_data.sql.
var clearDataTargets = []clearTarget{
	{userTable, seq.NameUser, "uid"},
	{aliasTable, seq.NameAlias, "alias_id"},
	{deviceTable, seq.NameDevice, "device_id"},
	{loginLogTable, seq.NameLoginLog, "login_log_id"},
	{profileTable, seq.NameProfile, "profile_id"},
	{otpTable, seq.NameOtp, "otp_id"},
	{classroomTable, seq.NameClassroom, "classroom_id"},
	{classroomMemberTable, seq.NameClassroomMember, "member_id"},
	// classroomInvitationTable is a legacy orphan — no live writes.
	{classroomProgramTable, seq.NameClassroomProgram, "classroom_program_id"},
	{exerciseTable, seq.NameClassroomExercise, "classroom_exercise_id"},
	{exerciseSubmissionTable, seq.NameClassroomExerciseSubmission, "classroom_exercise_submission_id"},
	{notificationTable, seq.NameNotification, "notification_id"},
	{bannerTable, seq.NameBanner, "banner_id"},
	{chatConversationTable, seq.NameChatConversation, "conversation_id"},
	{chatParticipantTable, seq.NameChatParticipant, "participant_id"},
	{chatMessageTable, seq.NameChatMessage, "message_id"},
	{aiExamTable, seq.NameAiExam, "ai_exam_id"},
	{userAiExamTable, seq.NameUserAiExam, "user_ai_exam_id"},
	{userExamTable, seq.NameUserExam, "user_exam_id"},
	{userExamDetailTable, seq.NameUserExamDetail, "user_exam_detail_id"},
}

// clearDataTargetByTable indexes clearDataTargets for O(1) allow-list checks in
// the table-scoped clear path.
var clearDataTargetByTable = func() map[string]clearTarget {
	m := make(map[string]clearTarget, len(clearDataTargets))
	for _, t := range clearDataTargets {
		m[t.table] = t
	}
	return m
}()

// clearDataTables is the ordered list of tables wiped by the full ClearData,
// derived from clearDataTargets so the two can never drift.
var clearDataTables = func() []string {
	tables := make([]string, len(clearDataTargets))
	for i, t := range clearDataTargets {
		tables[i] = t.table
	}
	return tables
}()

// clearDataSeqs lists the external-id counters reset back to 0 for the wiped
// aggregates only. Reference sequences (program/grade/semester/school)
// are intentionally left untouched so seeded rows keep their
// ids. Mirrors sql/clear_data.sql.
var clearDataSeqs = []string{
	seq.NameUser,
	seq.NameAlias,
	seq.NameDevice,
	seq.NameLoginLog,
	seq.NameProfile,
	seq.NameOtp,
	seq.NameClassroom,
	seq.NameClassroomMember,
	seq.NameClassroomInviation,
	seq.NameClassroomProgram,
	seq.NameClassroomExercise,
	seq.NameClassroomExerciseSubmission,
	seq.NameNotification,
	seq.NameBanner,
	seq.NameChatConversation,
	seq.NameChatParticipant,
	seq.NameChatMessage,
	seq.NameAiExam,
	seq.NameUserAiExam,
	seq.NameUserExam,
	seq.NameUserExamDetail,
}

// keepCol names the column a table-scoped keep filters on when preserving
// whitelisted users in ClearDataKeepingUsers.
type keepCol struct {
	table string
	col   string
}

// The three classifications below decide, for the whitelist path, HOW each
// wiped table relates to a user. Together with userExamDetailTable (kept by
// parent user_exam_id) and aiExamTable (kept by referenced ai_exam_id) they
// must cover every entry in clearDataTargets — enforced by init() so adding a
// table to clearDataTargets forces a matching classification here.

// clearKeepByUid — a row belongs to the user in its uid / sender_uid column;
// rows whose uid is whitelisted survive.
var clearKeepByUid = []keepCol{
	{userTable, "uid"},
	{aliasTable, "uid"},
	{deviceTable, "uid"},
	{loginLogTable, "uid"},
	{otpTable, "uid"},
	{notificationTable, "uid"},
	{profileTable, "uid"},
	{userAiExamTable, "uid"},
	{userExamTable, "uid"},
	{chatParticipantTable, "uid"},
	{chatMessageTable, "sender_uid"},
}

// clearKeepByProfile — a row is owned via profile_id; kept when that profile
// belongs to a whitelisted user.
var clearKeepByProfile = []keepCol{
	{classroomMemberTable, "profile_id"},
	{exerciseSubmissionTable, "profile_id"},
}

// clearFullWipe — no per-user ownership column; always emptied even in the
// whitelist path (shared/collaborative parents and the ai-exam cache is
// handled separately by reference).
var clearFullWipe = []string{
	classroomTable,
	classroomProgramTable,
	exerciseTable,
	chatConversationTable,
	bannerTable,
}

func init() {
	covered := map[string]bool{
		// Handled specially in ClearDataKeepingUsers (parent-id linkage).
		userExamDetailTable: true,
		aiExamTable:         true,
	}
	for _, t := range clearKeepByUid {
		covered[t.table] = true
	}
	for _, t := range clearKeepByProfile {
		covered[t.table] = true
	}
	for _, t := range clearFullWipe {
		covered[t] = true
	}
	for _, t := range clearDataTargets {
		if !covered[t.table] {
			panic("maintenance: clear-data keep-users classification missing table " + t.table)
		}
	}
}

// MaintenanceRepository owns destructive, cross-aggregate maintenance SQL that
// does not belong to any single aggregate repository (e.g. wiping all
// user-generated data). Keeping the raw SQL here honours the rule that no SQL
// leaks into the module / handler layer.
type MaintenanceRepository struct {
	db database.Executor
}

func NewMaintenanceRepository(db database.Executor) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

// ListStaleGuestUserIds returns guests whose last sign of life is older
// than `before`, newest-idle first, capped at `limit`.
//
// "Last sign of life" is the newest exam they were handed, falling back
// to when the account was opened — a guest who never got past the first
// generate has no sitting to date them by. Both are needed: the account's
// own timestamps never move, because nothing updates a guest row after
// it is created, so create_dt alone would retire an actively-used guest.
//
// The read spans ma_users and ma_user_ai_exams, which is why it lives
// here rather than in either aggregate's repository: this is a
// maintenance sweep, the same shape as ClearData.
func (r *MaintenanceRepository) ListStaleGuestUserIds(ctx context.Context, before time.Time, limit int) ([]int64, error) {
	if limit <= 0 {
		return nil, nil
	}
	query := `
		SELECT u.uid
		FROM ` + userTable + ` u
		LEFT JOIN ` + profileTable + ` p
			ON p.uid = u.uid AND p.deleted_dt IS NULL
		LEFT JOIN ` + userAiExamTable + ` a
			ON a.profile_id = p.profile_id AND a.deleted_dt IS NULL
		WHERE u.identity_code = ?
		  AND u.status IN (?) AND u.deleted_dt IS NULL
		GROUP BY u.uid, u.create_dt
		HAVING COALESCE(MAX(a.started_dt), u.create_dt) < ?
		ORDER BY COALESCE(MAX(a.started_dt), u.create_dt) ASC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, enum.IdentityCodeGuest, enum.StatusActive, before, limit)
	if err != nil {
		return nil, fmt.Errorf("maintenance repo list stale guests: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("maintenance repo scan stale guest: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("maintenance repo stale guests iteration: %w", err)
	}
	return ids, nil
}

// ClearData TRUNCATEs every user-generated table and resets the matching
// external-id counters in ma_seqs. It returns the tables cleared and the
// sequences reset so callers can report exactly what was wiped.
//
// The schema declares no foreign keys (integrity is enforced in the
// application layer inside UnitOfWork blocks), so each TRUNCATE is independent
// and order does not matter. TRUNCATE additionally resets each table's internal
// AUTO_INCREMENT id.
func (r *MaintenanceRepository) ClearData(ctx context.Context) ([]string, []string, error) {
	for _, table := range clearDataTables {
		// Table names come from the fixed internal allow-list above, never
		// from user input, so string concatenation is safe here.
		if _, err := r.db.Exec(ctx, "TRUNCATE TABLE "+table); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data: truncate %s: %w", table, err)
		}
	}

	placeholders := make([]string, len(clearDataSeqs))
	args := make([]any, len(clearDataSeqs))
	for i, name := range clearDataSeqs {
		placeholders[i] = "?"
		args[i] = name
	}
	resetQuery := `UPDATE ` + seqTable + ` SET current_value = 0 WHERE seq_name IN (` +
		strings.Join(placeholders, ", ") + `)`
	if _, err := r.db.Exec(ctx, resetQuery, args...); err != nil {
		return nil, nil, fmt.Errorf("maintenance repo clear-data: reset seqs: %w", err)
	}

	return clearDataTables, clearDataSeqs, nil
}

// ClearableTables returns the allow-list of tables the table-scoped clear may
// wipe, in canonical order. Callers validate a request against it before
// invoking ClearDataTables.
func (r *MaintenanceRepository) ClearableTables() []string {
	return append([]string(nil), clearDataTables...)
}

// ClearDataTables TRUNCATEs only the requested tables and resets each one's
// external-id counter. Every name must be in the clear-data allow-list
// (clearDataTargetByTable); an unknown name is rejected before any write, so a
// bad request wipes nothing. Duplicate names are collapsed.
func (r *MaintenanceRepository) ClearDataTables(ctx context.Context, tables []string) ([]string, []string, error) {
	targets := make([]clearTarget, 0, len(tables))
	seen := make(map[string]bool, len(tables))
	for _, name := range tables {
		target, ok := clearDataTargetByTable[name]
		if !ok {
			return nil, nil, fmt.Errorf("maintenance repo clear-data: table %q is not clearable", name)
		}
		if seen[target.table] {
			continue
		}
		seen[target.table] = true
		targets = append(targets, target)
	}

	cleared := make([]string, 0, len(targets))
	seqsReset := make([]string, 0, len(targets))
	for _, target := range targets {
		// Table name is an allow-listed constant, never raw user input.
		if _, err := r.db.Exec(ctx, "TRUNCATE TABLE "+target.table); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data: truncate %s: %w", target.table, err)
		}
		cleared = append(cleared, target.table)
		if target.seq != "" {
			seqsReset = append(seqsReset, target.seq)
		}
	}

	if len(seqsReset) > 0 {
		placeholders := make([]string, len(seqsReset))
		args := make([]any, len(seqsReset))
		for i, name := range seqsReset {
			placeholders[i] = "?"
			args[i] = name
		}
		resetQuery := `UPDATE ` + seqTable + ` SET current_value = 0 WHERE seq_name IN (` +
			strings.Join(placeholders, ", ") + `)`
		if _, err := r.db.Exec(ctx, resetQuery, args...); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data: reset seqs: %w", err)
		}
	}

	return cleared, seqsReset, nil
}

// ClearDataKeepingUsers wipes every user-generated table EXCEPT the rows owned
// by the whitelisted users (identified by ma_users.uid). A whitelisted user
// keeps their account (users/aliases/devices/otps/login_logs), profiles, and
// everything hanging off them (exams, journeys, answer details, submissions,
// classroom memberships, chat participation and messages, notifications).
//
// Deletes use DELETE, not TRUNCATE. Afterwards each wiped table's ma_seqs
// counter is reset to MAX(external id) of the rows that survived (0 when the
// table was emptied), so the counters shrink without ever re-minting an id that
// a kept row still holds — the next Seq.Next yields MAX+increment.
//
// Caveat: the shared parents in clearFullWipe (classrooms, chat conversations,
// exercises, banners) are emptied whole, so a kept user's classroom-membership
// / chat-participation rows may reference a now-deleted parent. There are no
// FKs, so this is inert data rather than an error.
//
// Ownership id-sets are read up front (before any delete), so the
// complement-deletes below are order-independent.
func (r *MaintenanceRepository) ClearDataKeepingUsers(ctx context.Context, keepUids []int64) ([]string, []string, error) {
	if len(keepUids) == 0 {
		return nil, nil, fmt.Errorf("maintenance repo clear-data keep: keepUids is required")
	}

	uidPh, uidArgs := int64InClause(keepUids)

	keepProfileIds, err := r.selectInt64s(ctx,
		`SELECT profile_id FROM `+profileTable+` WHERE uid IN (`+uidPh+`)`, uidArgs...)
	if err != nil {
		return nil, nil, err
	}
	keepUserExamIds, err := r.selectInt64s(ctx,
		`SELECT user_exam_id FROM `+userExamTable+` WHERE uid IN (`+uidPh+`)`, uidArgs...)
	if err != nil {
		return nil, nil, err
	}
	keepAiExamIds, err := r.selectInt64s(ctx,
		`SELECT DISTINCT ai_exam_id FROM `+userAiExamTable+` WHERE uid IN (`+uidPh+`)`, uidArgs...)
	if err != nil {
		return nil, nil, err
	}

	processed := make([]string, 0, len(clearDataTargets))

	// Keep-by-uid: delete every row whose uid is not whitelisted.
	for _, t := range clearKeepByUid {
		// Table + column are internal constants, never user input; the ids are
		// bound as parameters.
		q := `DELETE FROM ` + t.table + ` WHERE ` + t.col + ` NOT IN (` + uidPh + `)`
		if _, err := r.db.Exec(ctx, q, uidArgs...); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data keep: delete %s: %w", t.table, err)
		}
		processed = append(processed, t.table)
	}

	// Keep-by-profile: keep rows whose profile belongs to a whitelisted user.
	for _, t := range clearKeepByProfile {
		if err := r.deleteExcept(ctx, t.table, t.col, keepProfileIds); err != nil {
			return nil, nil, err
		}
		processed = append(processed, t.table)
	}

	// Answer details hang off the kept journey rows.
	if err := r.deleteExcept(ctx, userExamDetailTable, "user_exam_id", keepUserExamIds); err != nil {
		return nil, nil, err
	}
	processed = append(processed, userExamDetailTable)

	// ai-exam cache: keep only the exams a kept attempt actually references, so
	// the kept users' exam content survives; drop the rest of the cache.
	if err := r.deleteExcept(ctx, aiExamTable, "ai_exam_id", keepAiExamIds); err != nil {
		return nil, nil, err
	}
	processed = append(processed, aiExamTable)

	// Shared parents with no per-user owner: emptied whole.
	for _, table := range clearFullWipe {
		if _, err := r.db.Exec(ctx, "DELETE FROM "+table); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data keep: wipe %s: %w", table, err)
		}
		processed = append(processed, table)
	}

	// Re-align each wiped table's ma_seqs counter to the highest external id
	// that survived (0 when the table is now empty). Runs after every delete so
	// MAX reflects the final surviving rows; the cross-table subquery reads the
	// data table while updating ma_seqs, which MySQL allows.
	seqsReset := make([]string, 0, len(clearDataTargets))
	for _, t := range clearDataTargets {
		// table / column / seq name are internal constants, never user input.
		q := `UPDATE ` + seqTable + ` SET current_value = ` +
			`(SELECT COALESCE(MAX(` + t.extIDCol + `), 0) FROM ` + t.table + `) ` +
			`WHERE seq_name = ?`
		if _, err := r.db.Exec(ctx, q, t.seq); err != nil {
			return nil, nil, fmt.Errorf("maintenance repo clear-data keep: reset seq %s: %w", t.seq, err)
		}
		seqsReset = append(seqsReset, t.seq)
	}

	return processed, seqsReset, nil
}

// deleteExcept deletes every row of table whose col is not in keepIds. An empty
// keepIds means nothing is owned, so the table is emptied whole.
func (r *MaintenanceRepository) deleteExcept(ctx context.Context, table, col string, keepIds []int64) error {
	if len(keepIds) == 0 {
		if _, err := r.db.Exec(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("maintenance repo clear-data keep: wipe %s: %w", table, err)
		}
		return nil
	}
	ph, args := int64InClause(keepIds)
	q := `DELETE FROM ` + table + ` WHERE ` + col + ` NOT IN (` + ph + `)`
	if _, err := r.db.Exec(ctx, q, args...); err != nil {
		return fmt.Errorf("maintenance repo clear-data keep: delete %s: %w", table, err)
	}
	return nil
}

// selectInt64s runs a single-column int64 query and collects the values.
func (r *MaintenanceRepository) selectInt64s(ctx context.Context, query string, args ...any) ([]int64, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("maintenance repo clear-data keep: select ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("maintenance repo clear-data keep: scan id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("maintenance repo clear-data keep: iterate ids: %w", err)
	}
	return ids, nil
}

// int64InClause renders "?, ?, …" and the matching []any for an IN clause.
func int64InClause(ids []int64) (string, []any) {
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args[i] = id
	}
	return strings.Join(ph, ", "), args
}
