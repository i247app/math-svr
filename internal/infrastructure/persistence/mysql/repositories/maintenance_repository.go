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
// external ids. It is the single source of truth for what the clear-data
// endpoints may touch — both the full wipe and the table-scoped variant.
type clearTarget struct {
	table string
	seq   string
}

// clearDataTargets lists the user-generated tables wiped by ClearData, each
// paired with the sequence reset alongside it. Reference / seed tables
// (programs, grades, semesters, schools) are intentionally excluded so the
// curriculum data seeded outside the app survives. Mirrors sql/clear_data.sql.
var clearDataTargets = []clearTarget{
	{userTable, seq.NameUser},
	{aliasTable, seq.NameAlias},
	{deviceTable, seq.NameDevice},
	{loginLogTable, seq.NameLoginLog},
	{profileTable, seq.NameProfile},
	{otpTable, seq.NameOtp},
	{classroomTable, seq.NameClassroom},
	{classroomMemberTable, seq.NameClassroomMember},
	// classroomInvitationTable is a legacy orphan — no live writes.
	{classroomProgramTable, seq.NameClassroomProgram},
	{exerciseTable, seq.NameClassroomExercise},
	{exerciseSubmissionTable, seq.NameClassroomExerciseSubmission},
	{notificationTable, seq.NameNotification},
	{bannerTable, seq.NameBanner},
	{chatConversationTable, seq.NameChatConversation},
	{chatParticipantTable, seq.NameChatParticipant},
	{chatMessageTable, seq.NameChatMessage},
	{aiExamTable, seq.NameAiExam},
	{userAiExamTable, seq.NameUserAiExam},
	{userExamTable, seq.NameUserExam},
	{userExamDetailTable, seq.NameUserExamDetail},
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
